package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"secret-message/cmd/config"
	"secret-message/cmd/engine"
	"secret-message/cmd/storage/gen"
	"secret-message/cmd/storage/migrations"

	"github.com/exaring/otelpgx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type Storage struct {
	*gen.Queries
	log *logrus.Logger
	cfg *config.Config
	db  *pgxpool.Pool
}

type GameResultInput struct {
	RoomID       string   `json:"room_id"`
	WinnerName   string   `json:"winner_name"`    // Username переможця (може бути порожнім, якщо нічия)
	AllPlayers   []string `json:"all_players"`    // Усі гравці матчу для оновлення статистики
	FinalStateJS []byte   `json:"final_state_js"` // Серіалізований стан гри
}

type RoomResult struct {
	ID    string
	State json.RawMessage
}

// Внутрішня структура для уніфікації даних перед записом в БД
type internalTurnData struct {
	ActionType string
	PlayerID   string
	HandIndex  int
	TargetID   string
	GuessCard  int
}

func New(ctx context.Context, cfg *config.Config, log *logrus.Logger) (*Storage, error) {
	// Конвертуємо наш DSN рядок для конфігурації pgxpool
	poolCfg, err := pgxpool.ParseConfig(cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Налаштування лімітів пулу згідно з твоїм .env конфігом
	poolCfg.MaxConns = int32(cfg.DBPoolMax)
	poolCfg.MaxConnIdleTime = 1 * time.Hour
	poolCfg.MaxConnLifetime = 24 * time.Hour
	poolCfg.ConnConfig.Tracer = otelpgx.NewTracer()

	db, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("sql open: %w", err)
	}

	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}

	if err := otelpgx.RecordStats(db); err != nil {
		return nil, fmt.Errorf("unable to record database stats: %w", err)
	}

	s := &Storage{
		db:      db,
		Queries: gen.New(db),
		cfg:     cfg,
		log:     log,
	}

	// Автоматично запускаємо міграції при підключенні
	if err := s.runMigrations(ctx); err != nil {
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	return s, nil
}

func (s *Storage) Begin(ctx context.Context) (pgx.Tx, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("pool begin: %w", err)
	}
	return tx, nil
}

func (s *Storage) Conn(ctx context.Context) (*pgxpool.Conn, error) {
	db, err := s.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire: %w", err)
	}
	return db, nil
}

func (s *Storage) Close(ctx context.Context) error {
	s.db.Close()
	return nil
}

// runMigrations динамічно читає вміст embedded директорії та послідовно запускає .sql міграції
func (s *Storage) runMigrations(ctx context.Context) error {
	s.log.Info("Аналіз та автоматичний запуск міграцій бази даних...")

	// 1. Створюємо системну таблицю для контролю версій міграцій
	initSchema := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`
	if _, err := s.db.Exec(ctx, initSchema); err != nil {
		return fmt.Errorf("failed to create migration metadata table: %w", err)
	}

	// 2. Динамічно зчитуємо файли з embedded файлової системи migrations.FS
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to read embedded migrations directory: %w", err)
	}

	// 3. Відбираємо лише файли із розширенням .sql
	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}

	// 4. Сортуємо файли за алфавітом/числами (0001_..., 0002_...), щоб гарантувати порядок виконання
	sort.Strings(migrationFiles)

	// 5. Послідовно запускаємо кожну міграцію
	for _, file := range migrationFiles {
		var alreadyApplied bool
		err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)", file).Scan(&alreadyApplied)
		if err != nil {
			return fmt.Errorf("failed to check migration history for %s: %w", file, err)
		}

		// Якщо міграція вже є в базі — просто йдемо далі
		if alreadyApplied {
			continue
		}

		s.log.Infof("Виявлено нову міграцію: %s. Застосування...", file)

		// Зчитуємо текст SQL-запиту з файлу
		sqlContent, err := migrations.FS.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read embedded migration file %s: %w", file, err)
		}

		// Виконуємо міграцію всередині ACID транзакції
		tx, err := s.db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to start migration transaction: %w", err)
		}
		defer tx.Rollback(ctx)

		if _, err := tx.Exec(ctx, string(sqlContent)); err != nil {
			return fmt.Errorf("failed to execute migration content from %s: %w", file, err)
		}

		// Фіксуємо версію в системній таблиці
		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", file); err != nil {
			return fmt.Errorf("failed to record migration completion for %s: %w", file, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", file, err)
		}

		s.log.Infof("Міграція %s успішно застосована!", file)
	}

	s.log.Info("Всі міграції успішно перевірені та синхронізовані.")
	return nil
}

// SaveGameResult записує історію та атомарно оновлює статистику всіх гравців
func (s *Storage) SaveGameResult(ctx context.Context, input GameResultInput) error {
	// 1. Починаємо транзакцію, щоб дані збереглися атомарно
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start tx for game result: %w", err)
	}
	defer tx.Rollback(ctx)

	txQueries := s.Queries.WithTx(tx)

	var winnerUUID uuid.UUID
	var hasWinner bool

	// 2. Якщо є переможець, дізнаємося його UUID за username
	if input.WinnerName != "" {
		usr, err := txQueries.GetUserByUsername(ctx, input.WinnerName)
		if err == nil {
			winnerUUID = usr.ID
			hasWinner = true
		}
	}

	// 3. Записуємо матч в історію ігор (game_history)
	var winnerParam uuid.NullUUID
	if hasWinner {
		winnerParam = uuid.NullUUID{UUID: winnerUUID, Valid: true}
	}

	err = txQueries.LogGameHistory(ctx, gen.LogGameHistoryParams{
		RoomID:     input.RoomID,
		WinnerID:   winnerParam,
		FinalState: input.FinalStateJS,
	})
	if err != nil {
		return fmt.Errorf("failed to log game history: %w", err)
	}

	// 4. Оновлюємо статистику (user_stats) для кожного учасника
	for _, username := range input.AllPlayers {
		usr, err := txQueries.GetUserByUsername(ctx, username)
		if err != nil {
			// Якщо раптом користувача не знайдено в БД, пропускаємо (або логуємо)
			continue
		}

		isWinner := hasWinner && usr.Username == input.WinnerName
		gamesWonIncrement := 0
		if isWinner {
			gamesWonIncrement = 1
		}

		// Розрахунок балів (наприклад: 10 за участь, +50 за перемогу)
		scoreIncrement := 10
		if isWinner {
			scoreIncrement += 50
		}

		// Використовуємо наш UPSERT запит з ON CONFLICT
		err = txQueries.UpdateUserStats(ctx, gen.UpdateUserStatsParams{
			UserID:             usr.ID,
			GamesPlayed:        1, // Додаємо 1 зіграну гру
			GamesWon:           int32(gamesWonIncrement),
			SpyBonusesReceived: 0, // Можна розширити логіку під карти шпигунів, якщо є в двигуні
			TotalScore:         int32(scoreIncrement),
		})
		if err != nil {
			return fmt.Errorf("failed to update stats for user %s: %w", username, err)
		}
	}

	// 5. Коммітимо транзакцію
	return tx.Commit(ctx)
}

// SaveGameState серіалізує GameState в json.RawMessage та викликає згенерований sqlc метод
func (s *Storage) SaveGameState(ctx context.Context, roomID string, state engine.GameState) error {
	rawJSON, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal GameState to JSON: %w", err)
	}

	err = s.SaveRoom(ctx, gen.SaveRoomParams{
		ID:    roomID,
		State: json.RawMessage(rawJSON),
	})
	if err != nil {
		return fmt.Errorf("failed to execute SaveRoom sqlc query: %w", err)
	}

	return nil
}

// GetGameState зчитує запис та повертає десеріалізований об'єкт структури гри
func (s *Storage) GetGameState(ctx context.Context, roomID string) (engine.GameState, error) {
	room, err := s.GetRoom(ctx, roomID)
	if err != nil {
		return engine.GameState{}, fmt.Errorf("failed to execute GetRoom sqlc query: %w", err)
	}

	var state engine.GameState
	if err := json.Unmarshal(room.State, &state); err != nil {
		return engine.GameState{}, fmt.Errorf("failed to unmarshal JSON back to GameState: %w", err)
	}

	return state, nil
}

// LogApplyAction записує стандартний хід гравця (Apply) та його доменні події
func (s *Storage) LogApplyAction(ctx context.Context, roomID string, sequence uint64, action engine.Action, events []engine.DomainEvent) error {
	data := internalTurnData{
		ActionType: "APPLY",
		PlayerID:   action.PlayerID,
		HandIndex:  action.HandIndex,
		TargetID:   action.TargetID,
		GuessCard:  int(action.Guess), // Карта представлена як число
	}
	return s.writeTurnAndEvents(ctx, roomID, sequence, data, events)
}

// LogChancellorResolveAction записує специфічний хід резолву Канцлера
func (s *Storage) LogChancellorResolveAction(ctx context.Context, roomID string, sequence uint64, action engine.ChancellorResolveAction, events []engine.DomainEvent) error {
	data := internalTurnData{
		ActionType: "CHANCELLOR_RESOLVE",
		PlayerID:   action.PlayerID,
		HandIndex:  action.KeepHandIndex, // Мапимо обраний індекс карти, яку залишили
		TargetID:   "",                   // У Канцлера немає таргета при резолві
		GuessCard:  0,                    // Канцлер не вгадує карти
	}
	return s.writeTurnAndEvents(ctx, roomID, sequence, data, events)
}

func (s *Storage) writeTurnAndEvents(ctx context.Context, roomID string, sequence uint64, turn internalTurnData, events []engine.DomainEvent) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start database transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	txQueries := s.Queries.WithTx(tx)

	// Мапимо текстове поле в pgtype.Text
	targetID := pgtype.Text{
		String: turn.TargetID,
		Valid:  turn.TargetID != "",
	}

	var guessCard sql.NullInt32
	if turn.GuessCard > 0 {
		guessCard = sql.NullInt32{Int32: int32(turn.GuessCard), Valid: true}
	}

	// 1. Записуємо хід у таблицю game_turns та отримуємо згенерований UUID ходу
	// Примітка: залежно від генерації sqlc, turnUUID тут буде типу uuid.UUID
	turnUUID, err := txQueries.LogTurn(ctx, gen.LogTurnParams{
		RoomID:     roomID,
		SequenceID: int32(sequence),
		ActionType: turn.ActionType,
		PlayerID:   turn.PlayerID,
		HandIndex:  int32(turn.HandIndex),
		TargetID:   targetID,
		GuessCard:  guessCard,
	})
	if err != nil {
		return fmt.Errorf("db error saving game turn: %w", err)
	}

	// 2. Створюємо правильну структуру uuid.NullUUID згідно з перевизначенням sqlc
	turnIDParam := uuid.NullUUID{
		UUID:  turnUUID,
		Valid: true,
	}

	// 3. Логуємо кожну доменну подію
	for _, event := range events {
		payloadRaw, err := json.Marshal(event.Payload)
		if err != nil {
			return fmt.Errorf("failed to marshal domain event payload: %w", err)
		}

		err = txQueries.LogDomainEvent(ctx, gen.LogDomainEventParams{
			RoomID:    roomID,
			TurnID:    turnIDParam, // Передаємо згенерований нами uuid.NullUUID
			EventID:   int64(event.EventID),
			EventType: string(event.Type),
			Payload:   json.RawMessage(payloadRaw),
		})
		if err != nil {
			return fmt.Errorf("db error saving domain event: %w", err)
		}
	}

	// 4. Підтверджуємо транзакцію
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Storage) SeedAdmin(ctx context.Context, logger *logrus.Logger) error {
	adminUser := s.cfg.AdminUser
	adminEmail := s.cfg.AdminEmail
	adminPass := s.cfg.AdminPass

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	err = s.Queries.SeedAdminUser(ctx, gen.SeedAdminUserParams{
		Username:     adminUser,
		Email:        adminEmail,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		return fmt.Errorf("failed to seed admin user: %w", err)
	}

	logger.Infof("Перевірка Database Seeding: адміністратор '%s' готовий до роботи.", adminUser)
	return nil
}

// GetAllGameResults — Використовується для адмін-панелі (GET /api/admin/games).
// Оскільки в наданому sqlc-файлі немає прямого GetAllGameResults, але є GetUserGameHistory,
// ми створимо метод, який повертає історію ігор. Якщо winner_id порожній (NullUUID{Valid: false}),
// ми можемо отримати загальну історію або історію конкретного гравця.
// Для повноцінного адмін-методу ми передамо порожній WinnerID (або ви можете згенерувати окремий SQL запит).
func (s *Storage) GetAllGameResults(ctx context.Context) ([]gen.GameHistory, error) {
	// Для демонстрації використовуємо GetUserGameHistory із Valid: false, щоб отримати останні матчі,
	// або якщо sqlc налаштовано суворо на фільтрацію, цей метод поверне матчі без переможців (нічиї).
	// Якщо вам потрібні абсолютно всі матчі, додайте в users.sql: `-- name: GetAllGames :many SELECT * FROM game_history ORDER BY played_at DESC`
	params := gen.GetUserGameHistoryParams{
		WinnerID: uuid.NullUUID{Valid: false},
		Limit:    100, // Ліміт для адмінки
		Offset:   0,
	}

	games, err := s.Queries.GetUserGameHistory(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("storage get game history: %w", err)
	}
	return games, nil
}

// BlockUser — адаптує вхідні рядкові дані під sqlc метод BanUser
func (s *Storage) BlockUser(ctx context.Context, userIDStr string, reason string) error {
	// 1. Парсимо string у uuid.UUID
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user uuid format: %w", err)
	}

	// 2. Готуємо параметри для sqlc
	params := gen.BanUserParams{
		ID: userUUID,
		BanReason: sql.NullString{
			String: reason,
			Valid:  reason != "",
		},
	}

	// 3. Викликаємо згенерований sqlc метод
	if err := s.Queries.BanUser(ctx, params); err != nil {
		return fmt.Errorf("failed to execute sqlc BanUser: %w", err)
	}

	return nil
}

// UnblockUser — адаптує вхідні рядкові дані під sqlc метод UnbanUser
func (s *Storage) UnblockUser(ctx context.Context, userIDStr string) error {
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user uuid format: %w", err)
	}

	if err := s.Queries.UnbanUser(ctx, userUUID); err != nil {
		return fmt.Errorf("failed to execute sqlc UnbanUser: %w", err)
	}

	return nil
}

// GetLeaderboardTop — обгортка для отримання лідерборду з фіксованим лімітом
func (s *Storage) GetLeaderboardTop(ctx context.Context, limit int32) ([]gen.GetLeaderboardRow, error) {
	if limit <= 0 {
		limit = 10 // значення за замовчуванням
	}

	leaderboard, err := s.Queries.GetLeaderboard(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}
	return leaderboard, nil
}

func (s *Storage) GetRoomByID(ctx context.Context, id string) (RoomResult, error) {
	dbRoom, err := s.Queries.GetRoom(ctx, id)
	if err != nil {
		return RoomResult{}, err
	}

	return RoomResult{
		ID:    dbRoom.ID,
		State: dbRoom.State,
	}, nil
}

func (s *Storage) UpsertRoomState(ctx context.Context, id string, state json.RawMessage) error {
	return s.Queries.SaveRoom(ctx, gen.SaveRoomParams{
		ID:    id,
		State: state,
	})
}

func (s *Storage) DeleteRoomByID(ctx context.Context, id string) error {
	return s.Queries.DeleteRoom(ctx, id)
}
