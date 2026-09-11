package simulation

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"secret-message/cmd/engine"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func waitTiny() { time.Sleep(2 * time.Millisecond) }

func testLogger() *logrus.Logger {
	l := logrus.New()
	l.SetLevel(logrus.PanicLevel) // тиша під час тестів
	return l
}

func twoSmartSlots() []SlotConfig {
	return []SlotConfig{
		{BotID: "sim-bot-0", Strategy: "smart", Username: "Alpha"},
		{BotID: "sim-bot-1", Strategy: "smart", Username: "Beta"},
	}
}

// TestConfigNormalize перевіряє валідацію та клампінг конфігурації.
func TestConfigNormalize(t *testing.T) {
	// Замало слотів.
	if _, err := (Config{Count: 10, Slots: []SlotConfig{{Strategy: "smart"}}}).Normalize(); err == nil {
		t.Fatal("expected error for <2 slots")
	}

	// Забагато слотів.
	tooMany := make([]SlotConfig, MaxPlayers+1)
	if _, err := (Config{Count: 1, Slots: tooMany}).Normalize(); err == nil {
		t.Fatal("expected error for >MaxPlayers slots")
	}

	// Невідома стратегія.
	if _, err := (Config{Count: 1, Slots: []SlotConfig{{Strategy: "smart"}, {Strategy: "wizard"}}}).Normalize(); err == nil {
		t.Fatal("expected error for unknown strategy")
	}

	// Невалідний first_move_bot_id.
	badFirst := Config{Count: 1, Slots: twoSmartSlots(), FirstMoveBotID: "nope"}
	if _, err := badFirst.Normalize(); err == nil {
		t.Fatal("expected error for unknown first_move_bot_id")
	}

	// Count клампиться до межі.
	over := Config{Count: MaxBatchGames + 999, Slots: twoSmartSlots()}
	norm, err := over.Normalize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if norm.Count != MaxBatchGames {
		t.Fatalf("expected count clamped to %d, got %d", MaxBatchGames, norm.Count)
	}

	// Count<=0 стає 1; порожня стратегія стає smart; порожній BotID заповнюється.
	zero := Config{Count: 0, Slots: []SlotConfig{{}, {}}}
	norm, err = zero.Normalize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if norm.Count != 1 {
		t.Fatalf("expected count 1, got %d", norm.Count)
	}
	for _, s := range norm.Slots {
		if s.Strategy != "smart" || s.BotID == "" {
			t.Fatalf("slot defaults not applied: %+v", s)
		}
	}
}

// TestRunGameProducesWinner: одна повна гра має завершитись переможцем і
// згенерувати послідовний, зростаючий лог ходів.
func TestRunGameProducesWinner(t *testing.T) {
	cfg, err := (Config{Count: 1, Slots: twoSmartSlots(), Seed: 42}).Normalize()
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	runner := NewRunner(cfg, cfg.Seed, testLogger())
	rec, err := runner.RunGame(context.Background(), 0, cfg.Seed)
	if err != nil {
		t.Fatalf("run game: %v", err)
	}

	if rec.WinnerBot == "" {
		t.Fatal("expected a winner bot")
	}
	if rec.TotalTurns == 0 {
		t.Fatal("expected at least one move")
	}
	if rec.TotalRounds == 0 {
		t.Fatal("expected at least one round")
	}
	// Кожен зіграний хід має валідну назву карти й монотонний turn_index.
	prev := -1
	for _, mv := range rec.Moves {
		if mv.TurnIndex <= prev {
			t.Fatalf("turn indices must be strictly increasing: %d after %d", mv.TurnIndex, prev)
		}
		prev = mv.TurnIndex
		if CardName(engine.CardType(mv.PlayedCard)) == "Unknown" {
			t.Fatalf("unknown played card int: %d", mv.PlayedCard)
		}
	}
}

// TestDeterminismSameSeed: однаковий seed => ідентичний результат гри.
func TestDeterminismSameSeed(t *testing.T) {
	cfg, _ := (Config{Count: 1, Slots: []SlotConfig{
		{BotID: "b0", Strategy: "random"},
		{BotID: "b1", Strategy: "random"},
	}, Seed: 777}).Normalize()

	r1 := NewRunner(cfg, 777, testLogger())
	rec1, err := r1.RunGame(context.Background(), 0, 777)
	if err != nil {
		t.Fatalf("run1: %v", err)
	}
	r2 := NewRunner(cfg, 777, testLogger())
	rec2, err := r2.RunGame(context.Background(), 0, 777)
	if err != nil {
		t.Fatalf("run2: %v", err)
	}

	if rec1.WinnerBot != rec2.WinnerBot || rec1.TotalTurns != rec2.TotalTurns || rec1.TotalRounds != rec2.TotalRounds {
		t.Fatalf("expected deterministic results for same seed: %+v vs %+v", rec1, rec2)
	}
}

// TestFirstMoveOverride: примусовий перший хід має гарантувати, що вказаний
// слот ходить першим у кожній грі.
func TestFirstMoveOverride(t *testing.T) {
	cfg, err := (Config{
		Count:          1,
		Slots:          twoSmartSlots(),
		Seed:           5,
		FirstMoveBotID: "sim-bot-1",
	}).Normalize()
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	runner := NewRunner(cfg, cfg.Seed, testLogger())
	rec, err := runner.RunGame(context.Background(), 0, cfg.Seed)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if rec.FirstMoveBot != "sim-bot-1" {
		t.Fatalf("expected sim-bot-1 to move first, got %q", rec.FirstMoveBot)
	}
	// Перший зафіксований хід також має належати призначеному боту.
	if len(rec.Moves) > 0 && rec.Moves[0].BotID != "sim-bot-1" {
		t.Fatalf("expected first move by sim-bot-1, got %q", rec.Moves[0].BotID)
	}
}

// TestMetricsAggregation: колектор має коректно рахувати win-rate, суми ходів,
// розподіл карт та середні значення.
func TestMetricsAggregation(t *testing.T) {
	mc := NewMetricsCollector(2)
	slotStrats := map[string]string{"a": "smart", "b": "random"}

	slot0 := 0
	g1 := SimGameRecord{
		GameIndex: 0, WinnerBot: "a", WinnerSlot: &slot0, FirstMoveBot: "a",
		TotalTurns: 4, TotalRounds: 1, DurationMs: 10,
		Moves: []SimMoveRecord{
			{BotID: "a", BotStrategy: "smart", PlayedCard: int(engine.CardGuard), TargetID: "b"},
			{BotID: "b", BotStrategy: "random", PlayedCard: int(engine.CardSpy), InvalidAttempt: true, UsedFallback: true},
			{BotID: "a", BotStrategy: "smart", PlayedCard: int(engine.CardGuard), TargetID: "b", EliminatedReason: "guard_hit"},
		},
	}
	g2 := SimGameRecord{
		GameIndex: 1, WinnerBot: "b", FirstMoveBot: "a",
		TotalTurns: 6, TotalRounds: 2, DurationMs: 20,
		Moves: []SimMoveRecord{
			{BotID: "b", BotStrategy: "random", PlayedCard: int(engine.CardBaron), TargetID: "a"},
		},
	}
	mc.RecordGame(g1, slotStrats)
	mc.RecordGame(g2, slotStrats)

	sum := mc.Snapshot()
	if sum.CompletedGames != 2 {
		t.Fatalf("expected 2 completed games, got %d", sum.CompletedGames)
	}
	if sum.AvgTurns != 5.0 { // (4+6)/2
		t.Fatalf("expected avg turns 5.0, got %v", sum.AvgTurns)
	}
	if sum.AvgRounds != 1.5 { // (1+2)/2
		t.Fatalf("expected avg rounds 1.5, got %v", sum.AvgRounds)
	}
	if sum.TotalInvalid != 1 {
		t.Fatalf("expected 1 invalid move, got %d", sum.TotalInvalid)
	}
	if sum.TotalFallback != 1 {
		t.Fatalf("expected 1 fallback move, got %d", sum.TotalFallback)
	}

	smart := sum.PerStrategy["smart"]
	random := sum.PerStrategy["random"]
	if smart == nil || random == nil {
		t.Fatal("expected both strategies present")
	}
	// smart виграв 1 з 2 ігор -> win rate 0.5.
	if smart.Wins != 1 || smart.WinRate != 0.5 {
		t.Fatalf("smart winrate wrong: wins=%d rate=%v", smart.Wins, smart.WinRate)
	}
	// random виграв 1 з 2.
	if random.Wins != 1 || random.WinRate != 0.5 {
		t.Fatalf("random winrate wrong: wins=%d rate=%v", random.Wins, random.WinRate)
	}
	// smart зіграв Guard двічі.
	if smart.CardPlays["Guard"] != 2 {
		t.Fatalf("expected smart 2 Guard plays, got %d", smart.CardPlays["Guard"])
	}
	// First-move: 'a' (smart) ходив першим в обох іграх, виграв 1.
	if smart.FirstMoveGames != 2 || smart.FirstMoveWins != 1 {
		t.Fatalf("first-move stats wrong: games=%d wins=%d", smart.FirstMoveGames, smart.FirstMoveWins)
	}
	// Причина вибуття агрегована.
	if sum.EliminationsBy["guard_hit"] != 1 {
		t.Fatalf("expected 1 guard_hit elimination, got %d", sum.EliminationsBy["guard_hit"])
	}
}

// fakeStore — in-memory TelemetryStore для тесту контролера без реальної БД.
type fakeStore struct {
	mu       sync.Mutex
	created  int
	games    int
	finished int
}

func (f *fakeStore) CreateSimBatch(_ context.Context, _ uuid.UUID, _ int, _ json.RawMessage) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.created++
	return nil
}
func (f *fakeStore) UpdateSimBatchProgress(_ context.Context, _ uuid.UUID, _ int) error { return nil }
func (f *fakeStore) FinishSimBatch(_ context.Context, _ uuid.UUID, _ string, _ json.RawMessage, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.finished++
	return nil
}
func (f *fakeStore) SaveSimGame(_ context.Context, _ uuid.UUID, _ StorageGame) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.games++
	return nil
}

// TestControllerBatchLifecycle: контролер має запустити батч, довести його до
// completed і зберегти телеметрію у сховище.
func TestControllerBatchLifecycle(t *testing.T) {
	store := &fakeStore{}
	ctrl := NewController(store, testLogger())

	cfg := Config{
		Count:   5,
		Slots:   twoSmartSlots(),
		Seed:    100,
		Persist: true,
	}
	id, err := ctrl.StartBatch(cfg)
	if err != nil {
		t.Fatalf("start batch: %v", err)
	}

	// Чекаємо завершення (батч синхронно-швидкий, але виконується в горутині).
	var final BatchProgress
	for i := 0; i < 200; i++ {
		p, ok := ctrl.Progress(id)
		if !ok {
			t.Fatal("batch progress not found")
		}
		if p.Status == "completed" {
			final = p
			break
		}
		waitTiny()
	}
	if final.Status != "completed" {
		t.Fatalf("batch did not complete in time, last status=%q", final.Status)
	}
	if final.PlayedGames != 5 || final.Summary.CompletedGames != 5 {
		t.Fatalf("expected 5 played games, got played=%d summary=%d", final.PlayedGames, final.Summary.CompletedGames)
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	if store.created != 1 {
		t.Fatalf("expected 1 batch created, got %d", store.created)
	}
	if store.games != 5 {
		t.Fatalf("expected 5 games persisted, got %d", store.games)
	}
	if store.finished != 1 {
		t.Fatalf("expected 1 batch finalized, got %d", store.finished)
	}
}

// TestRunBatchSync: синхронний прогін (для CLI) повертає осмислений підсумок.
func TestRunBatchSync(t *testing.T) {
	sum, err := RunBatchSync(context.Background(), Config{
		Count: 10,
		Slots: []SlotConfig{
			{BotID: "s0", Strategy: "smart"},
			{BotID: "s1", Strategy: "random"},
		},
		Seed: 2024,
	}, testLogger())
	if err != nil {
		t.Fatalf("run batch sync: %v", err)
	}
	if sum.CompletedGames != 10 {
		t.Fatalf("expected 10 completed games, got %d", sum.CompletedGames)
	}
	totalWins := 0
	for _, st := range sum.PerStrategy {
		totalWins += st.Wins
	}
	if totalWins == 0 {
		t.Fatal("expected at least some wins recorded across strategies")
	}
}
