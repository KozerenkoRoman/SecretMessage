package engine_test

import (
	"math/rand"
	"testing"

	"secret-message/cmd/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seededRNG — детермінована реалізація engine.RNG для тестів.
type seededRNG struct{ r *rand.Rand }

func (s *seededRNG) Intn(n int) int { return s.r.Intn(n) }

func newSeededRNG(seed int64) *seededRNG {
	return &seededRNG{r: rand.New(rand.NewSource(seed))}
}

// TestPrepareNewDeck_LengthAndBurn перевіряє, що сумарний розмір
// (deck + burn) дорівнює канонічному розміру колоди Love Letter Premium (22 карти).
func TestPrepareNewDeck_LengthAndBurn(t *testing.T) {
	rng := newSeededRNG(42)
	deck, _ := engine.PrepareNewDeck(rng)

	// 20 карт у колоді + 1 burn = 21 загалом (totalDeckSize).
	require.Len(t, deck, 20, "deck має бути 20 картами (totalDeckSize - 1)")
}

// TestPrepareNewDeck_Composition перевіряє, що мультимножина карт
// (deck + burn) точно відповідає правилам гри: 1 Princess, 1 King, 1 Countess,
// по 2 Chancellor/Prince/Baron/Handmaid/Priest/Spy і 6 Guard.
func TestPrepareNewDeck_Composition(t *testing.T) {
	rng := newSeededRNG(123)
	deck, burn := engine.PrepareNewDeck(rng)

	counts := make(map[engine.CardType]int)
	for _, c := range deck {
		counts[c]++
	}
	counts[burn]++

	expected := map[engine.CardType]int{
		engine.CardPrincess:   1,
		engine.CardKing:       1,
		engine.CardCountess:   1,
		engine.CardChancellor: 2,
		engine.CardPrince:     2,
		engine.CardBaron:      2,
		engine.CardHandmaid:   2,
		engine.CardPriest:     2,
		engine.CardSpy:        2,
		engine.CardGuard:      6,
	}

	assert.Equal(t, expected, counts, "склад колоди має відповідати правилам гри")
}

// TestPrepareNewDeck_Deterministic перевіряє детермінізм:
// з однаковим seed ми завжди отримуємо однакову колоду + однакову burn.
// Це гарантує можливість replay/тесту партій.
func TestPrepareNewDeck_Deterministic(t *testing.T) {
	deck1, burn1 := engine.PrepareNewDeck(newSeededRNG(2024))
	deck2, burn2 := engine.PrepareNewDeck(newSeededRNG(2024))

	assert.Equal(t, burn1, burn2, "burn-карта має бути однаковою при однаковому seed")
	assert.Equal(t, deck1, deck2, "deck має бути однаковим при однаковому seed")
}

// TestPrepareNewDeck_TemplateImmutable перевіряє, що повторні виклики
// з різними RNG не псують внутрішнього шаблону. Якби PrepareNewDeck випадково
// шаффлив сам шаблон, цей тест почав би падати на 2-му виклику.
func TestPrepareNewDeck_TemplateImmutable(t *testing.T) {
	for i := 0; i < 5; i++ {
		deck, burn := engine.PrepareNewDeck(newSeededRNG(int64(i)))
		require.Len(t, deck, 20)

		counts := make(map[engine.CardType]int)
		for _, c := range deck {
			counts[c]++
		}
		counts[burn]++

		// Інваріанти на склад колоди.
		assert.Equal(t, 1, counts[engine.CardPrincess])
		assert.Equal(t, 6, counts[engine.CardGuard])
	}
}
