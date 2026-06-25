<template>
  <Transition name="scale">
    <div
      v-if="data && isValidReveal"
      class="fixed inset-0 bg-black/80 backdrop-blur-md flex items-center justify-center p-2 sm:p-4 z-50 h-app"
    >
      <!-- Картки Барон/Священник Reveal використовують глобальний
           --card-primary-h, тож виглядають ідентично карткам у руці гравця. -->
      <div
        class="bg-brand-bg border-2 border-amber-500/40 rounded-2xl p-4 sm:p-5 lg:p-6 w-full max-w-xl tall:max-w-2xl shadow-2xl text-center flex flex-col overflow-y-auto max-h-[90dvh] custom-scrollbar"
      >
        <div
          class="text-amber-400 font-bold text-xl mb-6 flex items-center justify-center gap-2 font-mono uppercase tracking-wide"
        >
          {{ isBaron ? $t("reveal.duelBaron") : $t("reveal.priestEffect") }}
        </div>

        <div
          class="flex justify-center gap-2 sm:gap-4 lg:gap-6 flex-wrap my-auto items-stretch"
        >
          <div
            v-for="(card, index) in displayCards"
            :key="index"
            class="flex flex-col gap-2 sm:gap-3 items-center flex-shrink-0"
          >
            <!-- Плашка гравця: підсвічуємо золотим, якщо це ефект Барона і гравець переміг -->
            <div
              class="flex items-center gap-2 pl-1.5 pr-3 py-1 rounded-full border transition-all duration-300 w-full justify-center"
              :class="[
                isBaron && winnerId === card.playerData?.id
                  ? 'bg-gradient-to-r from-amber-500/30 via-yellow-500/40 to-amber-500/30 border-yellow-400 shadow-[0_0_15px_color-mix(in_srgb,var(--color-brand-warning)_40%,transparent)] animate-winner-pulse'
                  : 'bg-brand-surface/60 border-brand-border/50',
              ]"
            >
              <div
                class="w-10 h-10 sm:w-16 sm:h-16 rounded-full border bg-brand-bg-dark/60 overflow-hidden flex-shrink-0"
                :class="
                  isBaron && winnerId === card.playerData?.id
                    ? 'border-yellow-400 scale-105'
                    : 'border-slate-600'
                "
              >
                <img
                  :src="
                    getAvatarUrl(
                      card.playerData?.avatar_seed ||
                        card.playerData?.username ||
                        card.playerData?.id
                    )
                  "
                  :alt="$t('common.avatarAlt')"
                  class="w-full h-full object-cover rounded-full"
                />
              </div>
              <span
                class="text-xs font-bold tracking-wide font-mono truncate max-w-[72px] sm:max-w-[120px]"
                :class="
                  isBaron && winnerId === card.playerData?.id
                    ? 'text-yellow-300 drop-shadow-[0_1px_2px_rgba(0,0,0,0.8)]'
                    : 'text-brand-text-subtle'
                "
              >
                {{ card.label }}
              </span>
            </div>

            <div
              class="card-primary rounded-xl border-2 shadow-2xl transition-all duration-300 select-none bg-cover bg-center relative overflow-hidden group hover:scale-105 flex flex-col justify-end"
              :class="[
                card.info.color,
                card.info.border || 'border-white/10',
                isBaron && winnerId === card.playerData?.id
                  ? 'ring-2 ring-yellow-400 ring-offset-2 ring-offset-brand-bg'
                  : '',
              ]"
              :style="
                card.info.image ? { backgroundImage: `url(${card.info.image})` } : {}
              "
              :data-tooltip="`${card.info.name} (${card.id}) - ${card.info.desc}`"
            >
              <span
                class="text-[12px] font-bold font-mono text-center block bg-brand-bg-dark/80 p-1 z-10 relative text-amber-300 uppercase tracking-wider rounded-b-xl border-t border-white/5 w-full"
              >
                {{ card.info.name }}
              </span>
            </div>
          </div>
        </div>

        <button
          @click="$emit('close')"
          type="button"
          class="w-full mt-8 py-3 bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-600 hover:to-amber-700 text-slate-950 font-black rounded-xl text-sm transition-all shadow-lg shadow-amber-500/10 active:scale-95 cursor-pointer transform font-mono uppercase tracking-wider"
        >
          {{ $t("reveal.close") }}
        </button>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { getCardInfoHelper } from "../constants/cards.js";
import { getAvatarUrl } from "../utils/avatar.js";

const { t } = useI18n();

const emit = defineEmits(["close"]);
const props = defineProps({
  data: { type: Object, default: null },
  players: { type: [Object, Array], default: () => ({}) },
});

const isBaron = computed(
  () =>
    props.data?.eventType === "ROUND_COMPARED" || props.data?.type === "ROUND_COMPARED"
);

const isValidReveal = computed(() => {
  if (!props.data) return false;
  if (isBaron.value) return true;
  // Перевіряємо обидва можливі ключі: card або cardType для зворотної сумісності
  return props.data.card !== undefined || props.data.cardType !== undefined;
});

// Обчислюємо ID переможця на основі сили карток
const winnerId = computed(() => {
  if (!isBaron.value || !props.data) return null;

  const pId = props.data.playerId || props.data.player_id;
  const tId = props.data.targetId || props.data.target_id;

  const pCard = Number(
    props.data.playerCard !== undefined ? props.data.playerCard : props.data.player_card
  );
  const tCard = Number(
    props.data.targetCard !== undefined ? props.data.targetCard : props.data.target_card
  );

  if (pCard > tCard) return pId;
  if (tCard > pCard) return tId;
  return null; // Нічия (однакові карти)
});

const getPlayerData = (id) => {
  if (!id || !props.players) return null;
  if (typeof props.players === "object" && props.players[id]) {
    return props.players[id];
  }
  if (Array.isArray(props.players)) {
    return props.players.find((p) => p && String(p.id) === String(id)) || null;
  }
  return null;
};

const getCardInfo = (id) => {
  return (
    getCardInfoHelper(id, t) || {
      name: t("cards.unknown"),
      desc: t("cards.serverHandled"),
      color: "bg-brand-surface",
      emoji: "❓",
      image: "",
    }
  );
};

const displayCards = computed(() => {
  if (!isValidReveal.value) return [];
  if (isBaron.value) {
    const pData = getPlayerData(props.data.playerId || props.data.player_id);
    const tData = getPlayerData(props.data.targetId || props.data.target_id);
    const pCard =
      props.data.playerCard !== undefined
        ? props.data.playerCard
        : props.data.player_card;
    const tCard =
      props.data.targetCard !== undefined
        ? props.data.targetCard
        : props.data.target_card;

    return [
      {
        id: pCard,
        label: pData?.username || t("common.opponent"),
        playerData: pData,
        info: getCardInfo(pCard),
      },
      {
        id: tCard,
        label: tData?.username || t("common.opponent"),
        playerData: tData,
        info: getCardInfo(tCard),
      },
    ];
  }

  const targetId = props.data.targetId || props.data.target_id;
  const cardId = props.data.card !== undefined ? props.data.card : props.data.cardType;
  const tData = getPlayerData(targetId);

  return [
    {
      id: cardId,
      label: tData?.username || t("common.opponent"),
      playerData: tData,
      info: getCardInfo(cardId),
    },
  ];
});

</script>

<style scoped>
.scale-enter-from,
.scale-leave-to {
  opacity: 0;
}
.scale-enter-from .bg-brand-bg,
.scale-leave-to .bg-brand-bg {
  transform: scale(0.9) translateY(10px);
}
.scale-enter-active,
.scale-leave-active {
  transition: opacity 0.3s ease;
}
.scale-enter-active .bg-brand-bg,
.scale-leave-active .bg-brand-bg {
  transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1), opacity 0.3s ease;
}
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: rgba(15, 23, 42, 0.6);
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(245, 158, 11, 0.3);
  border-radius: 2px;
}

/* М'яка золотиста пульсація для переможця */
@keyframes winnerPulse {
  0%,
  100% {
    box-shadow: 0 0 12px rgba(234, 179, 8, 0.4);
  }
  50% {
    box-shadow: 0 0 20px rgba(234, 179, 8, 0.7);
  }
}
.animate-winner-pulse {
  animation: winnerPulse 2s infinite ease-in-out;
}
</style>
