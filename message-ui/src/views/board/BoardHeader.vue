<template>
  <header
    class="w-full flex flex-wrap sm:flex-nowrap justify-between items-center bg-brand-surface/80 backdrop-blur px-2 py-2 sm:px-3 rounded-xl border border-brand-border flex-shrink-0 gap-2 sm:gap-4 z-10 min-h-[44px] short:min-h-[40px] tall:min-h-[56px]"
  >
    <div class="flex items-center gap-4">
      <button @click="$emit('leave')" type="button" class="btn-danger">
        {{ $t("board.leave") }}
      </button>
      <div class="h-6 w-[1px] bg-brand-surface-dim"></div>

      <div
        class="flex items-center gap-2 bg-brand-bg/60 pl-2 pr-1.5 py-0.5 rounded-lg border border-brand-border/40"
      >
        <span
          class="text-xs font-medium text-brand-text-subtle hidden sm:inline max-w-[80px] truncate"
          :title="myPlayer?.username"
        >
          {{ myPlayer?.username || $t("common.guest") }}
        </span>
        <div
          class="w-11 h-11 rounded-full bg-gradient-to-tr from-amber-500 to-yellow-400 p-[1px] shadow-inner flex-shrink-0"
        >
          <img
            :src="getAvatarUrl(myPlayer?.avatar_seed)"
            :alt="$t('common.avatarAlt')"
            class="w-full h-full object-cover rounded-full bg-brand-surface"
          />
        </div>
      </div>

      <div>
        <h2 class="text-sm font-bold text-yellow-400 font-mono leading-none mb-0.5">
          {{ $t("board.room") }}{{ roomID }}
        </h2>
        <p class="text-xs text-slate-400">
          {{ $t("board.time") }}
          <span class="text-amber-400 font-mono font-bold text-sm"
            >{{ secondsLeft }}{{ $t("board.timeUnit") }}</span
          >
        </p>
      </div>
    </div>

    <div class="flex flex-wrap items-center justify-end gap-2 sm:gap-4 w-full sm:w-auto">
      <div class="flex flex-wrap items-center justify-end gap-2">
        <template
          v-if="
            !gameState?.is_started &&
            (!gameState?.turn_order || gameState.turn_order.length === 0)
          "
        >
          <div class="flex flex-wrap items-center justify-end gap-2">
            <button
              v-if="canStartGame || playersCount < 4"
              @click="$emit('add-bot')"
              type="button"
              class="px-3 py-2 bg-slate-700 hover:bg-slate-600 active:bg-slate-800 text-amber-400 font-bold rounded-xl text-xs border border-brand-border transition-all flex items-center gap-1 disabled:opacity-50"
              :disabled="isBotSubmitting"
            >
              <span class="text-sm leading-none">{{ $t("board.addBot") }}</span>
            </button>
            <button
              v-if="canStartGame"
              @click="$emit('start-game')"
              type="button"
              class="btn-primary disabled:opacity-50 disabled:cursor-not-allowed transition-all whitespace-nowrap"
              :disabled="isStarting"
            >
              {{ $t("board.start") }}
            </button>
            <div
              v-else
              class="status-pill bg-brand-info/20 text-blue-400 border-blue-500/40 animate-pulse"
            >
              {{ $t("board.waiting") }}
            </div>
          </div>
        </template>
        <template v-else>
          <div
            v-if="showChancellorPanel"
            class="status-pill bg-gradient-to-r from-amber-500/20 to-orange-500/20 text-amber-400 border-yellow-500/40 animate-pulse"
          >
            {{ $t("board.chancellorBadge") }}
          </div>
          <div
            v-else-if="isMyTurn"
            class="status-pill bg-emerald-500/20 text-emerald-400 border-emerald-500/40 animate-pulse"
          >
            {{ $t("board.yourTurn") }}
          </div>
          <div
            v-else
            class="status-pill bg-brand-surface-dim text-brand-text-subtle border-transparent"
          >
            {{ $t("board.currentTurn") }}{{ currentTurnPlayerName }}
          </div>
        </template>
      </div>
      <div class="h-6 w-[1px] bg-brand-surface-dim"></div>
    </div>
  </header>
</template>

<script setup>
import { getAvatarUrl } from "../../utils/avatar";

defineProps({
  roomID: String,
  gameState: Object,
  myPlayer: Object,
  secondsLeft: Number,
  canStartGame: Boolean,
  playersCount: Number,
  isBotSubmitting: Boolean,
  isStarting: [String, Boolean],
  showChancellorPanel: Boolean,
  isMyTurn: Boolean,
  currentTurnPlayerName: String,
});

defineEmits(["leave", "add-bot", "start-game"]);
</script>
