<template>
  <Transition name="fade">
    <div
      v-if="hasError"
      class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-[60]"
    >
      <div
        class="bg-slate-800 border border-rose-500/40 rounded-2xl p-6 w-full max-w-sm shadow-2xl text-center transform scale-100 transition-all"
      >
        <div
          class="w-12 h-12 bg-rose-500/10 text-rose-400 rounded-full flex items-center justify-center mx-auto mb-4 text-2xl border border-rose-500/20"
        >
          ⚠️
        </div>
        <h3 class="text-lg font-bold text-rose-400 mb-1">{{ $t("gameError.title") }}</h3>
        <p class="text-xs text-slate-400 italic mb-4">{{ $t("gameError.server") }}</p>

        <div
          class="text-sm text-slate-200 bg-slate-900/60 p-3 rounded-xl border border-slate-700/50 text-center break-words leading-relaxed"
        >
          {{ translatedMessage }}
        </div>

        <p
          v-if="errorCode"
          class="text-[10px] text-slate-500 font-mono mt-2 uppercase tracking-wider"
        >
          {{ $t("gameError.codeLabel") }} {{ errorCode }}
        </p>

        <div class="mt-5 space-y-2">
          <button
            @click="$emit('close')"
            type="button"
            class="w-full py-2.5 bg-gradient-to-r from-slate-700 to-slate-700 hover:from-rose-600 hover:to-rose-700 text-white font-bold rounded-xl text-sm transition-all border border-slate-600 hover:border-rose-500 cursor-pointer shadow-lg active:scale-95"
          >
            {{ $t("gameError.ok") }}
          </button>

          <button
            v-if="isSoloError"
            @click="handleLeave"
            type="button"
            class="w-full py-2.5 bg-rose-600 hover:bg-rose-700 text-white font-bold rounded-xl text-sm transition-all shadow-lg shadow-rose-500/10 cursor-pointer uppercase font-mono tracking-wider active:scale-95"
          >
            {{ $t("gameError.leave") }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { computed } from "vue";
import { useGameStore } from "../stores/gameStore";
import { useRouter } from "vue-router";
import { EngineErrorCodes } from "../types/errors";

const props = defineProps({
  message: {
    type: [String, Object, null],
    default: null,
  },
});

const emit = defineEmits(["close"]);

const gameStore = useGameStore();
const router = useRouter();

const ERROR_TRANSLATIONS = Object.freeze({
  failed_to_next_round:
    "Неможливо розпочати наступний раунд: усі опоненти залишили кімнату.",
  [EngineErrorCodes.Internal]:
    "Сталася внутрішня помилка сервера. Спробуйте повторити дію через декілька секунд.",

  [EngineErrorCodes.InvalidPhase]: "Цю дію не можна виконати на поточній фазі гри.",
  [EngineErrorCodes.InvalidState]:
    "Стан гри неконсистентний. Перезавантажте сторінку та підключіться повторно.",
  [EngineErrorCodes.EmptyTurnOrder]: "У кімнаті не визначено порядок ходів.",
  [EngineErrorCodes.CurrentTurnOutOfRange]:
    "Внутрішня помилка кімнати: некоректний поточний хід.",
  [EngineErrorCodes.NoPlayers]: "У кімнаті немає активних гравців.",
  [EngineErrorCodes.CardEffectNotImplemented]:
    "Ефект цієї карти ще не реалізовано на сервері.",

  [EngineErrorCodes.PlayerNotFound]: "Вас не знайдено серед учасників цієї кімнати.",
  [EngineErrorCodes.PlayerAlreadyOut]:
    "Ви вже вибули з поточного раунду — дочекайтеся наступного.",
  [EngineErrorCodes.PlayerProtected]: "Ви знаходитесь під захистом Служниці.",
  [EngineErrorCodes.PlayerHasNoCards]: "У вас немає карт у руці для виконання цієї дії.",

  [EngineErrorCodes.OutOfTurn]: "Зараз хід іншого гравця. Зачекайте своєї черги.",
  [EngineErrorCodes.InvalidHandIndex]: "Невірний індекс карти у вашій руці.",
  [EngineErrorCodes.CannotTargetSelf]: "Цією картою не можна цілитися в самого себе.",
  [EngineErrorCodes.MustPlayCountess]:
    "Згідно з правилами, ви зобов’язані зіграти Графиню, якщо в руці є Принц або Король!",

  [EngineErrorCodes.TargetRequired]: "Для цієї карти необхідно обрати гравця-ціль.",
  [EngineErrorCodes.TargetNotFound]: "Обраного гравця не знайдено в кімнаті.",
  [EngineErrorCodes.TargetAlreadyOut]:
    "Обраний гравець уже вибув з раунду — оберіть іншу ціль.",
  [EngineErrorCodes.TargetProtected]:
    "Неможливо застосувати ефект: обраний гравець знаходиться під захистом Служниці.",
  [EngineErrorCodes.GuardCannotGuessGuard]:
    "Вартовий не може вгадувати іншого Вартового.",
  [EngineErrorCodes.GuardGuessRequired]:
    "Вкажіть, яку саме карту ви намагаєтесь вгадати.",
  [EngineErrorCodes.BaronNoCardsToCompare]: "Немає карт для порівняння Бароном.",

  [EngineErrorCodes.ChancellorInvalidBottomOrder]:
    "Неправильний порядок повернення карт у колоду.",
  [EngineErrorCodes.ChancellorWrongPhase]:
    "Канцлер не очікує вибору карти на поточному етапі.",
});

const FALLBACK_GENERIC =
  "Сервер відхилив вашу дію. Перевірте умови ходу та спробуйте ще раз.";
const FALLBACK_EMPTY = "Сталася невідома помилка.";

const errorCode = computed(() => {
  const raw = props.message;
  if (!raw) return "";
  if (typeof raw === "string") {
    return raw.startsWith("ERR_") || raw === "failed_to_next_round" ? raw : "";
  }
  if (typeof raw === "object" && typeof raw.code === "string") {
    return raw.code;
  }
  return "";
});

const rawMessageText = computed(() => {
  const raw = props.message;
  if (!raw) return "";
  if (typeof raw === "string") return raw;
  return raw.message || "";
});

const isSoloError = computed(() => {
  const code = errorCode.value;
  const text = rawMessageText.value;
  return (
    code === "failed_to_next_round" || text.includes("insufficient connected players")
  );
});

const hasError = computed(() => {
  const raw = props.message;
  if (!raw) return false;
  if (typeof raw === "string") return raw.length > 0;
  if (typeof raw === "object") return Boolean(raw.code || raw.message);
  return false;
});

const translatedMessage = computed(() => {
  const code = errorCode.value;
  if (code && ERROR_TRANSLATIONS[code]) {
    return ERROR_TRANSLATIONS[code];
  }

  const text = rawMessageText.value;
  if (text.includes("insufficient connected players")) {
    return "Неможливо розпочати наступний раунд: недостатньо підключених гравців (ви залишилися самі).";
  }

  return code ? FALLBACK_GENERIC : text || FALLBACK_GENERIC;
});

const handleLeave = () => {
  emit("close");
  gameStore.leaveCurrentRoom();
  router.push("/desktop");
};
</script>

<style scoped>
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
</style>
