// =============================================================================
// tlsutil/autocert.go
//
// Автоматичне отримання та оновлення TLS-сертифікатів через Let's Encrypt
// (ACME HTTP-01) за допомогою golang.org/x/crypto/acme/autocert.
//
// Дизайн:
//   - NewManager будує *autocert.Manager із білим списком доменів
//     (HostPolicy) — обов'язково, інакше Let's Encrypt видасть сертифікат
//     будь-якому домену, DNS якого вказує на цей сервер (типова помилка
//     конфігурації autocert).
//   - Кеш сертифікатів — локальна директорія (autocert.DirCache), щоб
//     сертифікати переживали перезапуск процесу і не витрачали ліміти
//     Let's Encrypt Rate Limits на кожен деплой.
//   - Умови використання Let's Encrypt приймаються автоматично
//     (autocert.AcceptTOS) — це ОБОВ'ЯЗКОВА умова видачі сертифікату,
//     тому пропускати її сенсу нема: якщо TLS_ENABLED=true, оператор
//     свідомо погоджується з ToS Let's Encrypt.
// =============================================================================

package tlsutil

import (
	"fmt"
	"os"

	"golang.org/x/crypto/acme/autocert"
)

// NewManager створює сконфігурований autocert.Manager для заданого білого
// списку доменів (domains) із персистентним кешем сертифікатів у cacheDir.
//
// email — контактна адреса, яку Let's Encrypt використовує для сповіщень
// про протермінування чи проблеми з відкликанням сертифікатів (необов'язково,
// але рекомендовано).
//
// Повертає помилку, якщо domains порожній (autocert без HostPolicy —
// небезпечна конфігурація) або якщо не вдалося створити cacheDir.
func NewManager(domains []string, cacheDir, email string) (*autocert.Manager, error) {
	if len(domains) == 0 {
		return nil, fmt.Errorf("tlsutil: at least one domain is required for autocert HostPolicy")
	}

	if cacheDir == "" {
		cacheDir = "./certs-cache"
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("tlsutil: failed to create cert cache directory %q: %w", cacheDir, err)
	}

	manager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domains...),
		Cache:      autocert.DirCache(cacheDir),
		Email:      email,
	}

	return manager, nil
}
