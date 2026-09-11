// =============================================================================
// storage/avatar.go
//
// Сервіс збереження завантажених зображень аватарів на локальному диску.
//
// Дизайн:
//   - Валідація типу файлу відбувається за сигнатурою байтів (http.DetectContentType),
//     а не за розширенням імені файлу чи Content-Type заголовком клієнта (обидва
//     легко підробити).
//   - Унікальне ім'я файлу генерується через uuid.New(), щоб уникнути колізій
//     та проблем із кешуванням браузера при повторному завантаженні аватара.
//   - Файл записується атомарно: спершу у тимчасовий файл у тій же директорії,
//     потім rename (os.Rename на тому ж томі — атомарна операція в POSIX/NTFS).
// =============================================================================

package storage

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// MaxAvatarUploadBytes — абсолютна верхня межа (незалежно від конфігу), яку
// ми готові прочитати в пам'ять перед валідацією. Захищає від переповнення
// пам'яті зловмисним запитом із підробленим Content-Length.
const MaxAvatarUploadBytes = 20 * 1024 * 1024 // 20 MB (hard ceiling)

// allowedAvatarMIMETypes — білий список дозволених типів зображень.
// Ключ - MIME-тип, який повертає http.DetectContentType; значення - розширення файлу.
var allowedAvatarMIMETypes = map[string]string{
	"image/png":     ".png",
	"image/jpeg":    ".jpg",
	"image/webp":    ".webp",
	"image/svg+xml": ".svg",
}

// ErrAvatarTooLarge - файл перевищує дозволений розмір.
type ErrAvatarTooLarge struct {
	MaxBytes int64
}

func (e *ErrAvatarTooLarge) Error() string {
	return fmt.Sprintf("avatar file exceeds maximum allowed size of %d bytes", e.MaxBytes)
}

// ErrAvatarInvalidType - файл не відповідає жодному з дозволених MIME-типів.
type ErrAvatarInvalidType struct {
	Detected string
}

func (e *ErrAvatarInvalidType) Error() string {
	return fmt.Sprintf("unsupported avatar file type: %s (allowed: png, jpeg, webp, svg)", e.Detected)
}

// MaxAvatarSizeMB повертає налаштований ліміт розміру аватара в мегабайтах
// (з конфігу, дефолт 5 MB, якщо не задано або задано некоректне значення).
func (s *Storage) MaxAvatarSizeMB() int {
	if s.cfg.MaxAvatarSizeMB <= 0 {
		return 5
	}
	return s.cfg.MaxAvatarSizeMB
}

// MaxAvatarSizeBytes - те саме, що MaxAvatarSizeMB, але в байтах.
func (s *Storage) MaxAvatarSizeBytes() int64 {
	return int64(s.MaxAvatarSizeMB()) * 1024 * 1024
}

// AvatarUploadDir повертає абсолютний шлях до директорії, куди зберігаються
// завантажені аватари, гарантуючи її існування (створює за потреби).
func (s *Storage) AvatarUploadDir() (string, error) {
	dir := s.cfg.AvatarUploadDir
	if dir == "" {
		dir = "./uploads/avatars"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create avatar upload directory: %w", err)
	}
	return dir, nil
}

// SaveAvatarImage валідує та зберігає завантажене зображення аватара на диск.
//
// Повертає публічний URL-шлях (наприклад "/api/avatars/<uuid>.png"), який
// потрібно записати в users.avatar_url.
//
// Валідація:
//  1. Розмір даних не перевищує maxBytes (переданий викликачем, з конфігу).
//  2. Реальна сигнатура байтів (http.DetectContentType) відповідає одному
//     з дозволених типів (PNG/JPEG/WebP/SVG) - не довіряємо клієнтському
//     Content-Type чи розширенню імені файлу.
func (s *Storage) SaveAvatarImage(userID uuid.UUID, data []byte, maxBytes int64) (string, error) {
	if int64(len(data)) > maxBytes {
		return "", &ErrAvatarTooLarge{MaxBytes: maxBytes}
	}

	detected := http.DetectContentType(data)
	ext, ok := allowedAvatarMIMETypes[detected]
	if !ok {
		// http.DetectContentType не розпізнає SVG (текстовий XML) як image/svg+xml
		// надійно у всіх випадках - додатково перевіряємо сигнатуру вручну.
		if looksLikeSVG(data) {
			ext = ".svg"
			ok = true
		}
	}
	if !ok {
		return "", &ErrAvatarInvalidType{Detected: detected}
	}

	dir, err := s.AvatarUploadDir()
	if err != nil {
		return "", err
	}

	filename := uuid.New().String() + ext
	finalPath := filepath.Join(dir, filename)
	tmpPath := finalPath + ".tmp"

	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return "", fmt.Errorf("failed to write temp avatar file: %w", err)
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("failed to finalize avatar file: %w", err)
	}

	baseURL := s.cfg.AvatarBaseURL
	if baseURL == "" {
		baseURL = "/api/avatars"
	}
	return baseURL + "/" + filename, nil
}

// DeleteAvatarFile видаляє попередній файл аватара з диску за його публічним
// URL-шляхом (найкраще старання - помилки лише логуються викликачем).
// Ігнорує значення, що не належать директорії завантажень (напр. DiceBear-посилання).
func (s *Storage) DeleteAvatarFile(avatarURL string) error {
	if avatarURL == "" {
		return nil
	}
	baseURL := s.cfg.AvatarBaseURL
	if baseURL == "" {
		baseURL = "/api/avatars"
	}
	prefix := baseURL + "/"
	if len(avatarURL) <= len(prefix) || avatarURL[:len(prefix)] != prefix {
		return nil
	}
	filename := filepath.Base(avatarURL[len(prefix):])
	if filename == "" || filename == "." || filename == ".." {
		return nil
	}

	dir, err := s.AvatarUploadDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, filename)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete old avatar file: %w", err)
	}
	return nil
}

// looksLikeSVG - грубе розпізнавання SVG за наявністю "<svg" у перших байтах.
// Достатньо для базової валідації; повний XML-парсинг тут надлишковий.
// http.DetectContentType не завжди повертає "image/svg+xml" для валідного SVG
// (напр. якщо файл починається з XML-декларації чи коментаря), тому цей
// фолбек ловить такі випадки.
func looksLikeSVG(data []byte) bool {
	limit := 512
	if len(data) < limit {
		limit = len(data)
	}
	return bytes.Contains(bytes.ToLower(data[:limit]), []byte("<svg"))
}
