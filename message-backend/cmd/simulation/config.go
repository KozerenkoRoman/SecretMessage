// =============================================================================
// simulation/config.go
//
// Конфігурація headless Bot-vs-Bot батча + фабрика стратегій.
// =============================================================================

package simulation

import (
	"fmt"
	"math/rand"

	"github.com/sirupsen/logrus"
)

const (
	// MinPlayers / MaxPlayers — межі кількості слотів у симуляції
	// (узгоджено з правилами Love Letter та MAX_PLAYERS у room.go).
	MinPlayers = 2
	MaxPlayers = 4

	// MaxBatchGames — жорсткий стеля на кількість ігор у батчі, щоб один
	// запит не з'їв усі ресурси. Клієнтський `count` клампиться до цієї межі.
	MaxBatchGames = 5000

	// hardTurnLimit — запобіжник від нескінченного циклу в одній грі
	// (теоретично неможливо через вичерпання колоди, але страхуємось).
	hardTurnLimit = 2000
)

// SlotConfig описує один слот-бот у симуляції.
type SlotConfig struct {
	// BotID — стабільний ідентифікатор слота (напр., "sim-bot-0").
	BotID string `json:"bot_id"`
	// Strategy — профіль поведінки: "smart" | "random".
	Strategy string `json:"strategy"`
	// Username — людсько-читабельне ім'я для звітів (опційно).
	Username string `json:"username"`
}

// Config — параметри запуску батча.
type Config struct {
	// Count — скільки ігор зіграти (клампиться у [1, MaxBatchGames]).
	Count int `json:"count"`
	// Slots — конфігурація ботів по слотах (2..4).
	Slots []SlotConfig `json:"slots"`
	// Seed — базовий seed. Кожна гра i використовує Seed+i (відтворюваність).
	// Якщо 0 — генерується випадковий базовий seed.
	Seed int64 `json:"seed"`
	// FirstMoveBotID — примусово змушує вказаний слот ходити першим у КОЖНІЙ
	// грі (тестування початкових умов). Порожньо = звичайний порядок.
	FirstMoveBotID string `json:"first_move_bot_id"`
	// Persist — чи зберігати детальну телеметрію у БД.
	Persist bool `json:"persist"`
}

// Normalize валідовує та підганяє конфіг під безпечні межі, повертаючи копію.
func (c Config) Normalize() (Config, error) {
	out := c

	if out.Count <= 0 {
		out.Count = 1
	}
	if out.Count > MaxBatchGames {
		out.Count = MaxBatchGames
	}

	if len(out.Slots) < MinPlayers {
		return out, fmt.Errorf("simulation requires at least %d bot slots, got %d", MinPlayers, len(out.Slots))
	}
	if len(out.Slots) > MaxPlayers {
		return out, fmt.Errorf("simulation supports at most %d bot slots, got %d", MaxPlayers, len(out.Slots))
	}

	seenIDs := make(map[string]bool)
	for i := range out.Slots {
		if out.Slots[i].BotID == "" {
			out.Slots[i].BotID = fmt.Sprintf("sim-bot-%d", i)
		}
		if seenIDs[out.Slots[i].BotID] {
			return out, fmt.Errorf("duplicate bot slot id: %s", out.Slots[i].BotID)
		}
		seenIDs[out.Slots[i].BotID] = true

		if out.Slots[i].Strategy == "" {
			out.Slots[i].Strategy = "smart"
		}
		if !isKnownStrategy(out.Slots[i].Strategy) {
			return out, fmt.Errorf("unknown strategy %q for slot %s", out.Slots[i].Strategy, out.Slots[i].BotID)
		}
		if out.Slots[i].Username == "" {
			out.Slots[i].Username = out.Slots[i].BotID
		}
	}

	if out.FirstMoveBotID != "" && !seenIDs[out.FirstMoveBotID] {
		return out, fmt.Errorf("first_move_bot_id %q does not match any slot", out.FirstMoveBotID)
	}

	return out, nil
}

func isKnownStrategy(name string) bool {
	switch name {
	case "smart", "random":
		return true
	default:
		return false
	}
}

// buildStrategy створює екземпляр стратегії для слота з детермінованим rng.
func buildStrategy(slot SlotConfig, rng *rand.Rand, logger *logrus.Logger) Strategy {
	switch slot.Strategy {
	case "random":
		return NewRandomStrategy(slot.BotID, rng)
	default: // "smart"
		return NewSmartStrategy(slot.BotID, logger)
	}
}
