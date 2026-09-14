package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	APIPort    int
	SocketPort int
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBPoolMax  int
	DBPoolMin  int
	DBLogging  bool

	AdminUser  string
	AdminEmail string
	AdminPass  string

	AvatarUploadDir string
	AvatarBaseURL   string
	MaxAvatarSizeMB int

	// TLSEnabled — увімкнути автоматичне отримання/оновлення сертифікатів
	// через Let's Encrypt (autocert). У false (за замовчуванням) сервер
	// продовжує працювати як звичайний HTTP-сервер на SocketPort — сумісно
	// з існуючою топологією, де TLS термінується на nginx.
	TLSEnabled bool
	// TLSDomains — білий список доменів, для яких autocert.Manager
	// дозволить видачу сертифікатів (HostPolicy). ОБОВ'ЯЗКОВО має бути
	// непорожнім, якщо TLSEnabled=true — інакше Let's Encrypt видасть
	// сертифікат будь-якому домену, що вкаже DNS на цей сервер.
	TLSDomains []string
	// TLSCacheDir — локальна директорія для персистентного кешу сертифікатів
	// (autocert.DirCache), щоб не перевипускати їх при кожному перезапуску.
	TLSCacheDir string
	// TLSEmail — контактний email, який Let's Encrypt використовує для
	// сповіщень про протермінування/відкликання сертифікатів.
	TLSEmail string

	// CORSAllowedOrigins — білий список Origin-ів, яким CORSMiddleware
	// дозволяє cross-origin запити до /api/* та /ws (Access-Control-Allow-Origin
	// повертається динамічно, лише якщо Origin запиту є в цьому списку).
	// За замовчуванням містить лише Vite dev-сервер — production-домен(и)
	// ОБОВ'ЯЗКОВО додавати через CORS_ALLOWED_ORIGINS (comma-separated),
	// інакше браузер блокуватиме запити з продакшен-фронтенду.
	CORSAllowedOrigins []string

	// FrontendProxyURL — базовий URL внутрішнього SPA-контейнера (nginx,
	// що лише роздає статичні файли index.html/assets). Коли заданий,
	// Go-бекенд стає ЄДИНОЮ публічною точкою входу (порти 80/443):
	// маршрути /api/* та /ws обробляються самим бекендом, а всі інші
	// запити (SPA-роутинг, статика) проксіюються на цей URL. Порожнє
	// значення вимикає проксі (напр. для локальної розробки через Vite).
	FrontendProxyURL string
}

func Load() *Config {
	_ = godotenv.Load(".env")

	conf := &Config{
		APIPort:    getEnvAsInt("PORT", 8080),
		SocketPort: getEnvAsInt("SOCKET_PORT", 3000),

		DBHost:     getEnv("POSTGRES_HOST", "127.0.0.1"),
		DBPort:     getEnvAsInt("POSTGRES_PORT", 5432),
		DBUser:     getEnv("POSTGRES_USER", "postgres"),
		DBPassword: getEnv("POSTGRES_PASSWORD", "111"),
		DBName:     getEnv("POSTGRES_DB", "love_letter"),
		DBPoolMax:  getEnvAsInt("POSTGRES_POOL_MAX", 20),
		DBPoolMin:  getEnvAsInt("POSTGRES_POOL_MIN", 5),
		DBLogging:  getEnvAsBool("POSTGRES_LOGGING", false),

		AdminUser:  getEnv("ADMIN_USER", "admin"),
		AdminEmail: getEnv("ADMIN_EMAIL", "admin@loveletter.local"),
		AdminPass:  getEnv("ADMIN_PASSWORD", "SuperSecureAdminPassword2026"),

		AvatarUploadDir: getEnv("AVATAR_UPLOAD_DIR", "./uploads/avatars"),
		AvatarBaseURL:   getEnv("AVATAR_BASE_URL", "/api/avatars"),
		MaxAvatarSizeMB: getEnvAsInt("MAX_AVATAR_SIZE_MB", 5),

		TLSEnabled:  getEnvAsBool("TLS_ENABLED", false),
		TLSDomains:  getEnvAsStringSlice("TLS_DOMAINS", ",", nil),
		TLSCacheDir: getEnv("TLS_CACHE_DIR", "./certs-cache"),
		TLSEmail:    getEnv("TLS_EMAIL", ""),

		CORSAllowedOrigins: getEnvAsStringSlice("CORS_ALLOWED_ORIGINS", ",", []string{"http://localhost:5173"}),

		FrontendProxyURL: getEnv("FRONTEND_PROXY_URL", ""),
	}

	if conf.DBPassword == "" {
		panic("Config error: POSTGRES_PASSWORD is required and cannot be empty")
	}
	if conf.DBName == "" {
		panic("Config error: POSTGRES_DB name is required")
	}
	if conf.TLSEnabled && len(conf.TLSDomains) == 0 {
		panic("Config error: TLS_DOMAINS is required (comma-separated) when TLS_ENABLED=true")
	}

	return conf
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return fallback
}

// getEnvAsStringSlice розбиває значення env-змінної на слайс рядків за
// роздільником sep, обрізаючи пробіли та ігноруючи порожні елементи
// (напр. "a.com, b.com," → ["a.com", "b.com"]). Повертає fallback, якщо
// змінна не задана або після розбору не лишилось жодного елемента.
func getEnvAsStringSlice(key, sep string, fallback []string) []string {
	raw := getEnv(key, "")
	if raw == "" {
		return fallback
	}
	out := make([]string, 0)
	for _, part := range strings.Split(raw, sep) {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
