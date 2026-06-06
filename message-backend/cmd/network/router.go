// cmd/network/router.go
package network

import "net/http"

func InitRoutes(mux *http.ServeMux, s *Server) {
	// 1. Публічні ендпоінти
	s.PUBLIC_GET(mux, "/api/health", s.HandleHealth)
	s.PUBLIC_POST(mux, "/api/auth", s.HandleAuth)

	// 2. Захищені ендпоінти звичайних гравців (AuthMiddleware)
	s.GET(mux, "/api/leaderboard", s.HandleGetLeaderboard)
	s.GET(mux, "/api/user/stats", s.HandleGetUserStats)
	s.GET(mux, "/api/room", s.HandleGetRoomByID)
	s.GET(mux, "/api/rooms", s.HandleGetRooms)
	s.POST(mux, "/api/rooms", s.HandleCreateRoom)
	s.POST(mux, "/api/user/avatar", s.HandleUpdateAvatar)

	// 3. АДМІНІСТРАТИВНІ ЕНДПОЇНТИ (AuthMiddleware + AdminOnlyMiddleware автоматично)
	s.ADMIN_GET(mux, "/api/admin/games", s.HandleAdminGetGames)              // Перегляд логів матчів
	s.ADMIN_GET(mux, "/api/admin/users", s.HandleAdminGetUsers)              // Список усіх користувачів
	s.ADMIN_POST(mux, "/api/admin/users/{id}/block", s.HandleAdminBlockUser) // Блокування за ID у шляху

	// Старий ендпоінт блокування (залишаємо для сумісності або замінюємо новим)
	s.ADMIN_POST(mux, "/api/admin/block", s.HandleBlockUser)

	// 4. Транспортні протоколи
	mux.Handle("/ws", AuthMiddleware(http.HandlerFunc(s.HandleWS)))
}
