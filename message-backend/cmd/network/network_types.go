package network

import "time"

const (
	UpdateTypeRoomUpdated = "ROOM_UPDATED"
	UpdateTypeLobbyList   = "LOBBY_LIST_UPDATED"
	// UpdateTypeGameStateSnapshot — повний знімок стану, який надсилається
	// одному гравцю одразу після успішного reconnect, щоб наздогнати
	// пропущені події.
	UpdateTypeGameStateSnapshot = "GAME_STATE_SNAPSHOT"
)

const (
	MsgAction            = "ACTION"
	MsgChancellorResolve = "CHANCELLOR_RESOLVE"
	MsgJoin              = "JOIN"
	MsgStartGame         = "START_GAME"
	MsgLeave             = "LEAVE"
	MsgNextRound         = "NEXT_ROUND"
	// MsgReconnect — клієнт просить відновити активну ігрову сесію,
	// передаючи збережений reconnection_token.
	MsgReconnect = "RECONNECT"
)

const (
	// disconnectGracePeriod — скільки часу гравцю дається на повернення
	// (напр., мобільний застосунок пішов у фон) перш ніж його виключать
	// з активної партії.
	disconnectGracePeriod = 45 * time.Second
)
