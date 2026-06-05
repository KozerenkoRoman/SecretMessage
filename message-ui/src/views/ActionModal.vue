<template>
  <Transition name="fade">
    <div
      v-if="isOpen"
      class="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4 z-50"
    >
      <div
        class="bg-slate-800 border border-slate-700 rounded-2xl p-6 w-full max-w-5xl shadow-2xl overflow-y-auto max-h-[95vh] custom-scrollbar"
      >
        <h3 class="text-lg font-bold text-cyan-400 mb-2">
          Розіграш карти: {{ cardInfo?.name || cardType }}
        </h3>
        <p class="text-xs text-slate-400 italic mb-4">{{ cardInfo?.desc }}</p>

        <div v-if="requiresTargetSelection" class="mb-4">
          <div v-if="availableTargets.length > 0">
            <label
              class="block text-xs uppercase text-slate-400 font-bold mb-2 tracking-wider font-mono"
            >
              Оберіть Ціль:</label
            >
            <div class="grid grid-cols-2 gap-2">
              <button
                v-for="p in availableTargets"
                :key="p.id"
                @click="targetID = p.id"
                type="button"
                class="p-2 rounded-lg border text-sm font-medium transition-all cursor-pointer text-center font-mono"
                :class="[
                  targetID === p.id
                    ? 'bg-cyan-600 border-cyan-400 text-white shadow-md shadow-cyan-500/20'
                    : 'bg-slate-700 border-slate-600 text-slate-200 hover:bg-slate-600',
                ]"
              >
                {{ p.username }}{{ p.id === myID ? " (Ви)" : "" }}
              </button>
            </div>
          </div>
          <div
            v-else
            class="text-sm text-amber-400 bg-amber-500/10 border border-amber-500/30 p-3 rounded-lg flex flex-col gap-1"
          >
            <span class="font-bold">⚠️ Немає доступних цілей!</span>
            <span>
              Усі інші гравці захищені ефектом Служниці або вибули. Карта буде скинута в
              загальний відбій без застосування ефекту.
            </span>
          </div>
        </div>

        <div
          v-if="!requiresTargetSelection"
          class="mb-4 text-sm text-emerald-400 bg-emerald-500/10 p-3 rounded-lg border border-emerald-500/20"
        >
          ✨ Ця карта застосовується автоматично на вас або скидається в стіл.
        </div>

        <div v-if="isGuardGuessRequired && availableTargets.length > 0" class="mb-5 mt-4">
          <label
            class="block text-xs uppercase text-slate-400 font-bold mb-2 tracking-wider font-mono"
          >
            Вгадайте карту опонента (натисніть для вибору):</label
          >

          <div
            class="flex flex-col gap-4 bg-slate-900/60 p-5 rounded-xl border border-slate-700/50"
          >
            <div class="flex justify-center gap-4">
              <div
                v-for="card in cardRows.row1"
                :key="card.type"
                @click="guessCard = card.type"
                class="w-44 h-64 flex-shrink-0 rounded-xl flex flex-col justify-between text-white shadow-md relative cursor-pointer transition-all duration-200 select-none bg-cover bg-center overflow-hidden"
                :class="[
                  getCardColor(card.type),
                  guessCard === card.type
                    ? 'ring-4 ring-cyan-400 scale-105 z-10 shadow-lg shadow-cyan-500/30 border-transparent'
                    : 'opacity-70 hover:opacity-100 hover:scale-102 border border-white/5',
                ]"
                :style="
                  getCardImage(card.type)
                    ? { backgroundImage: `url(${getCardImage(card.type)})` }
                    : {}
                "
                :data-tooltip="`${card.info.name}(${card.info.value}) — ${
                  card.info.desc || 'Опис відсутній'
                }`"
              >
                <span
                  class="text-[8px] font-bold font-mono text-center block bg-slate-950/80 p-1 mt-auto z-10 relative text-cyan-300 rounded-b uppercase tracking-wider"
                >
                  {{ guessCard === card.type ? "Обрано" : "Обрати" }}
                </span>
              </div>
            </div>

            <div class="flex justify-center gap-4">
              <div
                v-for="card in cardRows.row2"
                :key="card.type"
                @click="guessCard = card.type"
                class="w-44 h-64 flex-shrink-0 rounded-xl flex flex-col justify-between text-white shadow-md relative cursor-pointer transition-all duration-200 select-none bg-cover bg-center overflow-hidden"
                :class="[
                  getCardColor(card.type),
                  guessCard === card.type
                    ? 'ring-4 ring-cyan-400 scale-105 z-10 shadow-lg shadow-cyan-500/30 border-transparent'
                    : 'opacity-70 hover:opacity-100 hover:scale-102 border border-white/5',
                ]"
                :style="
                  getCardImage(card.type)
                    ? { backgroundImage: `url(${getCardImage(card.type)})` }
                    : {}
                "
                :data-tooltip="`${card.info.name}(${card.info.value}) — ${
                  card.info.desc || 'Опис відсутній'
                }`"
              >
                <span
                  class="text-[8px] font-bold font-mono text-center block bg-slate-950/80 p-1 mt-auto z-10 relative text-cyan-300 rounded-b uppercase tracking-wider"
                >
                  {{ guessCard === card.type ? "Обрано" : "Обрати" }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <div class="flex gap-3 justify-end mt-6 border-t border-slate-700/40 pt-4">
          <button
            @click="handleCancel"
            type="button"
            class="px-4 py-2 bg-slate-700 hover:bg-slate-600 rounded-lg text-sm font-medium cursor-pointer transition-colors font-mono"
          >
            Скасувати
          </button>
          <button
            @click="handleSubmit"
            :disabled="!isValid"
            type="button"
            class="px-5 py-2 bg-gradient-to-r from-cyan-500 to-indigo-600 hover:from-cyan-600 hover:to-indigo-700 disabled:opacity-40 disabled:cursor-not-allowed rounded-lg text-sm font-bold shadow-lg cursor-pointer transition-all font-mono uppercase tracking-wider"
          >
            Підтвердити Хід
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { ref, computed, watch } from "vue";
import { CARD_INFO } from "../constants/cards";

const props = defineProps({
  isOpen: { type: Boolean, required: true },
  cardType: { type: [String, Number], required: true },
  handIndex: { type: Number, required: true },
  players: { type: Array, required: true },
  myID: { type: String, required: true },
});

const emit = defineEmits(["close", "submit"]);

const targetID = ref("");
const guessCard = ref("");

const cardInfo = computed(() => CARD_INFO[props.cardType]);

const getCardColor = (type) => {
  return CARD_INFO[type]?.color || "bg-slate-700";
};
const getCardImage = (type) => {
  return CARD_INFO[type]?.image || "";
};

const requiresTargetSelection = computed(() => {
  return cardInfo.value?.targetType && cardInfo.value.targetType !== "SELF";
});
const isGuardGuessRequired = computed(() => {
  return !!cardInfo.value?.requiresGuess;
});

const availableTargets = computed(() => {
  if (!props.players || !cardInfo.value) return [];
  const rule = cardInfo.value.targetType;
  return props.players.filter((p) => {
    if (!p || p.is_out) return false;
    if (rule === "ANY" && p.id === props.myID) return true;
    if (p.id === props.myID) return false;
    return !p.is_protected;
  });
});

const filteredCardOptions = computed(() => {
  const uniqueCards = {};
  Object.keys(CARD_INFO).forEach((key) => {
    if (isNaN(key)) {
      if (key === "GUARD") return;
      uniqueCards[key] = CARD_INFO[key];
    }
  });
  return uniqueCards;
});

// НОВИЙ СOMPUTED ВЛАСТИВОСТІ ДЛЯ РОЗДІЛЕННЯ НА ДВА РЯДИ (5 ТА 4 КАРТИ)
const cardRows = computed(() => {
  const cardsArray = Object.keys(filteredCardOptions.value).map((key) => ({
    type: key,
    info: filteredCardOptions.value[key],
  }));

  return {
    row1: cardsArray.slice(0, 5), // Перші 5 карт
    row2: cardsArray.slice(5, 9), // Наступні 4 карти
  };
});

const isValid = computed(() => {
  if (!requiresTargetSelection.value) return true;
  if (availableTargets.value.length === 0) return true;
  if (!targetID.value) return false;
  if (isGuardGuessRequired.value && !guessCard.value) return false;
  return true;
});

watch(
  () => props.isOpen,
  (newVal) => {
    if (newVal) {
      targetID.value = "";
      guessCard.value = "";
      if (!requiresTargetSelection.value) {
        targetID.value = props.myID;
      } else if (
        availableTargets.value.length === 1 &&
        cardInfo.value?.targetType !== "ANY"
      ) {
        targetID.value = availableTargets.value[0].id;
      }
    }
  }
);

const handleCancel = () => {
  emit("close");
};

const handleSubmit = () => {
  if (!isValid.value) return;
  let guessCardId = null;
  if (isGuardGuessRequired.value && guessCard.value) {
    const foundKey = Object.keys(CARD_INFO).find(
      (key) => !isNaN(key) && CARD_INFO[key].type === guessCard.value
    );
    guessCardId = foundKey !== undefined ? Number(foundKey) : null;
  }
  const finalTarget = availableTargets.value.length === 0 ? null : targetID.value;
  emit("submit", {
    handIndex: props.handIndex,
    targetID: finalTarget,
    guessCardId: guessCardId,
  });
};
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: rgba(15, 23, 42, 0.3);
  border-radius: 999px;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(51, 65, 85, 0.8);
  border-radius: 999px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(148, 163, 184, 0.8);
}
</style>
