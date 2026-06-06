<template>
  <Transition name="fade">
    <div
      v-if="isOpen"
      class="fixed inset-0 bg-black/80 backdrop-blur-md flex items-center justify-center p-4 z-50"
    >
      <div
        class="bg-slate-900 border-2 border-amber-500/40 rounded-2xl p-6 w-full max-w-2xl shadow-2xl"
      >
        <div class="text-center mb-6">
          <h3
            class="text-amber-400 font-black text-xl tracking-wide uppercase flex items-center justify-center gap-2"
          >
            Ефект Канцлера
          </h3>
          <p class="text-sm text-slate-300 mt-2 font-medium">
            <span v-if="chosenKeepIndex === null" class="text-cyan-400">
              Крок 1: Оберіть 1 карту, яку хочете ЗАЛИШИТИ у себе в руці
            </span>
            <span v-else class="text-orange-400">
              Крок 2: Оберіть послідовність карт для відправки на дно колоди
            </span>
          </p>
        </div>

        <div class="flex justify-center gap-4 min-h-[280px] items-center my-4 flex-wrap">
          <div
            v-for="(cardType, index) in cards"
            :key="index"
            @click="handleCardClick(index)"
            class="group w-44 h-64 rounded-xl flex flex-col justify-between shadow-xl transition-all duration-300 transform select-none cursor-pointer border relative bg-cover bg-center overflow-hidden"
            :class="[getCardColor(cardType), getCardSelectionClass(index)]"
            :style="
              getCardImage(cardType)
                ? { backgroundImage: `url(${getCardImage(cardType)})` }
                : {}
            "
            :data-tooltip="`${getCardName(cardType)}(${getCardValue(
              cardType
            )}) — ${getCardDesc(cardType)}`"
          >
            <div
              v-if="chosenKeepIndex === index"
              class="absolute inset-0 bg-emerald-500/20 rounded-xl border-2 border-emerald-400 flex items-center justify-center z-20 backdrop-blur-[1px]"
            >
              <span
                class="bg-emerald-500 text-slate-950 font-black text-xs px-2 py-1 rounded-md uppercase tracking-wider shadow"
              >
                В Руку
              </span>
            </div>
            <div
              v-if="getBottomOrderIndex(index) !== -1"
              class="absolute inset-0 bg-orange-500/30 rounded-xl border-2 border-orange-400 flex items-center justify-center z-20 backdrop-blur-[1px]"
            >
              <span
                class="bg-orange-500 text-white font-black text-sm w-8 h-8 rounded-full flex items-center justify-center shadow-lg"
              >
                #{{ getBottomOrderIndex(index) + 1 }}</span
              >
            </div>
          </div>
        </div>

        <div
          v-if="bottomSelection.length > 0"
          class="bg-slate-950/60 rounded-xl p-3 mb-4 border border-slate-800 text-center text-xxs text-slate-400"
        >
          Порядок карт на дно:
          <span class="text-orange-400 font-bold">{{
            bottomSelection.map((idx) => getCardName(cards[idx])).join(" ➔ ")
          }}</span>
        </div>

        <div class="flex gap-3 mt-6">
          <button
            @click="resetSelection"
            type="button"
            :disabled="chosenKeepIndex === null"
            class="flex-1 py-2.5 bg-slate-800 hover:bg-slate-700 disabled:opacity-40 disabled:cursor-not-allowed text-slate-300 font-semibold rounded-xl text-sm transition-all"
          >
            Скинути вибір
          </button>
          <button
            @click="handleSubmit"
            type="button"
            :disabled="!isReadyToSubmit"
            class="flex-1 py-2.5 bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-600 hover:to-amber-700 disabled:from-slate-700 disabled:to-slate-700 disabled:opacity-50 disabled:text-slate-500 disabled:cursor-not-allowed text-slate-950 font-black rounded-xl text-sm transition-all shadow-lg shadow-amber-500/10"
          >
            Підтвердити хід
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
  cards: { type: Array, required: true },
});

const emit = defineEmits(["submit"]);

const chosenKeepIndex = ref(null);
const bottomSelection = ref([]);

const getCardDesc = (type) => CARD_INFO[type]?.desc || "Опис відсутній";
const getCardColor = (type) => CARD_INFO[type]?.color || "bg-slate-700";
const getCardName = (type) => CARD_INFO[type]?.name || "Невідома карта";
const getCardValue = (type) =>
  CARD_INFO[type]?.value !== undefined ? CARD_INFO[type].value : "?";
const getCardImage = (type) => CARD_INFO[type]?.image || "";

const isReadyToSubmit = computed(() => {
  return (
    chosenKeepIndex.value !== null &&
    bottomSelection.value.length === props.cards.length - 1
  );
});

const getBottomOrderIndex = (index) => {
  return bottomSelection.value.indexOf(index);
};

const getCardSelectionClass = (index) => {
  if (chosenKeepIndex.value === index)
    return "ring-4 ring-emerald-400 border-transparent scale-95";
  if (getBottomOrderIndex(index) !== -1)
    return "ring-4 ring-orange-500 border-transparent scale-95 opacity-90";
  return "border-white/10 hover:-translate-y-2 hover:scale-105 active:scale-95 ring-2 ring-amber-500/20";
};

const handleCardClick = (index) => {
  if (chosenKeepIndex.value === null) {
    chosenKeepIndex.value = index;
    return;
  }
  if (chosenKeepIndex.value === index) return;

  const orderIdx = getBottomOrderIndex(index);
  if (orderIdx !== -1) {
    bottomSelection.value.splice(orderIdx, 1);
  } else {
    if (bottomSelection.value.length < props.cards.length - 1) {
      bottomSelection.value.push(index);
    }
  }
};

const resetSelection = () => {
  chosenKeepIndex.value = null;
  bottomSelection.value = [];
};

watch(
  () => props.isOpen,
  (newVal) => {
    if (newVal === true) {
      resetSelection();
    }
  }
);

const handleSubmit = () => {
  if (!isReadyToSubmit.value) return;
  const remainingCards = bottomSelection.value.map((idx) => props.cards[idx]);
  emit("submit", {
    keepHandIndex: chosenKeepIndex.value,
    bottomOrder: remainingCards,
  });
};
</script>

<style scoped>
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
</style>
