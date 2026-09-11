-- =============================================================================
-- 0005_simulations.up.sql
--
-- Телеметрія для headless Bot-vs-Bot симуляцій.
--
-- Аудит наявної схеми:
--   • rooms        — лише поточний JSON-стан кімнати (не для аналітики).
--   • game_events  — сирий per-turn лог (room_id, turn_id, event_type, payload,
--                    state_before). Придатний для replay, але НЕ агрегує метрики
--                    й НЕ знає про тип/стратегію бота, тривалість ходу, причину
--                    вибуття у структурованому вигляді.
--   • user_stats   — глобальна агрегація по реальних користувачах (не по батчах
--                    симуляції, не по стратегіях).
--
-- Чого бракувало для пост-симуляційного аналізу і що додаємо тут:
--   1. sim_batches   — один запис на запуск батча (N ігор), його конфіг і статус.
--   2. sim_games     — один запис на зіграну гру: переможець, к-сть ходів,
--                      тривалість, seed, хто робив перший хід.
--   3. sim_moves     — один запис на ХІД бота: стратегія, зіграна карта, ціль,
--                      guess, тривалість прийняття рішення, чи був хід невалідним
--                      / чи спрацював fallback, причина вибуття (якщо ця дія
--                      призвела до вибуття).
-- =============================================================================

CREATE TABLE IF NOT EXISTS sim_batches (
    id           uuid PRIMARY KEY,
    status       text        NOT NULL DEFAULT 'running', -- running | completed | failed | cancelled
    total_games  int         NOT NULL,
    played_games int         NOT NULL DEFAULT 0,
    config       jsonb       NOT NULL DEFAULT '{}'::jsonb, -- сирий конфіг запуску (стратегії, first-move override, тощо)
    summary      jsonb       NOT NULL DEFAULT '{}'::jsonb, -- агрегований результат (win rates, avg turns, ...)
    error        text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    finished_at  timestamptz
);

CREATE INDEX IF NOT EXISTS idx_sim_batches_created_at ON sim_batches(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sim_batches_status ON sim_batches(status);

CREATE TABLE IF NOT EXISTS sim_games (
    id             bigserial PRIMARY KEY,
    batch_id       uuid        NOT NULL REFERENCES sim_batches(id) ON DELETE CASCADE,
    game_index     int         NOT NULL,               -- порядковий номер гри в батчі
    seed           bigint      NOT NULL,               -- seed RNG цієї гри (для відтворюваності)
    winner_slot    int,                                -- індекс слота-переможця у turn_order (NULL якщо нічия)
    winner_bot     text,                               -- ім'я/тип бота-переможця
    spy_winner_bot text,                               -- бот, що отримав бонус Шпигуна (може бути NULL)
    total_turns    int         NOT NULL DEFAULT 0,     -- скільки ходів було зіграно
    total_rounds   int         NOT NULL DEFAULT 0,     -- скільки раундів тривала гра
    first_move_bot text,                               -- хто робив перший хід (для first-move override аналізу)
    duration_ms    bigint      NOT NULL DEFAULT 0,     -- час виконання гри (headless, без мережевих затримок)
    created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_sim_games_batch ON sim_games(batch_id);

CREATE TABLE IF NOT EXISTS sim_moves (
    id             bigserial PRIMARY KEY,
    game_id        bigint      NOT NULL REFERENCES sim_games(id) ON DELETE CASCADE,
    turn_index     int         NOT NULL,               -- глобальний номер ходу в грі
    round_index    int         NOT NULL DEFAULT 0,
    bot_id         text        NOT NULL,               -- ID слота-бота, що ходив
    bot_strategy   text        NOT NULL,               -- профіль/стратегія бота
    played_card    int         NOT NULL,               -- CardType, що зіграно
    target_id      text,                               -- ціль (якщо була)
    guess_card     int,                                -- вгадана карта (для Вартового)
    decision_ms    bigint      NOT NULL DEFAULT 0,     -- час, витрачений на прийняття рішення
    invalid_attempt boolean    NOT NULL DEFAULT false, -- рушій відхилив хід (невалідний)
    used_fallback   boolean    NOT NULL DEFAULT false, -- спрацював запасний (fallback) вибір
    eliminated_reason text,                            -- якщо цей хід призвів до вибуття когось
    created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_sim_moves_game ON sim_moves(game_id);
CREATE INDEX IF NOT EXISTS idx_sim_moves_strategy ON sim_moves(bot_strategy);
CREATE INDEX IF NOT EXISTS idx_sim_moves_card ON sim_moves(played_card);
