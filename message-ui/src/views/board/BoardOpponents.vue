<template>
  <div
    class="w-full flex justify-center items-stretch gap-2 sm:gap-4 lg:gap-6 flex-wrap square:flex-wrap wide:flex-nowrap flex-shrink-0"
    style="height: calc(var(--card-secondary-h) + 4rem)"
  >
    <template v-for="player in opponents" :key="player.id">
      <div
        :class="[
          isTurnOfPlayer(player.id)
            ? 'border-yellow-400 bg-brand-surface ring-2 ring-yellow-400/30'
            : 'border-brand-border bg-brand-surface/60',
          player.is_out ? 'opacity-40 grayscale-[30%]' : '',
        ]"
        class="flex flex-col items-center px-1.5 sm:px-2 py-1.5 sm:py-2 rounded-xl border transition-all duration-300 h-full justify-start gap-1 shadow-md relative basis-[clamp(140px,18vw,260px)] flex-shrink min-w-0"
      >
        <div
          class="flex items-center justify-between w-full border-b border-brand-border pb-1 flex-shrink-0 min-h-[28px]"
        >
          <span
            class="font-bold text-[11px] tall:text-xs retina:text-sm truncate max-w-[60%] transition-all duration-200 px-1.5 py-0.5 rounded"
            :class="[
              player.is_out
                ? 'bg-rose-900 text-rose-100 font-black'
                : 'text-brand-text-muted',
            ]"
            :title="player.username"
          >
            {{ player.username || $t("common.opponent") }}
            <template v-if="player.is_protected">
              <span
                class="font-bold text-[11px] tall:text-xs truncate max-w-[150px] transition-all duration-200 px-1.5 py-0.5 rounded bg-brand-warning text-slate-950 font-black shadow"
                >{{ $t("board.protection") }}</span
              >
            </template>
            <template v-if="player.is_out"> ({{ $t("status.out") }})</template>
          </span>
          <span
            class="text-[12px] tall:text-xs retina:text-sm bg-brand-surface-dim text-yellow-400 px-1 py-0.5 rounded font-mono flex-shrink-0"
            >★{{ player.score || 0 }}</span
          >
        </div>

        <div
          class="flex justify-center items-center w-full relative px-2 sm:px-3 lg:px-5 py-2 flex-1 min-h-0"
        >
          <div
            class="flex flex-row items-center justify-center relative w-full -space-x-4 flex-shrink-0"
          >
            <div
              v-for="cIdx in getOpponentHandCount(player)"
              :key="`opp-card-${player.id}-${cIdx}`"
              class="card-secondary rounded-lg border border-amber-500 bg-brand-bg-dark p-[1px] shadow-[0_0_6px_rgba(245,158,11,0.4)] overflow-hidden relative"
            >
              <img
                :src="deckBackImage"
                :alt="$t('common.cardBackAlt')"
                class="w-full h-full object-contain rounded-sm select-none"
              />
            </div>

            <div
              v-if="lastPlayedCardsByPlayer[player.id] !== undefined"
              class="flex flex-col items-center justify-center flex-shrink-0 z-20 ml-2"
            >
              <div
                class="card-secondary rounded-lg border border-amber-500 bg-brand-bg-dark p-[1px] shadow-[0_0_6px_rgba(245,158,11,0.4)] overflow-hidden relative transition-transform duration-300"
                :class="getCardColor(lastPlayedCardsByPlayer[player.id])"
                :data-tooltip="getCardName(lastPlayedCardsByPlayer[player.id])"
              >
                <img
                  v-if="getCardImage(lastPlayedCardsByPlayer[player.id])"
                  :src="getCardImage(lastPlayedCardsByPlayer[player.id])"
                  :alt="getCardName(lastPlayedCardsByPlayer[player.id])"
                  class="w-full h-full object-contain rounded-sm pointer-events-none"
                />
                <div
                  v-else
                  class="w-full h-full flex items-center justify-center text-[9px] text-center p-1 font-bold"
                >
                  {{ getCardName(lastPlayedCardsByPlayer[player.id]) }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
defineProps({
  opponents: Array,
  lastPlayedCardsByPlayer: Object,
  deckBackImage: String,
  isTurnOfPlayer: Function,
  getOpponentHandCount: Function,
  getCardColor: Function,
  getCardName: Function,
  getCardImage: Function,
});
</script>
