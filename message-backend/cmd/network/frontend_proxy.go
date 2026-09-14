// =============================================================================
// network/frontend_proxy.go
//
// Реверс-проксі на внутрішній SPA-контейнер (nginx, що роздає лише статичні
// файли фронтенду). Використовується, коли Go-бекенд є ЄДИНОЮ публічною
// точкою входу (тримає порти 80/443 і сам термінує TLS через autocert) —
// усі запити, що НЕ підпадають під /api/* чи /ws, проксіюються на SPA.
// =============================================================================

package network

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/sirupsen/logrus"
)

// NewFrontendProxyHandler створює http.Handler, що проксіює усі запити на
// targetURL (напр. "http://frontend:80" — внутрішній nginx-контейнер зі
// статикою SPA). Помилки проксіювання логуються, клієнту повертається 502.
func NewFrontendProxyHandler(targetURL string, logger *logrus.Logger) (http.Handler, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorLog = nil // логуємо самі нижче, щоб мати єдиний формат (logrus)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.WithError(err).Warnf("Помилка проксіювання запиту %s на фронтенд (%s)", r.URL.Path, targetURL)
		w.WriteHeader(http.StatusBadGateway)
	}

	return proxy, nil
}
