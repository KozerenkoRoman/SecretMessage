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
	RoomID       string        `json:"room_id"`
	WinnerID     uuid.NullUUID `json:"winner_id"`
	SpyWinnerID  uuid.NullUUID `json:"spy_winner_id"`
	AllPlayers   []string      `json:"all_players"`
	FinalStateJS []byte        `json:"final_state_js"`
}

type RoomResult struct {
	ID    string
	State json.RawMessage
}

var SystemBotNames = []gen.User{
	{ID: uuid.MustParse("00000000-aaaa-0000-0000-111122223333"), Username: "BotAlpha", Email: "bot_alpha@local.host", AvatarSeed: "BotAlpha"},
	{ID: uuid.MustParse("00000000-bbbb-0000-0000-111122223333"), Username: "BotBeta", Email: "bot_beta@local.host", AvatarSeed: "BotBeta"},
	{ID: uuid.MustParse("00000000-cccc-0000-0000-111122223333"), Username: "BotGamma", Email: "bot_gamma@local.host", AvatarSeed: "BotGamma"},
	{ID: uuid.MustParse("00000000-dddd-0000-0000-111122223333"), Username: "BotDelta", Email: "bot_delta@local.host", AvatarSeed: "BotDelta"},
	{ID: uuid.MustParse("00000000-eeee-0000-0000-111122223333"), Username: "BotEpsilon", Email: "bot_epsilon@local.host", AvatarSeed: "BotEpsilon"},
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
	poolCfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET client_encoding TO 'UTF8';")
		return err
	}

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

	// Розраховуємо та оновлюємо глобальну ігрову статистику для кожного учасника
	for _, username := range input.AllPlayers {
		usr, err := txQueries.GetUserByUsername(ctx, username)
		if err != nil {
			s.log.Errorf("Failed to get user data for stats update %s:%v", username, err)
			continue
		}

		// Перевірка на звичайного переможця (порівнюємо UUID)
		isWinner := input.WinnerID.Valid && usr.ID == input.WinnerID.UUID
		gamesWonIncrement := 0
		if isWinner {
			gamesWonIncrement = 1
		}

		// Перевірка на переможця-шпигуна (порівнюємо UUID)
		isSpyWinner := input.SpyWinnerID.Valid && usr.ID == input.SpyWinnerID.UUID
		spyBonusesIncrement := 0
		if isSpyWinner {
			spyBonusesIncrement = 1
		}

		// Розрахунок балів: 10 за участь, +50 за перемогу, +25 за шпигуна
		scoreIncrement := 10
		if isWinner {
			scoreIncrement += 50
		}
		if isSpyWinner {
			scoreIncrement += 25
		}

		// Виконуємо атомарний атомарний UPSERT через генератор SQLc
		// Оскільки таблиця тепер містить RoundsPlayed та RoundsWon (з 0004 міграції), інкрементуємо їх за замовчуванням
		// (Ці лічильники можна точніше наповнювати, якщо передавати з рушія кімнати)
		err = txQueries.UpdateUserStats(ctx, gen.UpdateUserStatsParams{
			UserID:       usr.ID,
			GamesPlayed:  1,
			GamesWon:     gamesWonIncrement,
			RoundsPlayed: 1,
			RoundsWon:    gamesWonIncrement,
			SpyBonuses:   spyBonusesIncrement,
			TotalScore:   scoreIncrement,
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

func (s *Storage) LogApplyAction(ctx context.Context, roomID string, turnID int, events []engine.DomainEvent, stateBefore engine.GameState) ([]engine.DomainEvent, error) {
	return s.writeEventsLog(ctx, roomID, turnID, events, stateBefore)
}

func (s *Storage) LogChancellorResolveAction(ctx context.Context, roomID string, turnID int, events []engine.DomainEvent, stateBefore engine.GameState) ([]engine.DomainEvent, error) {
	return s.writeEventsLog(ctx, roomID, turnID, events, stateBefore)
}

func (s *Storage) writeEventsLog(ctx context.Context, roomID string, turnID int, events []engine.DomainEvent, stateBefore engine.GameState) ([]engine.DomainEvent, error) {
	if len(events) == 0 {
		return nil, nil
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start database transaction:%w", err)
	}
	defer tx.Rollback(ctx)

	txQueries := s.Queries.WithTx(tx)

	stateBeforeRaw, err := json.Marshal(stateBefore)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state_before snapshot:%w", err)
	}

	// Створюємо новий слайс, де збережемо івенти з реальними ID з бази
	savedEvents := make([]engine.DomainEvent, len(events))

	for i, event := range events {
		payloadRaw, err := json.Marshal(event.Payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal domain event payload:%w", err)
		}

		// 🌟 sqlc згенерував InsertGameEvent так, що він повертає (int, error), бо в кінці запиту стоїть RETURNING event_id
		actualEventID, err := txQueries.InsertGameEvent(ctx, gen.InsertGameEventParams{
			RoomID:      roomID,
			TurnID:      turnID,
			EventType:   string(event.Type),
			Payload:     json.RawMessage(payloadRaw),
			StateBefore: json.RawMessage(stateBeforeRaw),
		})
		if err != nil {
			return nil, fmt.Errorf("db error saving custom game event: %w", err)
		}

		// Записуємо згенерований базою ID назад у структуру
		event.EventID = actualEventID
		savedEvents[i] = event
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return savedEvents, nil
}

func (s *Storage) GetBotNames() []gen.User {
	res := make([]gen.User, len(SystemBotNames))
	copy(res, SystemBotNames)
	return res
}

func (s *Storage) SeedUsers(ctx context.Context, logger *logrus.Logger) {
	if err := s.seedAdmin(ctx, logger); err != nil {
		logger.Errorf("Попередження: не вдалося виконати початкове заповнення адміна: %v", err)
	}

	for _, botName := range SystemBotNames {
		if err := s.seedBot(ctx, logger, botName); err != nil {
			logger.Errorf("Попередження: не вдалося виконати початкове заповнення бота %s: %v", botName.Username, err)
		}
	}
}

func (s *Storage) seedAdmin(ctx context.Context, logger *logrus.Logger) error {
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
		AvatarSeed:   uuid.New().String(),
	})
	if err != nil {
		return fmt.Errorf("failed to seed admin user: %w", err)
	}

	logger.Infof("Перевірка Database Seeding: адміністратор '%s' готовий до роботи.", adminUser)
	return nil
}

func (s *Storage) seedBot(ctx context.Context, logger *logrus.Logger, bot gen.User) error {
	botPass := fmt.Sprintf("%s@%s", bot.Username, s.cfg.AdminPass)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(botPass), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash bot password: %w", err)
	}

	err = s.Queries.SeedBotUser(ctx, gen.SeedBotUserParams{
		ID:           bot.ID,
		Username:     bot.Username,
		Email:        bot.Email,
		PasswordHash: string(hashedPassword),
		AvatarSeed:   bot.AvatarSeed,
	})
	if err != nil {
		return fmt.Errorf("failed to seed static bot user %s: %w", bot.Username, err)
	}

	logger.Infof("Перевірка Database Seeding: бот '%s' [ID: %s] готовий до роботи.", bot.Username, bot.ID.String())
	return nil
}

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
func (s *Storage) GetLeaderboardTop(ctx context.Context, limit int) ([]gen.GetLeaderboardRow, error) {
	if limit <= 0 {
		limit = 10 // значення за замовчуванням
	}

	leaderboard, err := s.Queries.GetLeaderboard(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	if len(leaderboard) == 0 {
		leaderboard = make([]gen.GetLeaderboardRow, 0)
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
