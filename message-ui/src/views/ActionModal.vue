<template>
  <Transition name="fade">
    <div
      v-if="isOpen"
      class="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4 z-50"
    >
      <div
        class="bg-brand-surface border border-brand-border rounded-2xl p-6 w-full shadow-2xl overflow-y-auto max-h-[95vh] custom-scrollbar transition-all duration-300"
        :class="[
          isGuardGuessRequired && availableTargets.length > 0 ? 'max-w-5xl' : 'max-w-xl',
        ]"
      >
        <h3 class="text-lg font-bold text-amber-400 mb-2">
          {{ $t("action.title") }} {{ cardInfo?.name || cardType }}
        </h3>

        <div v-if="requiresTargetSelection" class="mb-4">
          <div v-if="availableTargets.length > 0">
            <div
              class="flex flex-wrap justify-center gap-6 bg-brand-bg/40 p-4 rounded-xl border border-brand-border/30"
            >
              <button
                v-for="p in availableTargets"
                :key="p.id"
                @click="targetID = p.id"
                type="button"
                class="flex flex-col items-center gap-2 group focus:outline-none cursor-pointer"
              >
                <div
                  class="w-16 h-16 rounded-full border-2 p-1 overflow-hidden transition-all duration-200"
                  :class="[
                    targetID === p.id
                      ? 'border-amber-400 bg-amber-950/40 scale-105 shadow-lg shadow-amber-500/20'
                      : 'border-slate-600 bg-brand-bg-dark/60 group-hover:border-slate-400 group-hover:scale-102',
                  ]"
                >
                  <img
                    :src="getAvatarUrl(p.avatar_seed || p.username || p.id)"
                    :alt="$t('common.avatarAlt')"
                    class="w-full h-full object-cover rounded-full"
                  />
                </div>
                <span
                  class="text-xs font-semibold font-mono tracking-wide max-w-[100px] truncate text-center transition-colors"
                  :class="[
                    targetID === p.id
                      ? 'text-amber-400 font-bold'
                      : 'text-brand-text-subtle group-hover:text-white',
                  ]"
                >
                  {{ p.username }}{{ p.id === myID ? $t("action.youSuffix") : "" }}
                </span>
              </button>
            </div>
          </div>
          <div
            v-else
            class="text-sm text-amber-400 bg-brand-accent/10 border border-amber-500/30 p-3 rounded-lg flex flex-col gap-1"
          >
            <span class="font-bold">{{ $t("action.noTargets") }}</span>
            <span>
              {{ $t("action.noTargetsHint") }}
            </span>
          </div>
        </div>

        <div
          v-if="!requiresTargetSelection"
          class="mb-4 text-sm text-amber-400 bg-brand-accent/5 p-3 rounded-lg border border-amber-500/20"
        >
          {{ $t("action.autoApply") }}
        </div>

        <div v-if="isGuardGuessRequired && availableTargets.length > 0" class="mb-5 mt-4">
          <label
            class="block text-xs uppercase text-slate-400 font-bold mb-2 tracking-wider font-mono"
          >
            {{ $t("action.guardGuess") }}
          </label>

          <div
            class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4 bg-brand-bg/60 p-6 rounded-xl border border-brand-border/50 justify-items-center"
          >
            <div
              v-for="card in allCards"
              :key="card.type"
              @click="guessCard = card.type"
              class="w-44 h-64 aspect-[2/3] rounded-xl flex flex-col justify-between text-white shadow-md relative cursor-pointer transition-all duration-200 select-none bg-cover bg-center overflow-hidden"
              :class="[
                getCardColor(card.type),
                guessCard === card.type
                  ? 'ring-4 ring-amber-400 scale-105 z-10 shadow-lg shadow-amber-500/30 border-transparent'
                  : 'opacity-70 hover:opacity-100 hover:scale-102 border border-white/5',
              ]"
              :style="
                getCardImage(card.type)
                  ? { backgroundImage: `url(${getCardImage(card.type)})` }
                  : {}
              "
              :data-tooltip="`${card.info.name}(${card.info.value}) — ${
                card.info.desc || $t('cards.noDescription')
              }`"
            >
              <span
                class="text-[12px] font-bold font-mono text-center block bg-brand-bg-dark/80 p-1 mt-auto z-10 relative text-amber-300 rounded-b uppercase tracking-wider"
              >
                {{ guessCard === card.type ? $t("action.chosen") : card.info.name }}
              </span>
            </div>
          </div>
        </div>

        <div class="flex gap-3 justify-end mt-6 border-t border-brand-border/40 pt-4">
          <button
            @click="handleCancel"
            type="button"
            class="px-5 py-2.5 bg-brand-surface hover:bg-brand-surface-dim text-amber-500 hover:text-amber-400 font-bold border border-brand-border rounded-xl shadow-md cursor-pointer transition-all active:scale-95 font-mono uppercase text-xs tracking-wider"
          >
            {{ $t("action.cancel") }}
          </button>
          <button
            @click="handleSubmit"
            :disabled="!isValid"
            type="button"
            class="px-5 py-2.5 bg-gradient-to-r from-amber-500 to-orange-600 hover:from-amber-600 hover:to-orange-700 disabled:opacity-40 disabled:cursor-not-allowed text-slate-950 font-black rounded-xl shadow-lg shadow-amber-950/20 cursor-pointer transition-all active:scale-95 font-mono uppercase text-xs tracking-wider"
          >
            {{ $t("action.submit") }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { ref, computed, watch } from "vue";
import { useI18n } from "vue-i18n";
import {
  CARD_INFO_NUMBERS,
  CARD_INFO_NAMES,
  getCardInfoHelper,
} from "../constants/cards";
import { getAvatarUrl } from "../utils/avatar";

const { t } = useI18n();

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

const cardInfo = computed(() => getCardInfoHelper(props.cardType, t));

const getCardColor = (type) => {
  return getCardInfoHelper(type)?.color || "bg-brand-surface-dim";
};

const getCardImage = (type) => {
  return getCardInfoHelper(type)?.image || "";
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
  Object.keys(CARD_INFO_NAMES).forEach((key) => {
    if (key === "GUARD") return;
    const raw = CARD_INFO_NAMES[key];
    uniqueCards[key] = {
      ...raw,
      name: t(raw.nameKey),
      desc: t(raw.descKey),
    };
  });
  return uniqueCards;
});

const allCards = computed(() => {
  return Object.keys(filteredCardOptions.value).map((key) => ({
    type: key,
    info: filteredCardOptions.value[key],
  }));
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
    const foundKey = Object.keys(CARD_INFO_NUMBERS).find(
      (key) => CARD_INFO_NUMBERS[key].type === guessCard.value
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
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.25s ease;
}
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
  background: rgba(148, 164, 184, 0.8);
}
</style>
