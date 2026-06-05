// =============================================================================
// types/errors.ts
//
// КОНТРАКТ КОДІВ ПОМИЛОК між Go-бекендом та Vue/TS-фронтендом.
//
// Дзеркало файлу `cmd/engine/errors.go`. Якщо ви додаєте/перейменовуєте
// код у Go — синхронізуйте цей файл.
//
// Принципи:
//   • Бекенд НІКОЛИ не надсилає сирий локалізований текст помилки.
//     Замість цього він шле стабільний string-код у полі `error_code`
//     WS-пакета (див. websocket.go::sendEngineError).
//   • Фронтенд читає `error_code`, мапить його у переклад через
//     `i18n/errorMessages.ts` і показує користувачеві.
//   • Будь-який невідомий/відсутній код → ERR_INTERNAL → "Сталася помилка".
//
// Типізація:
//   • EngineErrorCode — string-літеральний union (compile-time safety).
//   • Усі функції-маппери приймають `EngineErrorCode | string`, щоб
//     безпечно витримати майбутні нові коди з бекенду без падіння UI.
// =============================================================================

// -----------------------------------------------------------------------------
// EngineErrorCode — точна копія констант з cmd/engine/errors.go.
// Тримайте у тому ж порядку, що й Go-файл, для зручної ревізії.
// -----------------------------------------------------------------------------
export const EngineErrorCodes = {
  // --- Загальні / службові ---------------------------------------------------
  Internal: "ERR_INTERNAL",

  // --- Валідація стану гри ---------------------------------------------------
  InvalidPhase: "ERR_INVALID_PHASE",
  InvalidState: "ERR_INVALID_STATE",
  EmptyTurnOrder: "ERR_EMPTY_TURN_ORDER",
  CurrentTurnOutOfRange: "ERR_CURRENT_TURN_OUT_OF_RANGE",
  NoPlayers: "ERR_NO_PLAYERS",
  CardEffectNotImplemented: "ERR_CARD_EFFECT_NOT_IMPLEMENTED",

  // --- Валідація гравця ------------------------------------------------------
  PlayerNotFound: "ERR_PLAYER_NOT_FOUND",
  PlayerAlreadyOut: "ERR_PLAYER_ALREADY_OUT",
  PlayerProtected: "ERR_PLAYER_PROTECTED",
  PlayerHasNoCards: "ERR_PLAYER_HAS_NO_CARDS",

  // --- Валідація дії ---------------------------------------------------------
  OutOfTurn: "ERR_OUT_OF_TURN",
  InvalidHandIndex: "ERR_INVALID_HAND_INDEX",
  CannotTargetSelf: "ERR_CANNOT_TARGET_SELF",
  MustPlayCountess: "ERR_MUST_PLAY_COUNTESS",

  // --- Валідація цілі --------------------------------------------------------
  TargetRequired: "ERR_TARGET_REQUIRED",
  TargetNotFound: "ERR_TARGET_NOT_FOUND",
  TargetAlreadyOut: "ERR_TARGET_ALREADY_OUT",
  TargetProtected: "ERR_TARGET_PROTECTED",

  // --- Карто-специфічні валідації -------------------------------------------
  GuardCannotGuessGuard: "ERR_GUARD_CANNOT_GUESS_GUARD",
  GuardGuessRequired: "ERR_GUARD_GUESS_REQUIRED",
  BaronNoCardsToCompare: "ERR_BARON_NO_CARDS_TO_COMPARE",

  // --- Chancellor (резолв) ---------------------------------------------------
  ChancellorInvalidBottomOrder: "ERR_CHANCELLOR_INVALID_BOTTOM_ORDER",
  ChancellorWrongPhase: "ERR_CHANCELLOR_WRONG_PHASE",
} as const;

// EngineErrorCode — string-літеральний union усіх валідних кодів.
export type EngineErrorCode = (typeof EngineErrorCodes)[keyof typeof EngineErrorCodes];

// -----------------------------------------------------------------------------
// Структура error-пакета, який бекенд (websocket.go::sendEngineError) шле клієнту.
// -----------------------------------------------------------------------------
export interface ServerErrorPacket {
  /** Завжди "error". */
  status: "error";
  /**
   * Транспортна категорія: "action_rejected", "join_failed", "invalid_json",
   * "missing_payload" тощо. Корисна для логіки ретраїв і дев-логів.
   */
  code: string;
  /**
   * Стабільний код доменної помилки (engine.ErrorCode). Може бути відсутнім
   * для суто транспортних помилок (некоректний JSON). Якщо відсутній —
   * фронт показує стандартне повідомлення Internal.
   */
  error_code?: EngineErrorCode | string;
  /** Англомовний dev-message — НЕ для UI, лише для девтулзів і логів. */
  message?: string;
  /** Опціональний контекст з GameError.Details. */
  details?: Record<string, unknown>;
  /** Echo request_id, який клієнт надіслав у вихідній дії. */
  request_id?: string;
}

// -----------------------------------------------------------------------------
// isEngineErrorCode — type-guard: чи це відомий бекенд-код?
// -----------------------------------------------------------------------------
const KNOWN_CODES = new Set<string>(Object.values(EngineErrorCodes));

export function isEngineErrorCode(value: unknown): value is EngineErrorCode {
  return typeof value === "string" && KNOWN_CODES.has(value);
}
