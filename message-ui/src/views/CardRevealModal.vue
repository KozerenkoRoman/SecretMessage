<template>
  <Transition name="scale">
    <div
      v-if="data && isValidReveal"
      class="fixed inset-0 bg-black/80 backdrop-blur-md flex items-center justify-center p-4 z-50"
    >
      <div
        class="bg-slate-900 border-2 border-amber-500/40 rounded-2xl p-6 w-full max-w-xl shadow-2xl text-center flex flex-col"
      >
        <div
          class="text-amber-400 font-bold text-xl mb-6 flex items-center justify-center gap-2 font-mono uppercase tracking-wide"
        >
          {{ isBaron ? $t("reveal.duelBaron") : $t("reveal.priestEffect") }}
        </div>

        <div class="flex justify-center gap-6 flex-wrap my-auto items-stretch">
          <div
            v-for="(card, index) in displayCards"
            :key="index"
            class="flex flex-col gap-3 items-center flex-1 max-w-[190px]"
          >
            <div
              class="flex items-center gap-2 bg-slate-800/60 pl-1.5 pr-3 py-1 rounded-full border border-slate-700/50 w-full justify-center"
            >
              <div
                class="w-16 h-16 rounded-full border border-slate-600 bg-slate-950/60 overflow-hidden flex-shrink-0"
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
                class="text-xs text-slate-300 font-bold tracking-wide font-mono truncate max-w-[120px]"
              >
                {{ card.label }}
              </span>
            </div>

            <div
              class="w-44 h-64 rounded-xl border-2 shadow-2xl transition-all duration-300 select-none bg-cover bg-center relative overflow-hidden group hover:scale-105 flex flex-col justify-end"
              :class="[card.info.color, card.info.border || 'border-white/10']"
              :style="
                card.info.image ? { backgroundImage: `url(${card.info.image})` } : {}
              "
              :data-tooltip="`${card.info.name} (${card.id}) — ${card.info.desc}`"
            >
              <span
                class="text-[12px] font-bold font-mono text-center block bg-slate-950/80 p-1 z-10 relative text-amber-300 uppercase tracking-wider rounded-b-xl border-t border-white/5 w-full"
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

const isBaron = computed(() => props.data?.eventType === "ROUND_COMPARED");
const isValidReveal = computed(() => {
  if (!props.data) return false;
  return isBaron.value || props.data.cardType !== undefined;
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
      color: "bg-slate-800",
      emoji: "❓",
      image: "",
    }
  );
};

const displayCards = computed(() => {
  if (!isValidReveal.value) return [];
  if (isBaron.value) {
    const pData = getPlayerData(props.data.playerId);
    const tData = getPlayerData(props.data.targetId);
    return [
      {
        id: props.data.playerCard,
        label: pData?.username || t("common.opponent"),
        playerData: pData,
        info: getCardInfo(props.data.playerCard),
      },
      {
        id: props.data.targetCard,
        label: tData?.username || t("common.opponent"),
        playerData: tData,
        info: getCardInfo(props.data.targetCard),
      },
    ];
  }

  const tData = getPlayerData(props.data.targetId);
  return [
    {
      id: props.data.cardType,
      label: tData?.username || t("common.opponent"),
      playerData: tData,
      info: getCardInfo(props.data.cardType),
    },
  ];
});

</script>

<style scoped>
.scale-enter-from,
.scale-leave-to {
  opacity: 0;
}
.scale-enter-from .bg-slate-900,
.scale-leave-to .bg-slate-900 {
  transform: scale(0.9) translateY(10px);
}
.scale-enter-active,
.scale-leave-active {
  transition: opacity 0.3s ease;
}
.scale-enter-active .bg-slate-900,
.scale-leave-active .bg-slate-900 {
  transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1), opacity 0.3s ease;
}
</style>
