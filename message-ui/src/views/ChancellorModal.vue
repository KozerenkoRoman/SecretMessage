<template>
  <Transition name="fade">
    <div
      v-if="isOpen"
      class="fixed inset-0 bg-black/80 backdrop-blur-md flex items-center justify-center p-2 sm:p-4 z-50 h-app"
    >
      <!-- Використовуємо глобальний --card-primary-h: 3 картки Канцлера
           будуть точно того самого розміру, що й картки в руці на дошці.
           Якщо не вміщаються - модалка скролиться. -->
      <div
        class="bg-brand-bg border-2 border-amber-500/40 rounded-2xl p-4 sm:p-5 lg:p-6 w-full max-w-2xl tall:max-w-3xl shadow-2xl overflow-y-auto max-h-[90dvh] custom-scrollbar"
      >
        <div class="text-center mb-6">
          <h3
            class="text-amber-400 font-black text-xl tracking-wide uppercase flex items-center justify-center gap-2"
          >
            {{ $t("chancellor.title") }}
          </h3>
          <p class="text-sm text-brand-text-subtle mt-2 font-medium">
            <span v-if="chosenKeepIndex === null" class="text-cyan-400">
              {{ $t("chancellor.step1") }}
            </span>
            <span v-else class="text-orange-400">
              {{ $t("chancellor.step2") }}
            </span>
          </p>
        </div>

        <!-- Контейнер карт: висота прив'язана до --card-primary-h, тому
             автоматично адаптується. Не потрібно ставити min-height вручну. -->
        <div
          class="flex justify-center gap-2 sm:gap-3 lg:gap-4 items-center my-4 flex-wrap"
        >
          <div
            v-for="(cardType, index) in cards"
            :key="index"
            @click="handleCardClick(index)"
            class="group card-primary rounded-xl flex flex-col justify-between shadow-xl transition-all duration-300 transform select-none cursor-pointer border relative bg-cover bg-center overflow-hidden"
            :class="[getCardColor(cardType), getCardSelectionClass(index)]"
            :style="
              getCardImage(cardType)
                ? { backgroundImage: `url(${getCardImage(cardType)})` }
                : {}
            "
            :data-tooltip="`${getCardName(cardType)}(${getCardValue(
              cardType
            )}) - ${getCardDesc(cardType)}`"
          >
            <div
              v-if="chosenKeepIndex === index"
              class="absolute inset-0 bg-emerald-500/20 rounded-xl border-2 border-emerald-400 flex items-center justify-center z-20 backdrop-blur-[1px]"
            >
              <span
                class="bg-emerald-500 text-slate-950 font-black text-xs px-2 py-1 rounded-md uppercase tracking-wider shadow"
              >
                {{ $t("chancellor.keepBadge") }}
              </span>
            </div>

            <div
              v-if="getBottomOrderIndex(index) !== -1"
              class="absolute inset-0 bg-brand-card-orange/30 rounded-xl border-2 border-orange-400 flex items-center justify-center z-20 backdrop-blur-[1px]"
            >
              <span
                class="bg-brand-card-orange text-white font-black text-sm w-8 h-8 rounded-full flex items-center justify-center shadow-lg"
              >
                #{{ getBottomOrderIndex(index) + 1 }}
              </span>
            </div>

            <span
              class="text-[12px] font-bold font-mono text-center block bg-brand-bg-dark/80 p-1 mt-auto z-10 relative text-amber-300 uppercase tracking-wider mx-[-0.5rem] mb-[-0.5rem] rounded-b-xl border-t border-white/5"
            >
              {{ getCardName(cardType) }}
            </span>
          </div>
        </div>

        <div
          v-if="bottomSelection.length > 0"
          class="bg-brand-bg-dark/60 rounded-xl p-3 mb-4 border border-slate-800 text-center text-xxs text-slate-400"
        >
          {{ $t("chancellor.bottomOrderTitle") }}
          <span class="text-orange-400 font-bold">{{
            bottomSelection.map((idx) => getCardName(cards[idx])).join(" ➔ ")
          }}</span>
        </div>

        <div class="flex gap-3 mt-6">
          <button
            @click="resetSelection"
            type="button"
            :disabled="chosenKeepIndex === null"
            class="flex-1 py-2.5 bg-brand-surface hover:bg-brand-surface-dim disabled:opacity-40 disabled:cursor-not-allowed text-brand-text-subtle font-semibold rounded-xl text-sm transition-all"
          >
            {{ $t("chancellor.reset") }}
          </button>
          <button
            @click="handleSubmit"
            type="button"
            :disabled="!isReadyToSubmit"
            class="flex-1 py-2.5 bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-600 hover:to-amber-700 disabled:from-slate-700 disabled:to-slate-700 disabled:opacity-50 disabled:text-slate-500 disabled:cursor-not-allowed text-slate-950 font-black rounded-xl text-sm transition-all shadow-lg shadow-amber-500/10"
          >
            {{ $t("chancellor.submit") }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { ref, computed, watch } from "vue";
import { useI18n } from "vue-i18n";
import { getCardInfoHelper } from "../constants/cards";

const { t } = useI18n();

const props = defineProps({
  isOpen: { type: Boolean, required: true },
  cards: { type: Array, required: true },
});

const emit = defineEmits(["submit"]);

const chosenKeepIndex = ref(null);
const bottomSelection = ref([]);

const getCardDesc = (type) => getCardInfoHelper(type, t)?.desc || t("cards.noDescription");
const getCardColor = (type) => getCardInfoHelper(type)?.color || "bg-brand-surface-dim";
const getCardName = (type) => getCardInfoHelper(type, t)?.name || t("cards.unknown");
const getCardValue = (type) => {
  const val = getCardInfoHelper(type)?.value;
  return val !== undefined ? val : "?";
};
const getCardImage = (type) => getCardInfoHelper(type)?.image || "";

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
