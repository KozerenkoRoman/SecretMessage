// =============================================================================
// engine/errors.go
//
// I18N-FIRST ERROR MODEL
// -----------------------------------------------------------------------------
// Двигун (engine) НІКОЛИ не повертає сирий англійський (чи будь-який інший)
// текст помилки клієнту. Замість цього він повертає стабільний string-код
// у вигляді ErrorCode (напр., "ERR_PLAYER_ALREADY_OUT").
//
// Чому саме так:
//   1. Бекенд не знає мови користувача — це відповідальність UI (i18n).
//   2. Коди є стабільним публічним API: ми можемо змінювати англійський
//      "developer-facing" опис, але код залишається тим самим, і всі
//      перекладені рядки на фронті продовжують працювати.
//   3. Коди — це enum-подібний контракт: фронт може зробити exhaustive
//      switch і компілятор вкаже на пропущені варіанти.
//
// Архітектура:
//   - ErrorCode — типізований string (узгоджено з constants нижче).
//   - GameError — реалізує `error`, несе в собі Code + dev-повідомлення
//     (для логів/отладки) + Details (опціональний контекст).
//   - NewError / WrapError — конструктори.
//   - CodeOf(err) — публічний хелпер, який витягує ErrorCode з будь-якої
//     помилки. Якщо помилка не GameError — повертає ErrInternal.
//
// Для room/network шару це означає: достатньо викликати engine.CodeOf(err)
// і покласти результат у поле "error_code" відповіді WebSocket.
// =============================================================================

package engine

import (
	"errors"
	"fmt"
)

// ErrorCode — стабільний публічний код помилки.
// Усі значення мають префікс "ERR_" для зручної ідентифікації.
type ErrorCode string

// =============================================================================
// КОДИ ПОМИЛОК
// =============================================================================
//
// УВАГА: Будь-яка зміна значень нижче — це BREAKING CHANGE для фронтенду.
// Дозволено ДОДАВАТИ нові коди; ВИДАЛЯТИ або ПЕРЕЙМЕНОВУВАТИ — ні.
// -----------------------------------------------------------------------------

const (
	// --- Загальні / службові -------------------------------------------------
	ErrInternal ErrorCode = "ERR_INTERNAL" // Невідома/непередбачена внутрішня помилка.

	// --- Валідація стану гри -------------------------------------------------
	ErrInvalidPhase           ErrorCode = "ERR_INVALID_PHASE"             // Дія прийшла у невідповідній фазі (напр., MAIN_ACTION очікувалась).
	ErrInvalidState           ErrorCode = "ERR_INVALID_STATE"             // Загальна неконсистентність стану (turn_order пустий тощо).
	ErrEmptyTurnOrder         ErrorCode = "ERR_EMPTY_TURN_ORDER"          // У стані відсутній порядок ходів.
	ErrCurrentTurnOutOfRange  ErrorCode = "ERR_CURRENT_TURN_OUT_OF_RANGE" // CurrentTurn вийшов за межі TurnOrder.
	ErrNoPlayers              ErrorCode = "ERR_NO_PLAYERS"                // У стані немає гравців.
	ErrCardEffectNotImplemented ErrorCode = "ERR_CARD_EFFECT_NOT_IMPLEMENTED" // Ефект карти не реалізовано (баг).

	// --- Валідація гравця ----------------------------------------------------
	ErrPlayerNotFound    ErrorCode = "ERR_PLAYER_NOT_FOUND"    // Гравця з вказаним ID нема в кімнаті.
	ErrPlayerAlreadyOut  ErrorCode = "ERR_PLAYER_ALREADY_OUT"  // Гравець вже вибув з раунду.
	ErrPlayerProtected   ErrorCode = "ERR_PLAYER_PROTECTED"    // Гравець під захистом Покоївки.
	ErrPlayerHasNoCards  ErrorCode = "ERR_PLAYER_HAS_NO_CARDS" // У гравця нема карт у руці.

	// --- Валідація дії -------------------------------------------------------
	ErrOutOfTurn        ErrorCode = "ERR_OUT_OF_TURN"        // Не черга цього гравця.
	ErrInvalidHandIndex ErrorCode = "ERR_INVALID_HAND_INDEX" // hand_index вийшов за межі руки.
	ErrCannotTargetSelf ErrorCode = "ERR_CANNOT_TARGET_SELF" // Карта не дозволяє цілитись у себе.
	ErrMustPlayCountess ErrorCode = "ERR_MUST_PLAY_COUNTESS" // У руці Графиня + (Король/Принц) — обов'язково Графиня.

	// --- Валідація цілі ------------------------------------------------------
	ErrTargetRequired     ErrorCode = "ERR_TARGET_REQUIRED"     // Карта потребує ціль, але її не передано.
	ErrTargetNotFound     ErrorCode = "ERR_TARGET_NOT_FOUND"    // Ціль з вказаним ID відсутня.
	ErrTargetAlreadyOut   ErrorCode = "ERR_TARGET_ALREADY_OUT"  // Ціль вже вибула з раунду.
	ErrTargetProtected    ErrorCode = "ERR_TARGET_PROTECTED"    // Ціль під захистом Покоївки.

	// --- Карто-специфічні валідації -----------------------------------------
	ErrGuardCannotGuessGuard ErrorCode = "ERR_GUARD_CANNOT_GUESS_GUARD" // Вартовий не може вгадувати Вартового.
	ErrGuardGuessRequired    ErrorCode = "ERR_GUARD_GUESS_REQUIRED"    // У дії Вартового відсутнє поле guess_card.
	ErrBaronNoCardsToCompare ErrorCode = "ERR_BARON_NO_CARDS_TO_COMPARE" // Один з гравців немає карти для Баронового порівняння.

	// --- Chancellor (резолв) -------------------------------------------------
	ErrChancellorInvalidBottomOrder ErrorCode = "ERR_CHANCELLOR_INVALID_BOTTOM_ORDER" // BottomOrder не відповідає набору карт у руці.
	ErrChancellorWrongPhase         ErrorCode = "ERR_CHANCELLOR_WRONG_PHASE"          // ResolveChancellor викликано не у фазі RESOLVE_CHANCELLOR.
)

// =============================================================================
// GameError — типізована доменна помилка двигуна.
// =============================================================================

// GameError реалізує інтерфейс `error`. Поле Message — це developer-facing
// текст для логів/отладки (НЕ для відображення користувачу). Поле Details
// дозволяє приклеїти структурований контекст (наприклад, який саме
// player_id був не знайдений).
type GameError struct {
	Code    ErrorCode      `json:"code"`
	Message string         `json:"message,omitempty"` // Англомовний dev-message — НЕ для UI.
	Details map[string]any `json:"details,omitempty"`
	cause   error          // обгорнута причина (errors.Unwrap)
}

// Error імплементує стандартний інтерфейс error.
// Формат: "ERR_CODE: developer message".
// Цей текст призначений ВИКЛЮЧНО для логів сервера.
func (e *GameError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message == "" {
		return string(e.Code)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap дозволяє використовувати errors.Is / errors.As з обгорнутою причиною.
func (e *GameError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// Is дозволяє порівняння через errors.Is(err, &GameError{Code: ErrPlayerNotFound}).
// Достатньо співпадіння Code.
func (e *GameError) Is(target error) bool {
	if t, ok := target.(*GameError); ok {
		return t.Code == e.Code
	}
	return false
}

// =============================================================================
// КОНСТРУКТОРИ
// =============================================================================

// NewError створює GameError з даним кодом і необов'язковим dev-повідомленням.
// Приклад:
//   return NewError(ErrPlayerNotFound, "player_id=%s", action.PlayerID)
func NewError(code ErrorCode, format string, args ...any) *GameError {
	msg := ""
	if format != "" {
		msg = fmt.Sprintf(format, args...)
	}
	return &GameError{Code: code, Message: msg}
}

// WrapError огортає вже існуючу помилку доменним кодом.
// Корисно, коли ми отримали `error` з нижнього шару (БД, JSON-парсер) і хочемо
// зберегти оригінал, але промаркувати його гарним кодом.
func WrapError(code ErrorCode, cause error, format string, args ...any) *GameError {
	msg := ""
	if format != "" {
		msg = fmt.Sprintf(format, args...)
	} else if cause != nil {
		msg = cause.Error()
	}
	return &GameError{Code: code, Message: msg, cause: cause}
}

// WithDetails додає контекстне поле до помилки (chainable).
//   return NewError(ErrInvalidHandIndex, "").WithDetails("index", idx, "hand_size", n)
func (e *GameError) WithDetails(kv ...any) *GameError {
	if e == nil || len(kv) == 0 {
		return e
	}
	if e.Details == nil {
		e.Details = make(map[string]any, len(kv)/2)
	}
	for i := 0; i+1 < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if !ok {
			continue
		}
		e.Details[key] = kv[i+1]
	}
	return e
}

// =============================================================================
// ПУБЛІЧНІ ХЕЛПЕРИ
// =============================================================================

// CodeOf витягує ErrorCode з будь-якої помилки.
// Якщо err == nil — повертає "" (порожній код).
// Якщо err не є GameError — повертає ErrInternal.
//
// Це ОСНОВНА точка інтеграції для network-шару:
//   if err := engine.Apply(...); err != nil {
//       packet.ErrorCode = engine.CodeOf(err) // → "ERR_PLAYER_NOT_FOUND"
//   }
func CodeOf(err error) ErrorCode {
	if err == nil {
		return ""
	}
	var ge *GameError
	if errors.As(err, &ge) && ge != nil {
		return ge.Code
	}
	return ErrInternal
}

// AsGameError намагається кастувати err до *GameError. Зручно, коли потрібен
// доступ до Details чи Message без втрати типу.
func AsGameError(err error) (*GameError, bool) {
	if err == nil {
		return nil, false
	}
	var ge *GameError
	if errors.As(err, &ge) {
		return ge, true
	}
	return nil, false
}
