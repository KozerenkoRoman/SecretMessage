<template>
  <Transition name="scale">
    <div
      v-if="data && isValidReveal"
      class="fixed inset-0 bg-black/80 backdrop-blur-md flex items-center justify-center p-4 z-50"
    >
      <div
        class="bg-slate-900 border-2 border-amber-500/40 rounded-2xl p-6 w-full max-w-lg shadow-2xl text-center flex flex-col"
      >
        <div
          class="text-amber-400 font-bold text-xl mb-6 flex items-center justify-center gap-2 font-mono uppercase tracking-wide"
        >
          ✨{{ isBaron ? "Дуель Барона" : "Ефект Священника" }}
        </div>
        <div class="flex justify-center gap-6 flex-wrap my-auto items-stretch">
          <div
            v-for="(card, index) in displayCards"
            :key="index"
            class="flex flex-col gap-2 items-center"
          >
            <span
              class="text-xs text-slate-400 uppercase font-bold tracking-wider font-mono bg-slate-800/50 px-2 py-0.5 rounded border border-slate-700/40"
            >
              {{ card.label }}
            </span>
            <div
              class="w-44 h-64 rounded-xl border-2 shadow-2xl transition-all duration-300 select-none bg-cover bg-center relative overflow-hidden group hover:scale-105"
              :class="[card.info.color, card.info.border || 'border-white/10']"
              :style="
                card.info.image ? { backgroundImage: `url(${card.info.image})` } : {}
              "
              :data-tooltip="`${card.info.name} (${card.id}) — ${card.info.desc}`"
            >
              <div
                class="absolute bottom-2 left-1/2 transform -translate-x-1/2 bg-slate-950/80 backdrop-blur-[2px] px-3 py-1 rounded-full border border-white/10 z-10"
              >
                <span
                  class="text-[9px] font-bold uppercase tracking-widest text-slate-300"
                  >Розкрито</span
                >
              </div>
            </div>
          </div>
        </div>
        <button
          @click="$emit('close')"
          type="button"
          class="w-full mt-8 py-3 bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-600 hover:to-amber-700 text-slate-950 font-black rounded-xl text-sm transition-all shadow-lg shadow-amber-500/10 active:scale-95 cursor-pointer transform font-mono uppercase tracking-wider"
        >
          Закрити та продовжити
        </button>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { computed } from "vue";
import { CARD_INFO } from "../constants/cards.js";

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

const getCardInfo = (id) => {
  return (
    CARD_INFO[id] || {
      name: "Невідома карта",
      desc: "Ефект оброблюється сервером",
      color: "bg-slate-800",
      emoji: "❓",
      image: "",
    }
  );
};

const displayCards = computed(() => {
  if (!isValidReveal.value) return [];
  if (isBaron.value) {
    const playerName = getTargetName(props.data.playerId);
    const targetName = getTargetName(props.data.targetId);
    return [
      {
        id: props.data.playerCard,
        label: playerName,
        info: getCardInfo(props.data.playerCard),
      },
      {
        id: props.data.targetCard,
        label: targetName,
        info: getCardInfo(props.data.targetCard),
      },
    ];
  }
  return [
    {
      id: props.data.cardType,
      label: `Карта гравця ${getTargetName(props.data.targetId)}`,
      info: getCardInfo(props.data.cardType),
    },
  ];
});

const getTargetName = (id) => {
  if (!id || !props.players) return "Опонент";
  if (typeof props.players === "object" && props.players[id]) {
    return props.players[id].username;
  }
  if (Array.isArray(props.players)) {
    const found = props.players.find((p) => p && String(p.id) === String(id));
    return found ? found.username : "Опонент";
  }
  return "Опонент";
};
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
