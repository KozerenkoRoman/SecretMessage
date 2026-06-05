// cmd/network/router_helpers.go
package network

import (
	"net/http"
)

// GET реєструє маршрут для методу GET та автоматично огортає його в AuthMiddleware
func (s *Server) GET(mux *http.ServeMux, path string, handler http.HandlerFunc) {
	mux.Handle("GET "+path, AuthMiddleware(handler))
}

// POST реєструє маршрут для методу POST та автоматично огортає його в AuthMiddleware
func (s *Server) POST(mux *http.ServeMux, path string, handler http.HandlerFunc) {
	mux.Handle("POST "+path, AuthMiddleware(handler))
}

// PUBLIC_POST реєструє публічний маршрут для POST (наприклад, для авторизації) без AuthMiddleware
func (s *Server) PUBLIC_POST(mux *http.ServeMux, path string, handler http.HandlerFunc) {
	mux.HandleFunc("POST "+path, handler)
}

// PUBLIC_GET реєструє публічний маршрут для GET (наприклад, health check) без AuthMiddleware
func (s *Server) PUBLIC_GET(mux *http.ServeMux, path string, handler http.HandlerFunc) {
	mux.HandleFunc("GET "+path, handler)
}

// ADMIN_GET реєструє захищений маршрут для адміна (метод GET)
func (s *Server) ADMIN_GET(mux *http.ServeMux, path string, handler http.HandlerFunc) {
	// Створюємо ланцюжок: Auth -> AdminOnly -> Handler
	chain := AuthMiddleware(AdminOnlyMiddleware(handler))
	mux.Handle("GET "+path, chain)
}

// ADMIN_POST реєструє захищений маршрут для адміна (метод POST)
func (s *Server) ADMIN_POST(mux *http.ServeMux, path string, handler http.HandlerFunc) {
	chain := AuthMiddleware(AdminOnlyMiddleware(handler))
	mux.Handle("POST "+path, chain)
}
