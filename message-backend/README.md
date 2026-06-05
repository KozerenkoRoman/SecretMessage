- GameState - центральна структура, що містить стан гри, колоду, гравців, BurnCard, фазу та ін.
- Player - описує стан кожного гравця (рука, discard, прапорці, очки).
- Action / ChancellorResolveAction - дії гравців.
- DomainEvent - універсальна подія, яка містить конкретний Payload.
- Payload - спеціалізовані структури для різних ефектів карт.
- ApplyResult - результат застосування карти: новий стан + список подій.
- RNG / Clock - інтерфейси для випадковості та часу.


#### Відповідність карт і Payload‑ів
| **Карта** | **Основні Payload‑и** | **Призначення** |
| --- | --- | --- |
| **[Spy](ca://s?q=Spy_card_payloads)** | ``CardPlayedPayload`` | Логування факту розіграшу. |
| **[Guard](ca://s?q=Guard_card_payloads)** | ``CardPlayedPayload``, ``CardRevealedPayload`` (якщо вгадано), ``PlayerEliminatedPayload`` (якщо вгадано Princess) | Вгадування карти іншого гравця. |
| **[Priest](ca://s?q=Priest_card_payloads)** | ``CardPlayedPayload``, ``CardRevealedPayload`` | Дозволяє побачити карту іншого гравця. |
| **[Handmaid](ca://s?q=Handmaid_card_payloads)** | ``CardPlayedPayload`` | Надає захист від таргетування. |
| **[Baron](ca://s?q=Baron_card_payloads)** | ``CardPlayedPayload``, ``CardRevealedPayload``, ``BaronResultPayload``, ``PlayerEliminatedPayload`` | Порівняння карт двох гравців, вибуття слабшого. |
| **[Prince](ca://s?q=Prince_card_payloads)** | ``CardPlayedPayload``, ``PlayerEliminatedPayload`` (якщо скинута Princess), ``CardDrawnPayload`` | Примушує таргета скинути карту і добрати нову. |
| **[Chancellor](ca://s?q=Chancellor_card_payloads)** | ``CardPlayedPayload``, ``ChancellorDrawnPayload``, ``ChancellorResolvedPayload`` | Добір двох карт, вибір однієї, решта вниз колоди. |
| **[Countess](ca://s?q=Countess_card_payloads)** | ``CardPlayedPayload`` | Примусове скидання, якщо в руці King або Prince. |
| **[King](ca://s?q=King_card_payloads)** | ``CardPlayedPayload``, ``HandsSwappedPayload`` | Обмін руками між двома гравцями. |
| **[Princess](ca://s?q=Princess_card_payloads)** | ``CardPlayedPayload``, ``PlayerEliminatedPayload`` | Якщо зіграна або скинута — миттєве вибуття. |

#### FSM фази та переходи
| **Фаза** | **Подія / Payload** | **Наступна фаза** | **Призначення** |
| --- | --- | --- | --- |
| **[PhaseMainAction](ca://s?q=PhaseMainAction_FSM)** | ``CardPlayedPayload`` + ефект карти (Guard, Priest, Baron, Prince, King, Chancellor, Countess, Handmaid, Spy, Princess) | Залишається у MainAction або переходить у спеціальну фазу | Виконання основного ходу гравця. |
| **[PhaseResolveChancellor](ca://s?q=PhaseResolveChancellor_FSM)** | ``ChancellorDrawnPayload`` → ``ChancellorResolvedPayload`` | Повернення у MainAction | Добір двох карт, вибір однієї, решта вниз колоди. |
| **[PhaseRoundEnd](ca://s?q=PhaseRoundEnd_FSM)** | ``PlayerEliminatedPayload``, ``BaronResultPayload``, ``CardRevealedPayload``, ``SpyPointsAwarded`` | Завершення раунду або старт нового | Перевірка, чи лишився один гравець, нарахування бонусів Spy, підрахунок очок. |
| **GameEndCondition** | Перевірка ``Score`` + ``IsGameOver ``= ``true`` | Завершення матчу | Якщо хтось досяг цільових очок (наприклад, 7). |

#### HTTP API
| **HTTP Метод** | **URL Шлях** | **Назва обробника в коді** |
| --- | --- | --- | 
| GET | /api/health | s.HandleHealth | 
| GET | /api/leaderboard|s.HandleGetLeaderboard |
| GET | /api/room | s.HandleGetRoomByID |
| GET | /api/rooms | s.HandleGetRooms |
| GET | /api/user/stats|s.HandleGetUserStats |
| POST | /api/admin/block | s.HandleBlockUser |
| POST | /api/auth | s.HandleAuth | 
| GET | /ws | s.HandleWS |

#### Run sqlc

```
cd storage
sqlc generate
```

#### Тести
 ```
 go test ./cmd/engine -v
 go test -v -run TestHandmaidProtectionLifecycle
```

#### wscat
```npx wscat -c ws://localhost:3000/ws```

#### Build
```
go build -o ./bin/server.exe ./cmd/backend/main.go
./bin/server.exe
```

#### Air
```
go install github.com/air-verse/air@latest 
air init
```
.air.toml

```
#:schema https://json.schemastore.org/any.json

env_files = []
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  # Глобальні налаштування збірки (про всяк випадок)
  cmd = "go build -o ./tmp/main.exe ./cmd/backend/main.go"
  bin = "tmp\\main.exe"
  delay = 1000
  include_ext = ["go", "toml", "json", "html"]
  exclude_dir = ["assets", "tmp", "vendor", "bin", ".git"]
  exclude_regex = ["_test.go"]
  stop_on_error = true
  log = "build-errors.log"

[build.windows]
    cmd = "go build -o ./tmp/main.exe ./cmd/backend/main.go" # Вказали точний шлях замість крапки
    bin = "tmp\\main.exe"
    entrypoint = ["tmp\\main.exe"]
    full_bin = ""
    args_bin = []
    post_cmd = []
    pre_cmd = []

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  mode = ""
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  silent = false
  time = false

[misc]
  clean_on_exit = false

[proxy]
  app_port = 0
  app_start_timeout = 0
  enabled = false
  proxy_port = 0

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```