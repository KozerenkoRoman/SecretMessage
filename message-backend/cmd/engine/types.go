package engine

import "time"

type (
	CardType  int
	EventType string
	Phase     string
)

const (
	CardSpy        CardType = iota // Шпигун
	CardGuard                      // Вартовий
	CardPriest                     // Священик
	CardBaron                      // Барон
	CardHandmaid                   // Служниця
	CardPrince                     // Принц
	CardChancellor                 // Канцлер
	CardKing                       // Король
	CardCountess                   // Графиня
	CardPrincess                   // Принцеса
)

const (
	PhaseMainAction        Phase         = "MAIN_ACTION"
	PhaseRoundEnd          Phase         = "ROUND_END"
	PhaseResolveChancellor Phase         = "RESOLVE_CHANCELLOR"
	PhaseFinished          Phase         = "FINISHED"
	TurnDuration           time.Duration = 5 * time.Minute
)

type GameState struct {
	Seed        int64             `json:"seed"`
	Sequence    uint64            `json:"seq"`
	Phase       Phase             `json:"phase"`
	Deck        []CardType        `json:"deck"`
	Players     map[string]Player `json:"players"`
	TurnOrder   []string          `json:"turn_order"`
	CurrentTurn int               `json:"current_turn"`

	PendingAction  PendingAction `json:"pending_action,omitempty"`
	IsGameOver     bool          `json:"is_game_over,omitempty"`
	WinnerID       string        `json:"winner_id,omitempty"`
	TransitionHash string        `json:"transition_hash,omitempty"`
	BurnCard       *CardType     `json:"burn_card,omitempty"` // карта, вилучена з колоди на початку партії
}

type Player struct {
	ID               string     `json:"id"`
	Username         string     `json:"username"`
	Hand             []CardType `json:"hand"`
	DiscardPile      []CardType `json:"discard_pile"`
	IsOut            bool       `json:"is_out"`
	IsProtected      bool       `json:"is_protected"`
	Score            int        `json:"score"`
	SpyPointsAwarded bool       `json:"spy_points_awarded"`
	AvatarSeed       string     `json:"avatar_seed"`
}

// Очікуваний екшен (для Chancellor)
type PendingAction struct {
	Type     string `json:"type"`
	PlayerID string `json:"player_id"`
}

// Екшен гравця
type Action struct {
	PlayerID  string   `json:"player_id"`
	HandIndex int      `json:"hand_index"`
	TargetID  string   `json:"target_id,omitempty"`
	Guess     CardType `json:"guess_card,omitempty"`
}

// Екшен резолву Chancellor
type ChancellorResolveAction struct {
	PlayerID      string     `json:"player_id"`
	KeepHandIndex int        `json:"keep_hand_index"`
	BottomOrder   []CardType `json:"bottom_order"`
}

// RNG інтерфейс
type RNG interface {
	Intn(n int) int
}

// Clock інтерфейс
type Clock interface {
	Now() time.Time
}

// ============================================================================
// EVENT PAYLOAD CONTRACT
// ============================================================================
//
// Кожний payload реалізовує маркерний інтерфейс EventPayload.
// Це усуває потребу в `any` всередині DomainEvent та робить компілятор
// нашим союзником: ми не можемо випадково присвоїти map[string]string
// у Payload — компіляція впаде.
//
// Метод Mask(viewerID) повертає копію payload, з якої видалено
// чутливі дані (значення карт), якщо viewer не є учасником події.
// Якщо payload не містить секретів, Mask просто повертає себе.
//
// Усі payloads використовують snake_case JSON-теги (узгоджено з фронтом).
// Числові поля (CardType, points, ...) маршалізуються як JSON-числа,
// оскільки в основі CardType = int, а encoding/json серіалізує int як integer.
// ============================================================================

// EventPayload — маркерний інтерфейс для усіх типизованих payload-ів.
type EventPayload interface {
	// IsEventPayload — суто маркерний метод, не виконує жодної логіки.
	IsEventPayload()
	// Mask повертає payload, безпечний для відправки гравцеві viewerID.
	// Для більшості подій це той самий payload. Для подій з секретною
	// інформацією (PRIEST_EFFECT, ROUND_COMPARED) — копія з вирізаними
	// значеннями карт, якщо viewerID не є учасником.
	Mask(viewerID string) EventPayload
}

// DomainEvent — суворо типізована подія домену.
// Поле Payload завжди реалізовує EventPayload, ніколи не nil-able map.
type DomainEvent struct {
	EventID   uint64       `json:"event_id"`
	Type      EventType    `json:"type"`
	Payload   EventPayload `json:"payload"`
	Timestamp time.Time    `json:"timestamp"`
}

// Результат застосування карти
type ApplyResult struct {
	NewState     GameState     `json:"new_state"`
	DomainEvents []DomainEvent `json:"domain_events"`
}

// Clone виконує глибоке копіювання мапи гравців
func (s GameState) Clone() GameState {
	cloned := s

	if s.TurnOrder != nil {
		cloned.TurnOrder = make([]string, len(s.TurnOrder))
		copy(cloned.TurnOrder, s.TurnOrder)
	}

	if s.Deck != nil {
		cloned.Deck = make([]CardType, len(s.Deck))
		copy(cloned.Deck, s.Deck)
	}

	cloned.Players = make(map[string]Player, len(s.Players))
	for id, player := range s.Players {
		handCopy := make([]CardType, len(player.Hand))
		copy(handCopy, player.Hand)

		discardCopy := make([]CardType, len(player.DiscardPile))
		copy(discardCopy, player.DiscardPile)

		cloned.Players[id] = Player{
			ID:               player.ID,
			Username:         player.Username,
			Hand:             handCopy,
			DiscardPile:      discardCopy,
			IsOut:            player.IsOut,
			IsProtected:      player.IsProtected,
			Score:            player.Score,
			SpyPointsAwarded: player.SpyPointsAwarded,
			AvatarSeed:       player.AvatarSeed,
		}
	}

	return cloned
}

// ============================================================================
// PAYLOADS
// ============================================================================

// CardPlayedPayload — гравець зіграв карту з руки.
// Подія: EventCardPlayed.
type CardPlayedPayload struct {
	PlayerID string   `json:"player_id"`
	Card     CardType `json:"card"`
	TargetID string   `json:"target_id,omitempty"`
}

func (CardPlayedPayload) IsEventPayload()              {}
func (p CardPlayedPayload) Mask(_ string) EventPayload { return p }

// CardDrawnPayload — гравець добрав карту з колоди (або BurnCard).
// Видна тільки самому гравцю; для інших card зрізається до 0.
// Подія: EventCardDrawn.
type CardDrawnPayload struct {
	PlayerID string   `json:"player_id"`
	Card     CardType `json:"card"`
}

func (CardDrawnPayload) IsEventPayload() {}
func (p CardDrawnPayload) Mask(viewerID string) EventPayload {
	if viewerID == p.PlayerID {
		return p
	}
	masked := p
	masked.Card = 0
	return masked
}

// PriestEffectPayload — священик показує карту цілі ініціатору.
// Видна тільки viewer'у та самій цілі (target теж знає свою карту).
// Подія: EventPriestEffect.
type PriestEffectPayload struct {
	ViewerID string   `json:"viewer_id"`
	TargetID string   `json:"target_id"`
	Card     CardType `json:"card"`
}

func (PriestEffectPayload) IsEventPayload() {}
func (p PriestEffectPayload) Mask(viewerID string) EventPayload {
	if viewerID == p.ViewerID || viewerID == p.TargetID {
		return p
	}
	masked := p
	masked.Card = 0
	return masked
}

// PlayerEliminatedPayload — гравець вибуває з раунду.
// Reason: "guard_hit", "baron_lost", "princess_played", "left", тощо.
// Подія: EventPlayerEliminated.
type PlayerEliminatedPayload struct {
	PlayerID string `json:"player_id"`
	Reason   string `json:"reason"`
}

func (PlayerEliminatedPayload) IsEventPayload()              {}
func (p PlayerEliminatedPayload) Mask(_ string) EventPayload { return p }

// GuardHitPayload — гравець вгадав карту через Guard.
// Подія: EventGuardHit.
type GuardHitPayload struct {
	PlayerID string   `json:"player_id"`
	TargetID string   `json:"target_id"`
	Guess    CardType `json:"guess"`
}

func (GuardHitPayload) IsEventPayload()              {}
func (p GuardHitPayload) Mask(_ string) EventPayload { return p }

// GuardMissPayload — гравець НЕ вгадав карту через Guard.
// Подія: EventGuardMiss.
type GuardMissPayload struct {
	PlayerID string   `json:"player_id"`
	TargetID string   `json:"target_id"`
	Guess    CardType `json:"guess"`
}

func (GuardMissPayload) IsEventPayload()              {}
func (p GuardMissPayload) Mask(_ string) EventPayload { return p }

// RoundComparedPayload — порівняння карт двох гравців (Baron або кінець раунду).
// Видно ТІЛЬКИ учасникам порівняння; для решти player_card/target_card зрізаються.
// Подія: EventRoundCompared.
type RoundComparedPayload struct {
	PlayerID   string   `json:"player_id"`
	TargetID   string   `json:"target_id"`
	PlayerCard CardType `json:"player_card"`
	TargetCard CardType `json:"target_card"`
}

func (RoundComparedPayload) IsEventPayload() {}
func (p RoundComparedPayload) Mask(viewerID string) EventPayload {
	if viewerID == p.PlayerID || viewerID == p.TargetID {
		return p
	}
	masked := p
	masked.PlayerCard = 0
	masked.TargetCard = 0
	return masked
}

// SpyBonusPayload — гравець отримав бонусні очки за Spy.
// Подія: EventSpyBonus.
type SpyBonusPayload struct {
	PlayerID string `json:"player_id"`
	Points   int    `json:"points"`
}

func (SpyBonusPayload) IsEventPayload()              {}
func (p SpyBonusPayload) Mask(_ string) EventPayload { return p }

// RoundEndPayload — раунд завершено.
// Подія: EventRoundEnd.
type RoundEndPayload struct {
	WinnerID string `json:"winner_id"`
	Reason   string `json:"reason"`
}

func (RoundEndPayload) IsEventPayload()              {}
func (p RoundEndPayload) Mask(_ string) EventPayload { return p }

// HandsSwappedPayload — гравці помінялися руками (King).
// Подія: EventHandsSwapped.
type HandsSwappedPayload struct {
	PlayerID string `json:"player_id"`
	TargetID string `json:"target_id"`
}

func (HandsSwappedPayload) IsEventPayload()              {}
func (p HandsSwappedPayload) Mask(_ string) EventPayload { return p }

// ChancellorDrawnPayload — гравець добрав карти за ефектом Chancellor.
// Видно тільки самому гравцю; для інших cards зрізається до nil.
// Подія: EventChancellorDrawn.
type ChancellorDrawnPayload struct {
	PlayerID string     `json:"player_id"`
	Cards    []CardType `json:"cards"`
}

func (ChancellorDrawnPayload) IsEventPayload() {}
func (p ChancellorDrawnPayload) Mask(viewerID string) EventPayload {
	if viewerID == p.PlayerID {
		return p
	}
	masked := p
	masked.Cards = nil
	return masked
}

// ChancellorResolvedPayload — гравець визначив, яку карту лишити.
// Видно тільки самому гравцю; для інших kept/bottom_order зрізаються.
// Подія: EventChancellorResolved.
type ChancellorResolvedPayload struct {
	PlayerID    string     `json:"player_id"`
	Kept        CardType   `json:"kept"`
	BottomOrder []CardType `json:"bottom_order"`
}

func (ChancellorResolvedPayload) IsEventPayload() {}
func (p ChancellorResolvedPayload) Mask(viewerID string) EventPayload {
	if viewerID == p.PlayerID {
		return p
	}
	masked := p
	masked.Kept = 0
	masked.BottomOrder = nil
	return masked
}

// BaronResultPayload — підсумок Baron-дуелі (хто переміг, хто вибув).
// Без секретів — карти роздаються через RoundComparedPayload.
// Подія: EventBaronResult.
type BaronResultPayload struct {
	WinnerID string `json:"winner_id"`
	LoserID  string `json:"loser_id"`
}

func (BaronResultPayload) IsEventPayload()              {}
func (p BaronResultPayload) Mask(_ string) EventPayload { return p }

// PlayerLeftPayload — гравець покинув кімнату.
// Подія: EventPlayerLeft.
type PlayerLeftPayload struct {
	PlayerID string `json:"player_id"`
}

func (PlayerLeftPayload) IsEventPayload()              {}
func (p PlayerLeftPayload) Mask(_ string) EventPayload { return p }
