<template>
  <div
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-fade-in"
  >
    <!-- Головний адаптивний контейнер -->
    <div
      class="bg-brand-bg border border-slate-800 rounded-2xl max-w-3xl w-full max-h-[85vh] flex flex-col overflow-hidden shadow-2xl text-white"
    >
      <!-- Хедер -->
      <div
        class="p-4 border-b border-slate-800 flex items-center justify-between bg-brand-bg-dark"
      >
        <div class="flex items-center gap-2">
          <h3 class="text-lg font-bold text-amber-500 font-mono">
            {{ $t(RULES_I18N_KEYS.modalTitle) }}
          </h3>
        </div>
        <button
          @click="$emit('close')"
          class="text-slate-400 hover:text-white text-xl font-bold font-mono transition-colors p-1"
        >
          &times;
        </button>
      </div>

      <!-- Основний контент зі скролом -->
      <div
        class="p-6 overflow-y-auto custom-scrollbar flex-1 space-y-6 text-sm text-slate-300 leading-relaxed"
      >
        <!-- Блок 1: Загальні правила та Хід -->
        <section
          class="space-y-3 bg-brand-bg-dark/40 p-4 rounded-xl border border-slate-800/50"
        >
          <h4
            class="text-amber-400 font-mono font-bold uppercase tracking-wider border-b border-slate-800 pb-1.5 flex items-center gap-2"
          >
            {{ $t(RULES_I18N_KEYS.generalHeader) }}
          </h4>

          <ul class="list-disc list-inside space-y-2 pl-1 text-sm">
            <li v-for="(ruleKey, index) in RULES_I18N_KEYS.generalRules" :key="index">
              {{ $t(ruleKey) }}
            </li>
          </ul>
        </section>

        <!-- Блок 2: Опис дій усіх карт (від 0 до 9) -->
        <section class="space-y-4">
          <h4
            class="text-amber-400 font-mono font-bold uppercase tracking-wider border-b border-slate-800 pb-1.5 flex items-center gap-2"
          >
            {{ $t(RULES_I18N_KEYS.cardsHeader) }}
          </h4>

          <div class="grid gap-3">
            <div
              v-for="cardId in sortedCardIds"
              :key="cardId"
              class="flex gap-4 p-3 bg-brand-bg-dark/60 rounded-xl border border-slate-800 hover:border-slate-700/50 transition-all items-center"
            >
              <!-- Колонка карти ліворуч -->
              <div
                class="w-16 h-24 sm:w-20 sm:h-28 bg-brand-bg-dark border border-amber-500/20 rounded-lg flex-shrink-0 flex items-center justify-center overflow-hidden shadow-md relative"
                :class="getCardColor(cardId)"
              >
                <img
                  v-if="getCardImage(cardId)"
                  :src="getCardImage(cardId)"
                  :alt="getCardName(cardId)"
                  class="w-full h-full object-cover pointer-events-none"
                />
                <div v-else class="text-center p-1 text-sm font-bold text-slate-600">
                  {{ getCardName(cardId) }}
                </div>

                <!-- Великий ігровий емодзі по центру, якщо картинка не завантажиться -->
                <div v-if="!getCardImage(cardId)" class="absolute text-xl opacity-30">
                  {{ getCardEmoji(cardId) }}
                </div>

                <!-- Сила карти в кутку -->
                <div
                  class="absolute top-1 left-1 bg-black/80 text-amber-400 font-mono text-[11px] font-bold w-5 h-5 flex items-center justify-center rounded-full border border-amber-500/40 shadow"
                >
                  {{ getCardValue(cardId) }}
                </div>
              </div>

              <!-- Колонка опису праворуч -->
              <div class="flex-1 space-y-1">
                <div class="flex items-baseline gap-2">
                  <span class="text-lg">{{ getCardEmoji(cardId) }}</span>
                  <h5 class="font-bold text-amber-500 font-mono text-base leading-none">
                    {{ getCardName(cardId) }}
                  </h5>

                  <span class="text-sm text-slate-500 font-mono">
                    ({{ $t(RULES_I18N_KEYS.strength) }}: {{ getCardValue(cardId) }})
                  </span>
                </div>
                <p class="text-sm text-slate-300 leading-normal">
                  {{ getCardDesc(cardId) }}
                </p>
              </div>
            </div>
          </div>
        </section>

        <!-- Блок 3: Важливі нюанси -->
        <section
          class="space-y-3 bg-brand-bg-dark/40 p-4 rounded-xl border border-slate-800/50"
        >
          <h4
            class="text-amber-400 font-mono font-bold uppercase tracking-wider border-b border-amber-500/10 pb-1 flex items-center gap-2 text-sm"
          >
            {{ $t(RULES_I18N_KEYS.nuancesHeader) }}
          </h4>
          <ul class="list-disc list-inside space-y-2 pl-1 text-sm text-slate-400">
            <li v-for="(nuanceKey, index) in RULES_I18N_KEYS.nuances" :key="index">
              {{ $t(nuanceKey) }}
            </li>
          </ul>
        </section>

        <!-- Блок 4: Стратегія -->
        <section
          class="space-y-3 bg-brand-bg-dark/40 p-4 rounded-xl border border-slate-800/50"
        >
          <h4
            class="text-amber-400 font-mono font-bold uppercase tracking-wider border-b border-amber-500/10 pb-1 flex items-center gap-2 text-sm"
          >
            {{ $t(RULES_I18N_KEYS.strategyHeader) }}
          </h4>
          <ul class="list-disc list-inside space-y-2 pl-1 text-sm text-slate-400">
            <li v-for="(strategyKey, index) in RULES_I18N_KEYS.strategies" :key="index">
              {{ $t(strategyKey) }}
            </li>
          </ul>
        </section>
      </div>

      <!-- Кнопка закриття у футері -->
      <div class="p-4 border-t border-slate-800 bg-brand-bg-dark flex justify-end">
        <button
          @click="$emit('close')"
          class="btn-ghost text-xs uppercase tracking-wider py-2 px-5"
        >
          {{ $t("reveal.close") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import {
  getCardInfoHelper,
  CARD_INFO_NUMBERS,
  RULES_I18N_KEYS,
} from "../constants/cards";

defineEmits(["close"]);
const { t } = useI18n();

// Динамічно збираємо ключі (0-9) та сортуємо відповідно до ігрового валуативного рейтингу
const sortedCardIds = computed(() => {
  return Object.keys(CARD_INFO_NUMBERS)
    .map(Number)
    .sort((a, b) => CARD_INFO_NUMBERS[a].value - CARD_INFO_NUMBERS[b].value);
});

// Проксі-методи безпечного вилучення даних з вашого хелпера
const getCardName = (id) => getCardInfoHelper(id, t)?.name || t("cards.unknown");
const getCardDesc = (id) => getCardInfoHelper(id, t)?.desc || t("cards.noDescription");
const getCardImage = (id) => getCardInfoHelper(id)?.image || "";
const getCardColor = (id) => getCardInfoHelper(id)?.color || "bg-brand-surface-dim";
const getCardEmoji = (id) => getCardInfoHelper(id)?.emoji || "🃏";
const getCardValue = (id) => {
  const val = getCardInfoHelper(id)?.value;
  return val !== undefined ? val : id;
};
</script>
```eof ```javascript:src/constants/cards.js /* ===== FILE: src/constants/cards.js ===== */
export const CARD_I18N_KEYS = { SPY: { name: "cards.SPY.name", desc: "cards.SPY.desc" },
GUARD: { name: "cards.GUARD.name", desc: "cards.GUARD.desc" }, PRIEST: { name:
"cards.PRIEST.name", desc: "cards.PRIEST.desc" }, BARON: { name: "cards.BARON.name", desc:
"cards.BARON.desc" }, HANDMAID: { name: "cards.HANDMAID.name", desc: "cards.HANDMAID.desc"
}, PRINCE: { name: "cards.PRINCE.name", desc: "cards.PRINCE.desc" }, CHANCELLOR: { name:
"cards.CHANCELLOR.name", desc: "cards.CHANCELLOR.desc" }, KING: { name: "cards.KING.name",
desc: "cards.KING.desc" }, COUNTESS: { name: "cards.COUNTESS.name", desc:
"cards.COUNTESS.desc" }, PRINCESS: { name: "cards.PRINCESS.name", desc:
"cards.PRINCESS.desc" } }; const getCardImg = (fileName) => { return new
URL(`../assets/cards/${fileName}`, import.meta.url).href; }; // ВИПРАВЛЕНО: у 7 (Король)
встановлено value: 7, а у 8 (Графиня) встановлено value: 8 export const CARD_INFO_NUMBERS
= { 0: { nameKey: CARD_I18N_KEYS.SPY.name, descKey: CARD_I18N_KEYS.SPY.desc, type: "SPY",
value: 0, color: "bg-brand-card-gray", emoji: "🦅", targetType: "SELF", image:
getCardImg("00_Spy.png") }, 1: { nameKey: CARD_I18N_KEYS.GUARD.name, descKey:
CARD_I18N_KEYS.GUARD.desc, type: "GUARD", value: 1, color: "bg-brand-card-red", emoji:
"⚔️", targetType: "OPPONENT", requiresGuess: true, image: getCardImg("01_Guard.png") }, 2:
{ nameKey: CARD_I18N_KEYS.PRIEST.name, descKey: CARD_I18N_KEYS.PRIEST.desc, type:
"PRIEST", value: 2, color: "bg-brand-card-blue", emoji: "⛪", targetType: "OPPONENT",
image: getCardImg("02_Priest.png") }, 3: { nameKey: CARD_I18N_KEYS.BARON.name, descKey:
CARD_I18N_KEYS.BARON.desc, type: "BARON", value: 3, color: "bg-brand-card-green", emoji:
"🎭", targetType: "OPPONENT", image: getCardImg("03_Baron.png") }, 4: { nameKey:
CARD_I18N_KEYS.HANDMAID.name, descKey: CARD_I18N_KEYS.HANDMAID.desc, type: "HANDMAID",
value: 4, color: "bg-brand-card-yellow", emoji: "🛡️", targetType: "SELF", image:
getCardImg("04_Handmaid.png") }, 5: { nameKey: CARD_I18N_KEYS.PRINCE.name, descKey:
CARD_I18N_KEYS.PRINCE.desc, type: "PRINCE", value: 5, color: "bg-brand-card-cyan", emoji:
"👑", targetType: "ANY", image: getCardImg("05_Prince.png") }, 6: { nameKey:
CARD_I18N_KEYS.CHANCELLOR.name, descKey: CARD_I18N_KEYS.CHANCELLOR.desc, type:
"CHANCELLOR", value: 6, color: "bg-brand-card-purple", emoji: "📜", targetType: "SELF",
image: getCardImg("06_Minister.png") }, 7: { nameKey: CARD_I18N_KEYS.KING.name, descKey:
CARD_I18N_KEYS.KING.desc, type: "KING", value: 7, color: "bg-brand-card-amber", emoji:
"⚜️", targetType: "OPPONENT", image: getCardImg("07_King.png") }, 8: { nameKey:
CARD_I18N_KEYS.COUNTESS.name, descKey: CARD_I18N_KEYS.COUNTESS.desc, type: "COUNTESS",
value: 8, color: "bg-brand-card-orange", emoji: "💃", targetType: "SELF", image:
getCardImg("08_Countess.png") }, 9: { nameKey: CARD_I18N_KEYS.PRINCESS.name, descKey:
CARD_I18N_KEYS.PRINCESS.desc, type: "PRINCESS", value: 9, color: "bg-brand-card-pink",
emoji: "👸", targetType: "SELF", image: getCardImg("09_Princess.png") } }; export const
CARD_INFO_NAMES = { "SPY": { nameKey: CARD_I18N_KEYS.SPY.name, descKey:
CARD_I18N_KEYS.SPY.desc, type: "SPY", value: 0, color: "bg-brand-card-gray", emoji: "🦅",
targetType: "SELF", image: getCardImg("00_Spy.png") }, "GUARD": { nameKey:
CARD_I18N_KEYS.GUARD.name, descKey: CARD_I18N_KEYS.GUARD.desc, type: "GUARD", value: 1,
color: "bg-brand-card-red", emoji: "⚔️", targetType: "OPPONENT", requiresGuess: true,
image: getCardImg("01_Guard.png") }, "PRIEST": { nameKey: CARD_I18N_KEYS.PRIEST.name,
descKey: CARD_I18N_KEYS.PRIEST.desc, type: "PRIEST", value: 2, color:
"bg-brand-card-blue", emoji: "⛪", targetType: "OPPONENT", image:
getCardImg("02_Priest.png") }, "BARON": { nameKey: CARD_I18N_KEYS.BARON.name, descKey:
CARD_I18N_KEYS.BARON.desc, type: "BARON", value: 3, color: "bg-brand-card-green", emoji:
"🎭", targetType: "OPPONENT", image: getCardImg("03_Baron.png") }, "HANDMAID": { nameKey:
CARD_I18N_KEYS.HANDMAID.name, descKey: CARD_I18N_KEYS.HANDMAID.desc, type: "HANDMAID",
value: 4, color: "bg-brand-card-yellow", emoji: "🛡️", targetType: "SELF", image:
getCardImg("04_Handmaid.png") }, "PRINCE": { nameKey: CARD_I18N_KEYS.PRINCE.name, descKey:
CARD_I18N_KEYS.PRINCE.desc, type: "PRINCE", value: 5, color: "bg-brand-card-cyan", emoji:
"👑", targetType: "ANY", image: getCardImg("05_Prince.png") }, "CHANCELLOR": { nameKey:
CARD_I18N_KEYS.CHANCELLOR.name, descKey: CARD_I18N_KEYS.CHANCELLOR.desc, type:
"CHANCELLOR", value: 6, color: "bg-brand-card-purple", emoji: "📜", targetType: "SELF",
image: getCardImg("06_Minister.png") }, "KING": { nameKey: CARD_I18N_KEYS.KING.name,
descKey: CARD_I18N_KEYS.KING.desc, type: "KING", value: 7, border: "border-amber-600",
color: "bg-brand-card-amber", emoji: "⚜️", targetType: "OPPONENT", image:
getCardImg("07_King.png") }, "COUNTESS": { nameKey: CARD_I18N_KEYS.COUNTESS.name, descKey:
CARD_I18N_KEYS.COUNTESS.desc, type: "COUNTESS", value: 8, color: "bg-brand-card-orange",
emoji: "💃", targetType: "SELF", image: getCardImg("08_Countess.png") }, "PRINCESS": {
nameKey: CARD_I18N_KEYS.PRINCESS.name, descKey: CARD_I18N_KEYS.PRINCESS.desc, type:
"PRINCESS", value: 9, color: "bg-brand-card-pink", emoji: "👸", targetType: "SELF", image:
getCardImg("09_Princess.png") } }; export const RULES_I18N_KEYS = { modalTitle:
"rules.modalTitle", generalHeader: "rules.generalHeader", cardsHeader:
"rules.cardsHeader", nuancesHeader: "rules.nuancesHeader", strategyHeader:
"rules.strategyHeader", strength: "rules.strength", generalRules: [ "rules.turnStep1",
"rules.turnStep2", "rules.turnStep3", "rules.turnStep4", "rules.turnStep5", ], nuances: [
"rules.nuance1", "rules.nuance2", "rules.nuance3" ], strategies: [ "rules.strategy1",
"rules.strategy2", "rules.strategy3", "rules.strategy4", "rules.strategy5",
"rules.strategy6" ] }; const resolveCard = (raw, t) => { if (!raw) return raw; if (typeof
t !== "function") return raw; return { ...raw, name: t(raw.nameKey), desc: t(raw.descKey),
}; }; export const getCardInfoHelper = (key, t) => { if (key === undefined || key ===
null) return null; const raw = (!isNaN(key) && key !== "") ?
CARD_INFO_NUMBERS[Number(key)] : CARD_INFO_NAMES[String(key)]; return resolveCard(raw, t);
}; ```eof ```typescript:src/i18n/uk.ts /* ===== ДОДАТИ / ОНОВИТИ В src/i18n/uk.ts ===== */
export const uk = { // ... ваш існуючий файл локалізації ... desktop: { // переконайтеся,
що цей ключ є тут rulesButton: "Правила гри", // ... інші ключі }, rules: { modalTitle:
"Правила гри «Secret Message»", generalHeader: "Загальні правила та Хід гри", cardsHeader:
"Реєстр карт та їхні ефекти", nuancesHeader: "Важливі нюанси", strategyHeader: "Коротко
про стратегію", strength: "Значення", turnStep1: "Мета раунду: залишитися останнім у грі
або мати карту з найбільшим значенням на руці, коли закінчиться колода.", turnStep2:
"Кожна карта має вагу (число) і текстовий ефект (дію на гравця).", turnStep3: "На початку
ходу візьміть 1 карту з колоди (у вас в руці стає 2 карти).", turnStep4: "З двох карт на
руці виберіть та зіграйте одну горілиць перед собою.", turnStep5: "Негайно виконайте її
текстовий ефект.", nuance1: "Усі зіграні та скинуті карти лежать відкрито перед гравцями
протягом усього раунду.", nuance2: "Якщо наприкінці гри колода закінчилася і кілька
гравців мають однакову старшу карту, порівнюється сума значень усіх їхніх раніше зіграних
карт. Якщо знову нічия — перемагають усі нічийні гравці.", nuance3: "Для повної перемоги в
партії потрібно набрати певну кількість жетонів прихильності (залежить від кількості
учасників у кімнаті).", strategy1: "Вартовий (1) і Барон (3) фокусуються на вибиванні
суперників", strategy2: "Священник (2) збирає інформацію", strategy3: "Служниця (4)
гарантує безпеку на коло", strategy4: "Принц (5) здатний змусити скинути Принцесу",
strategy5: "Канцлер (6) ідеально контролює руку й колоду", strategy6: "Шпигун (0) дозволяє
заробити додатковий жетон прихильності, навіть якщо ви програли сам раунд", }, };
