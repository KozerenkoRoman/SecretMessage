// =============================================================================
// SECRET MESSAGE — DOMAIN EVENT TYPES
// =============================================================================
//
// Цей файл є джерелом правди для всіх подій, які бекенд (Go, engine.DomainEvent)
// надсилає фронтенду через WebSocket.
//
// Усі ключі — snake_case (узгоджено з Go json-тегами та форматом наших логів).
// Числові поля (card, guess, points, *_card) — це ЦІЛІ ЧИСЛА (CardType enum
// у Go серіалізується як int), тому фронтенд НЕ повинен виконувати Number()
// перетворення з рядка. Якщо в payload приходить string — це баг бекенду.
//
// Архітектура:
//   1. CardType — enum, дзеркало engine/types.go.
//   2. *Payload інтерфейси — рівно одна на подію, поля 1:1 з Go-структурами.
//   3. EventTypeMap — мапа типу подій → інтерфейс payload.
//   4. GameEvent — discriminated union: коли event.type === 'CARD_PLAYED',
//      компілятор автоматично знає, що event.payload — CardPlayedPayload.
// =============================================================================

// -----------------------------------------------------------------------------
// CardType — дзеркало engine.CardType (iota з Go).
// Порядок ВАЖЛИВИЙ — змінюйте лише разом з backend/cmd/engine/types.go.
// -----------------------------------------------------------------------------
export enum CardType {
  Spy = 0,
  Guard = 1,
  Priest = 2,
  Baron = 3,
  Handmaid = 4,
  Prince = 5,
  Chancellor = 6,
  Countess = 7,
  King = 8,
  Princess = 9,
}

// -----------------------------------------------------------------------------
// Літерали типів подій (значення поля `type` у DomainEvent).
// -----------------------------------------------------------------------------
export type EventType =
  | "CARD_PLAYED"
  | "CARD_DRAWN"
  | "PRIEST_EFFECT"
  | "GUARD_HIT"
  | "GUARD_MISS"
  | "PLAYER_ELIMINATED"
  | "ROUND_COMPARED"
  | "SPY_BONUS"
  | "ROUND_END"
  | "HANDS_SWAPPED"
  | "BARON_RESULT"
  | "CHANCELLOR_DRAWN"
  | "CHANCELLOR_RESOLVED"
  | "PLAYER_LEFT";

// -----------------------------------------------------------------------------
// Стандартні reasons (string-літерали з backend/cmd/engine/event_types.go).
// -----------------------------------------------------------------------------
export type EliminationReason =
  | "guard_hit"
  | "baron_lost"
  | "princess_played"
  | "prince_discard"
  | "left";

export type RoundEndReason =
  | "last_standing"
  | "deck_empty"
  | "opponent_left";

// =============================================================================
// PAYLOADS
// =============================================================================

/**
 * CARD_PLAYED — гравець зіграв карту з руки.
 * `target_id` присутній лише для націлених карт (Guard, Priest, Baron, King,
 * Prince). Для Handmaid/Countess/Spy/Princess поле може бути відсутнім або "".
 */
export interface CardPlayedPayload {
  player_id: string;
  card: CardType;
  target_id?: string;
  discarded_card?: CardType;
}

/**
 * CARD_DRAWN — гравець добрав карту з колоди (або BurnCard).
 * УВАГА: бекенд маскує `card` до 0 для гравців, які не є власником руки.
 * Фронтенд має ігнорувати CARD_DRAWN, якщо `player_id !== self`.
 */
export interface CardDrawnPayload {
  player_id: string;
  card: CardType;
}

/**
 * PRIEST_EFFECT — Священник показав карту цілі ініціатору.
 * Видно тільки viewer_id та target_id; для решти `card` приходить як 0.
 */
export interface PriestEffectPayload {
  viewer_id: string;
  target_id: string;
  card: CardType;
}

/**
 * GUARD_HIT — Вартовий правильно вгадав карту цілі (ціль вибуває).
 * Зазвичай негайно супроводжується подією PLAYER_ELIMINATED з reason="guard_hit".
 */
export interface GuardHitPayload {
  player_id: string;
  target_id: string;
  guess: CardType;
}

/**
 * GUARD_MISS — Вартовий НЕ вгадав карту цілі (нічого не відбувається).
 */
export interface GuardMissPayload {
  player_id: string;
  target_id: string;
  guess: CardType;
}

/**
 * PLAYER_ELIMINATED — гравець вибуває з раунду.
 * `reason` дозволяє UI показати правильну анімацію/текст.
 */
export interface PlayerEliminatedPayload {
  player_id: string;
  reason: EliminationReason | string;
}

/**
 * ROUND_COMPARED — карти двох гравців порівняно (Baron-дуель або кінець раунду).
 * Видно ТІЛЬКИ учасникам; для решти `player_card` та `target_card` = 0.
 */
export interface RoundComparedPayload {
  player_id: string;
  target_id: string;
  player_card: CardType;
  target_card: CardType;
}

/**
 * SPY_BONUS — гравець отримав бонусний бал за Spy.
 * `points` — це ЦІЛЕ ЧИСЛО, не рядок.
 */
export interface SpyBonusPayload {
  player_id: string;
  points: number;
}

/**
 * ROUND_END — раунд завершено, переможець визначений.
 */
export interface RoundEndPayload {
  winner_id: string;
  reason: RoundEndReason | string;
}

/**
 * HANDS_SWAPPED — гравці помінялися руками (King).
 */
export interface HandsSwappedPayload {
  player_id: string;
  target_id: string;
}

/**
 * BARON_RESULT — підсумок Baron-дуелі (без секретних карт; вони у ROUND_COMPARED).
 */
export interface BaronResultPayload {
  winner_id: string;
  loser_id: string;
}

/**
 * CHANCELLOR_DRAWN — гравець добрав 2 карти за ефектом Chancellor.
 * Видно тільки гравцю; для решти `cards` = null.
 */
export interface ChancellorDrawnPayload {
  player_id: string;
  cards: CardType[] | null;
}

/**
 * CHANCELLOR_RESOLVED — гравець визначив, яку карту лишити.
 * Видно тільки гравцю; для решти `kept` = 0 і `bottom_order` = null.
 */
export interface ChancellorResolvedPayload {
  player_id: string;
  kept: CardType;
  bottom_order: CardType[] | null;
}

/**
 * PLAYER_LEFT — гравець покинув кімнату.
 */
export interface PlayerLeftPayload {
  player_id: string;
}

// =============================================================================
// EventTypeMap — точна відповідність type → payload.
// Якщо backend додасть нову подію, її потрібно додати сюди ОДИН РАЗ.
// =============================================================================
export interface EventTypeMap {
  CARD_PLAYED: CardPlayedPayload;
  CARD_DRAWN: CardDrawnPayload;
  PRIEST_EFFECT: PriestEffectPayload;
  GUARD_HIT: GuardHitPayload;
  GUARD_MISS: GuardMissPayload;
  PLAYER_ELIMINATED: PlayerEliminatedPayload;
  ROUND_COMPARED: RoundComparedPayload;
  SPY_BONUS: SpyBonusPayload;
  ROUND_END: RoundEndPayload;
  HANDS_SWAPPED: HandsSwappedPayload;
  BARON_RESULT: BaronResultPayload;
  CHANCELLOR_DRAWN: ChancellorDrawnPayload;
  CHANCELLOR_RESOLVED: ChancellorResolvedPayload;
  PLAYER_LEFT: PlayerLeftPayload;
}

// =============================================================================
// DomainEvent — generic-обгортка (точна копія Go-структури).
// =============================================================================
export interface DomainEvent<T extends EventType = EventType> {
  event_id: number;
  type: T;
  payload: EventTypeMap[T];
  timestamp: string; // RFC3339 (Go time.Time JSON-format)
}

// =============================================================================
// GameEvent — DISCRIMINATED UNION.
//
// Це головний тип, який має використовуватись в обробниках WebSocket.
// Завдяки тому, що поле `type` — це літеральний рядок, TypeScript
// автоматично звужує тип `payload` всередині блоку switch/if:
//
//   switch (ev.type) {
//     case 'CARD_PLAYED':
//        // ev.payload: CardPlayedPayload
//        ev.payload.card; // CardType
//        break;
//     case 'PRIEST_EFFECT':
//        // ev.payload: PriestEffectPayload
//        ev.payload.viewer_id;
//        break;
//   }
// =============================================================================
export type GameEvent = {
  [K in EventType]: {
    event_id: number;
    type: K;
    payload: EventTypeMap[K];
    timestamp: string;
  };
}[EventType];

// =============================================================================
// Type-guards (helper-функції для runtime-перевірки).
// Корисно, коли ви отримуєте DomainEvent з WebSocket і хочете звузити тип
// без switch-statement.
// =============================================================================

/**
 * isEvent — runtime-перевірка типу події з автоматичним звуженням типу.
 *
 * Приклад:
 *   if (isEvent(ev, 'PRIEST_EFFECT')) {
 *     // тут ev.payload типізовано як PriestEffectPayload
 *     console.log(ev.payload.card);
 *   }
 */
export function isEvent<K extends EventType>(
  ev: GameEvent,
  type: K,
): ev is Extract<GameEvent, { type: K }> {
  return ev.type === type;
}

/**
 * isPayloadMasked — допоміжна перевірка, чи поле `card` (або інше секретне)
 * було зрізане бекендом до 0 для цього viewer'а.
 */
export function isCardMasked(card: CardType | number): boolean {
  // 0 = CardType.Spy. Через те що Spy існує реально, ми не можемо просто
  // дивитись на 0 для CARD_DRAWN власної руки. Цей хелпер призначений
  // саме для PRIEST_EFFECT/ROUND_COMPARED/CARD_DRAWN-чужих, де 0 неможливе
  // легітимно (карта Spy ніколи не "відкривається" таким способом, бо Spy
  // не лежить на руці у момент Priest/Baron-розкриттів у звичайному флоу).
  // Якщо у вашому варіанті гри Spy може бути на руці — використовуйте
  // окремий булівський прапорець на бекенді.
  return card === 0;
}
