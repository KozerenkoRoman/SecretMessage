<template>
  <div
    class="flex-1 min-h-0 flex items-center justify-center p-2 sm:p-3 bg-brand-bg-dark/40 rounded-2xl border border-slate-800/60 w-full overflow-hidden relative"
  >
    <div
      class="flex gap-3 sm:gap-4 lg:gap-6 items-center w-full mx-auto justify-between h-full"
    >
      <div class="flex flex-col items-center justify-center gap-1 flex-shrink-0">
        <div
          class="relative card-secondary bg-gradient-to-br from-yellow-800 to-indigo-900 rounded-xl border-2 border-amber-500 shadow-lg shadow-amber-950/20 flex flex-col items-center justify-center bg-cover bg-center overflow-hidden p-[2px]"
          :style="{ backgroundImage: `url(${deckBackImage})` }"
        >
          <div
            class="absolute inset-0 bg-brand-bg-dark/50 flex flex-col items-center justify-center p-1"
          >
            <span
              class="text-[12px] font-bold uppercase text-amber-200 tracking-wider bg-brand-bg-dark/80 px-2 py-1 rounded backdrop-blur-[1px]"
              >{{ $t("board.deck") }}</span
            >
            <span
              class="text-2xl font-black text-white leading-none mt-1.5 bg-brand-bg-dark/70 px-2.5 py-1 rounded font-mono shadow-md border border-white/5"
              >{{ gameState?.deck ? gameState.deck.length : 0 }}</span
            >
          </div>
        </div>
        <div
          class="text-xs sm:text-sm tall:text-base font-bold text-slate-400 font-mono mt-0.5"
        >
          {{ $t("board.discardPile") }}{{ globalDiscardPile.length }}
        </div>
      </div>
      <div class="h-full max-h-[80%] w-[1px] bg-brand-surface flex-shrink-0"></div>
      <div class="flex-scroll flex flex-col justify-start gap-1 pl-1 h-full w-full">
        <div
          class="flex justify-between items-center border-b border-slate-800 pb-1 flex-shrink-0"
        >
          <span
            class="text-[10px] uppercase text-amber-400 font-bold font-mono tracking-wider"
            >{{ $t("board.table") }}</span
          >
        </div>
        <div
          class="discard-grid custom-scrollbar w-full flex-1 min-h-0 px-1 pt-2 sm:pt-4 pb-1"
        >
          <div
            v-for="entry in globalDiscardPile"
            :key="`discard-${entry.seq}`"
            class="game-card game-card-discard card-secondary p-0 border border-brand-border/50 overflow-hidden bg-brand-bg-dark"
            :class="getCardColor(entry.type)"
          >
            <img
              v-if="getCardImage(entry.type)"
              :src="getCardImage(entry.type)"
              :alt="getCardName(entry.type)"
              class="w-full h-full pointer-events-none rounded-[0.4rem]"
            />
            <div
              v-else
              class="w-full h-full flex items-center justify-center text-[10px] text-center p-1"
            >
              {{ getCardName(entry.type) }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <Transition name="slide-banner">
      <div
        v-if="latestTurnAlert"
        class="absolute inset-0 flex justify-center items-center z-30 pointer-events-none bg-brand-bg-dark/40 backdrop-blur-sm rounded-2xl"
      >
        <div
          class="bg-brand-bg-dark border-2 border-amber-500 rounded-2xl shadow-[0_0_25px_rgba(245,158,11,0.5)] px-3 py-6 flex flex-col items-center gap-3 w-52 pointer-events-auto animate-pulse-subtle"
        >
          <div
            class="text-center font-bold text-sm tracking-wide text-white truncate max-w-full"
          >
            {{ $t("common.player") || "Гравець" }}:
            <span class="text-amber-400">{{ latestTurnAlert.playerName }}</span>
          </div>

          <div
            class="game-card card-primary w-32 h-48 bg-cover bg-center flex flex-col justify-between overflow-hidden"
            :class="getCardColor(latestTurnAlert.cardType)"
            :style="
              getCardImage(latestTurnAlert.cardType)
                ? { backgroundImage: `url(${getCardImage(latestTurnAlert.cardType)})` }
                : {}
            "
          >
            <span
              class="text-[12px] font-bold font-mono text-center block bg-brand-bg-dark/80 p-1 mt-auto z-10 relative text-amber-300 uppercase tracking-wider mx-[-0.5rem] mb-[-0.5rem] rounded-b-xl border-t border-white/5"
            >
              {{ getCardName(latestTurnAlert.cardType) }}
            </span>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
defineProps({
  gameState: Object,
  globalDiscardPile: Array,
  deckBackImage: String,
  getCardColor: Function,
  getCardName: Function,
  getCardImage: Function,
  getCardValue: Function,
  latestTurnAlert: Object,
});
</script>

<style scoped>
.slide-banner-enter-from {
  opacity: 0;
  transform: scale(0.85);
}
.slide-banner-leave-to {
  opacity: 0;
  transform: scale(0.95);
}
.slide-banner-enter-active,
.slide-banner-leave-active {
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes pulseSubtle {
  0%,
  100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.01);
  }
}
.animate-pulse-subtle {
  animation: pulseSubtle 2s infinite ease-in-out;
}
</style>
