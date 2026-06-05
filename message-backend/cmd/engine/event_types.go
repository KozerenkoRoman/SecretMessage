package engine

const (
	EventCardPlayed         EventType = "CARD_PLAYED"
	EventCardDrawn          EventType = "CARD_DRAWN"
	EventPlayerEliminated   EventType = "PLAYER_ELIMINATED"
	EventHandsSwapped       EventType = "HANDS_SWAPPED"
	EventRoundCompared      EventType = "ROUND_COMPARED"
	EventRoundEnd           EventType = "ROUND_END"
	EventBaronResult        EventType = "BARON_RESULT"
	EventSpyBonus           EventType = "SPY_BONUS"
	EventChancellorDrawn    EventType = "CHANCELLOR_DRAWN"
	EventChancellorResolved EventType = "CHANCELLOR_RESOLVED"
	EventPriestEffect       EventType = "PRIEST_EFFECT"
	EventPlayerLeft         EventType = "PLAYER_LEFT"
	EventGuardHit           EventType = "GUARD_HIT"
	EventGuardMiss          EventType = "GUARD_MISS"
)

// Стандартні reasons для PlayerEliminatedPayload / RoundEndPayload.
// Тримаємо їх як string-константи, щоб уникнути магічних літералів у коді.
const (
	ReasonGuardHit       = "guard_hit"
	ReasonBaronLost      = "baron_lost"
	ReasonPrincessPlayed = "princess_played"
	ReasonPrinceDiscard  = "prince_discard"
	ReasonPlayerLeft     = "left"

	ReasonRoundLastStanding = "last_standing"
	ReasonRoundDeckEmpty    = "deck_empty"
	ReasonRoundOpponentLeft = "opponent_left"
)
