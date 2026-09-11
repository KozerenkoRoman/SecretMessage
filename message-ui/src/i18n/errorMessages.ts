// =============================================================================
// i18n/errorMessages.ts
//
// ЛОКАЛІЗАЦІЯ КОДІВ ПОМИЛОК БЕКЕНДУ.
//
// Цей файл - єдиний шлях, через який стабільні коди (EngineErrorCode)
// перетворюються на текст, видимий користувачу. Бекенд НЕ знає про мови;
// він шле "ERR_PLAYER_ALREADY_OUT", а ми тут вирішуємо, як це виглядатиме.
//
// Архітектура:
//   • Locale - string-union доступних мов ("uk" | "en" | ...).
//   • ErrorMessages<Locale> - мапа Locale → (EngineErrorCode → string).
//   • translateErrorCode(code, locale) - публічний помічник.
//   • useErrorTranslator() - composable для Vue (реактивний переклад).
//
// Як додати нову мову:
//   1. Додати літерал у тип Locale.
//   2. Додати об'єкт у MESSAGES з усіма ключами.
//   3. TypeScript змусить заповнити всі коди (Record<EngineErrorCode, string>).
// =============================================================================

import { ref, computed, type Ref, type ComputedRef } from "vue";
import {
  EngineErrorCodes,
  isEngineErrorCode,
  type EngineErrorCode,
  type ServerErrorPacket,
} from "../types/errors";

export type Locale = "uk" | "en";
export const DEFAULT_LOCALE: Locale = "uk";

// -----------------------------------------------------------------------------
// Тип мапи перекладів. Record<EngineErrorCode, string> змушує TypeScript
// перевіряти, що для КОЖНОГО коду з бекенду є переклад. Якщо ви забудете
// перекласти новий код - компілятор не пропустить.
// -----------------------------------------------------------------------------
type ErrorDictionary = Record<EngineErrorCode, string>;

const uk: ErrorDictionary = {
  [EngineErrorCodes.BaronNoCardsToCompare]: "Немає карт для порівняння Бароном.",
  [EngineErrorCodes.CannotTargetSelf]: "Цією картою не можна цілити в себе.",
  [EngineErrorCodes.CardEffectNotImplemented]: "Ефект цієї карти ще не реалізовано.",
  [EngineErrorCodes.ChancellorInvalidBottomOrder]: "Невірний порядок карт для повернення в колоду.",
  [EngineErrorCodes.ChancellorWrongPhase]: "Канцлер не очікує вибору в цій фазі.",
  [EngineErrorCodes.CurrentTurnOutOfRange]: "Помилка кімнати: некоректний поточний хід.",
  [EngineErrorCodes.EmptyTurnOrder]: "Помилка кімнати: не визначено порядок ходів.",
  [EngineErrorCodes.GuardCannotGuessGuard]: "Вартовий не може вгадувати Вартового.",
  [EngineErrorCodes.GuardGuessRequired]: "Вкажіть карту, яку ви вгадуєте.",
  [EngineErrorCodes.Internal]: "Сталася внутрішня помилка. Спробуйте ще раз.",
  [EngineErrorCodes.InvalidHandIndex]: "Невірний індекс карти в руці.",
  [EngineErrorCodes.InvalidPhase]: "Ця дія недоступна на поточному етапі гри.",
  [EngineErrorCodes.InvalidState]: "Стан гри неконсистентний. Перезавантажте сторінку.",
  [EngineErrorCodes.MustPlayCountess]: "Ви маєте розіграти Графиню разом з Королем або Принцом.",
  [EngineErrorCodes.NoPlayers]: "У кімнаті немає гравців.",
  [EngineErrorCodes.OutOfTurn]: "Зараз не ваш хід.",
  [EngineErrorCodes.PlayerAlreadyOut]: "Гравець вже вибув з раунду!",
  [EngineErrorCodes.PlayerHasNoCards]: "У гравця немає карт у руці.",
  [EngineErrorCodes.PlayerNotFound]: "Гравця не знайдено в цій кімнаті.",
  [EngineErrorCodes.PlayerProtected]: "Гравець під захистом Покоївки.",
  [EngineErrorCodes.TargetAlreadyOut]: "Цей гравець уже вибув з раунду.",
  [EngineErrorCodes.TargetNotFound]: "Цільового гравця не знайдено.",
  [EngineErrorCodes.TargetProtected]: "Цей гравець захищений Покоївкою - оберіть іншу ціль.",
  [EngineErrorCodes.TargetRequired]: "Оберіть гравця-ціль.",
  [EngineErrorCodes.ErrInvalidAction]: "Невідома дія.",
};

const en: ErrorDictionary = {
  [EngineErrorCodes.BaronNoCardsToCompare]: "There are no cards to compare with the Baron.",
  [EngineErrorCodes.CannotTargetSelf]: "You cannot target yourself with this card.",
  [EngineErrorCodes.CardEffectNotImplemented]: "This card's effect is not implemented yet.",
  [EngineErrorCodes.ChancellorInvalidBottomOrder]: "Invalid order of cards to return to the deck.",
  [EngineErrorCodes.ChancellorWrongPhase]: "The Chancellor is not awaiting a choice in this phase.",
  [EngineErrorCodes.CurrentTurnOutOfRange]: "Room error: current turn index is out of range.",
  [EngineErrorCodes.EmptyTurnOrder]: "Room error: turn order is missing.",
  [EngineErrorCodes.GuardCannotGuessGuard]: "The Guard cannot guess another Guard.",
  [EngineErrorCodes.GuardGuessRequired]: "Please specify which card you are guessing.",
  [EngineErrorCodes.Internal]: "An internal error occurred. Please try again.",
  [EngineErrorCodes.InvalidHandIndex]: "Invalid card index in hand.",
  [EngineErrorCodes.InvalidPhase]: "This action is not allowed in the current game phase.",
  [EngineErrorCodes.InvalidState]: "Game state is inconsistent. Please refresh the page.",
  [EngineErrorCodes.MustPlayCountess]: "You must discard the Countess when holding the King or Prince.",
  [EngineErrorCodes.NoPlayers]: "There are no players in the room.",
  [EngineErrorCodes.OutOfTurn]: "It is not your turn.",
  [EngineErrorCodes.PlayerAlreadyOut]: "Player is already out of the round!",
  [EngineErrorCodes.PlayerHasNoCards]: "Player has no cards in hand.",
  [EngineErrorCodes.PlayerNotFound]: "Player was not found in this room.",
  [EngineErrorCodes.PlayerProtected]: "Player is protected by the Handmaid.",
  [EngineErrorCodes.TargetAlreadyOut]: "This player is already out of the round.",
  [EngineErrorCodes.TargetNotFound]: "Target player was not found.",
  [EngineErrorCodes.TargetProtected]: "This player is protected by the Handmaid - choose another target.",
  [EngineErrorCodes.TargetRequired]: "Please select a target player.",
  [EngineErrorCodes.ErrInvalidAction]: "Unknown action.",
};

// -----------------------------------------------------------------------------
// Реєстр локалей. Ключ - код локалі, значення - словник.
// -----------------------------------------------------------------------------
const MESSAGES: Record<Locale, ErrorDictionary> = { uk, en };

// =============================================================================
// ПУБЛІЧНИЙ API
// =============================================================================

/**
 * translateErrorCode - синхронний переклад коду в текст для UI.
 *
 * • Якщо `code` відомий → повертає переклад для заданої локалі.
 * • Якщо `code` невідомий (нова версія бекенду випередила фронт) →
 *   повертає переклад для ERR_INTERNAL та логує предупередження.
 *
 * НЕ кидає винятки - UI має завжди мати ЩО показати.
 */
export function translateErrorCode(
  code: EngineErrorCode | string | undefined | null,
  locale: Locale = DEFAULT_LOCALE,
): string {
  const dict = MESSAGES[locale] ?? MESSAGES[DEFAULT_LOCALE];

  if (isEngineErrorCode(code)) {
    return dict[code];
  }

  // Невідомий або відсутній код - fallback на Internal.
  if (code) {
    // eslint-disable-next-line no-console
    console.warn(
      `[i18n] Unknown error_code "${code}" - falling back to ERR_INTERNAL`,
    );
  }
  return dict[EngineErrorCodes.Internal];
}

/**
 * translateServerError - зручний хелпер для готового пакета помилки з WebSocket.
 *
 * Використання у gameStore:
 *
 *   socket.onmessage = (event) => {
 *     const packet = JSON.parse(event.data);
 *     if (packet.status === 'error') {
 *       error.value = translateServerError(packet, currentLocale.value);
 *     }
 *   };
 */
export function translateServerError(
  packet: ServerErrorPacket | null | undefined,
  locale: Locale = DEFAULT_LOCALE,
): string {
  if (!packet) return translateErrorCode(EngineErrorCodes.Internal, locale);
  return translateErrorCode(packet.error_code, locale);
}

// -----------------------------------------------------------------------------
// Реактивна локаль (мінімальний "i18n manager" без зовнішніх залежностей).
// Якщо у проекті з'явиться vue-i18n або pinia-store локалі - перенесіть
// `currentLocale` туди і вилучіть цей блок.
// -----------------------------------------------------------------------------
export const currentLocale: Ref<Locale> = ref<Locale>(DEFAULT_LOCALE);

export function setLocale(locale: Locale): void {
  currentLocale.value = locale;
}

/**
 * useErrorTranslator - composable. Повертає реактивну функцію перекладу,
 * яка автоматично перерахується при зміні currentLocale.
 *
 * Використання у компоненті:
 *
 *   <script setup lang="ts">
 *   import { useErrorTranslator } from '@/i18n/errorMessages';
 *   const { t } = useErrorTranslator();
 *   const errMsg = computed(() => t(serverError.value?.error_code));
 *   </script>
 */
export function useErrorTranslator(): {
  t: (code: EngineErrorCode | string | undefined | null) => string;
  locale: Ref<Locale>;
  translatedFor: (
    code: EngineErrorCode | string | undefined | null,
  ) => ComputedRef<string>;
} {
  const t = (code: EngineErrorCode | string | undefined | null): string =>
    translateErrorCode(code, currentLocale.value);

  const translatedFor = (
    code: EngineErrorCode | string | undefined | null,
  ): ComputedRef<string> =>
    computed(() => translateErrorCode(code, currentLocale.value));

  return { t, locale: currentLocale, translatedFor };
}
