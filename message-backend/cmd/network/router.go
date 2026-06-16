package network

import "net/http"

func InitRoutes(mux *http.ServeMux, s *Server) {
	s.PUBLIC_GET(mux, "/api/health", s.HandleHealth)
	s.PUBLIC_POST(mux, "/api/auth", s.HandleAuth)
	s.PUBLIC_POST(mux, "/api/register", s.HandleRegister)

	s.GET(mux, "/api/leaderboard", s.HandleGetLeaderboard)
	s.GET(mux, "/api/user/stats", s.HandleGetUserStats)

	// Робота з кімнатами
	s.GET(mux, "/api/room", s.HandleGetRoomByID)
	s.GET(mux, "/api/rooms", s.HandleGetRooms)
	s.POST(mux, "/api/rooms", s.HandleCreateRoom)
	s.POST(mux, "/api/rooms/{id}/bot", s.HandleAddBotToRoom)

	s.POST(mux, "/api/user", s.HandleUpdateUser)

	s.ADMIN_GET(mux, "/api/admin/games", nil) // todo
	s.ADMIN_GET(mux, "/api/admin/users", s.HandleAdminGetUsers)
	s.ADMIN_POST(mux, "/api/admin/users/{id}/block", s.HandleAdminBlockUser)
	s.ADMIN_POST(mux, "/api/admin/block", s.HandleBlockUser)

	mux.Handle("/ws", AuthMiddleware(http.HandlerFunc(s.HandleWS)))
}
