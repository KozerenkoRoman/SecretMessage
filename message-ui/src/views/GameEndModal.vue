<template>
  <Transition name="fade">
    <div
      v-if="isOpen"
      class="fixed inset-0 bg-brand-bg-dark/80 backdrop-blur-md flex items-center justify-center p-2 sm:p-4 z-50 font-sans animate-fade-in h-app"
    >
      <div
        class="bg-gradient-to-b from-slate-900 to-slate-950 border rounded-2xl w-full max-w-md tall:max-w-lg p-4 sm:p-5 lg:p-6 shadow-2xl text-center transform scale-100 transition-all duration-300 relative overflow-y-auto overflow-x-hidden border-slate-800 max-h-[90dvh] custom-scrollbar"
        :class="{ 'border-amber-500 shadow-amber-500/10': isAmIWinner }"
      >
        <div
          v-if="isAmIWinner"
          class="absolute -top-12 -left-12 w-32 h-32 bg-brand-accent/10 blur-3xl rounded-full"
        ></div>
        <div
          v-if="isAmIWinner"
          class="absolute -top-12 -right-12 w-32 h-32 bg-brand-warning/10 blur-3xl rounded-full"
        ></div>

        <div class="mb-4">
          <span
            class="text-[10px] font-black uppercase tracking-widest px-3 py-1 rounded-full font-mono border"
            :class="
              gameState?.is_game_over
                ? 'bg-brand-accent/20 text-amber-400 border-amber-500/30'
                : 'bg-brand-surface text-slate-400 border-brand-border'
            "
          >
            {{ gameState?.is_game_over ? $t("gameEnd.final") : $t("gameEnd.roundEnd") }}
          </span>
        </div>

        <div class="mt-4 mb-2">
          <h2
            v-if="isAmIWinner"
            class="text-2xl font-black tracking-wide text-transparent bg-clip-text bg-gradient-to-r from-amber-400 via-orange-400 to-yellow-500 uppercase animate-pulse"
          >
            {{ gameState?.is_game_over ? $t("gameEnd.winner") : $t("gameEnd.loser") }}
          </h2>
          <h2
            v-else
            class="text-2xl font-black tracking-wide text-brand-text-muted uppercase"
          >
            {{
              gameState?.is_game_over ? $t("gameEnd.gameOver") : $t("gameEnd.roundOver")
            }}
          </h2>

          <p class="text-xs text-slate-400 mt-2">
            {{ $t("gameEnd.winnerLabel")
            }}<span class="text-amber-400 font-bold font-mono">{{ winnerName }}</span>
          </p>
        </div>

        <div
          class="my-4 sm:my-5 short:my-3 text-5xl sm:text-6xl short:text-4xl drop-shadow-lg select-none"
        >
          <span v-if="isAmIWinner">🏆</span>
          <span v-else-if="gameState?.is_game_over">🥈</span>
          <span v-else>💀</span>
        </div>

        <div
          class="bg-brand-bg-dark/60 border border-slate-800 rounded-xl p-4 my-5 text-left"
        >
          <h4
            class="text-[10px] uppercase text-slate-500 font-black tracking-wider mb-3 font-mono"
          >
            {{ $t("gameEnd.scoreboardTitle") }}
          </h4>
          <div class="space-y-2">
            <div
              v-for="p in sortedPlayers"
              :key="p.id"
              class="flex items-center justify-between p-2 rounded-lg border transition-all"
              :class="
                p.id === gameState?.winner_id
                  ? 'bg-brand-accent/5 border-amber-500/20 shadow-sm'
                  : 'bg-brand-bg/40 border-transparent'
              "
            >
              <div class="flex items-center gap-2 truncate">
                <span v-if="p.id === gameState?.winner_id" class="text-xs">👑</span>
                <span v-else class="text-xs opacity-40">👤</span>
                <span
                  class="text-sm font-medium truncate"
                  :class="
                    p.id === myID ? 'text-amber-400 font-bold' : 'text-brand-text-subtle'
                  "
                >
                  {{ p.username || $t("common.opponent") }}
                  <span
                    v-if="p.id === myID"
                    class="text-[10px] text-slate-500 font-normal"
                    >({{ $t("common.you") }})</span
                  >
                </span>
              </div>

              <div class="flex items-center gap-1 font-mono text-sm font-bold">
                <span class="text-amber-400">★</span>
                <span
                  :class="
                    p.id === gameState?.winner_id
                      ? 'text-amber-400'
                      : 'text-brand-text-subtle'
                  "
                >
                  {{ p.score || 0 }}
                  <span class="text-slate-600 text-xs">{{
                    $t("gameEnd.scoreOutOf")
                  }}</span>
                </span>
              </div>
            </div>
          </div>
        </div>

        <div class="mt-6 space-y-2">
          <button
            v-if="!gameState?.is_game_over"
            type="button"
            :disabled="hasOpponentLeft"
            @click="$emit('next-round')"
            class="w-full py-3 font-bold rounded-xl shadow-lg transition-all text-sm uppercase tracking-wider"
            :class="
              hasOpponentLeft
                ? 'bg-brand-surface text-slate-500 cursor-not-allowed border border-brand-border shadow-none'
                : 'bg-gradient-to-r from-amber-500 to-orange-600 hover:from-amber-600 hover:to-orange-700 text-slate-950 font-black shadow-amber-500/10 cursor-pointer active:scale-95'
            "
          >
            {{
              hasOpponentLeft ? $t("gameEnd.waitingForPlayers") : $t("gameEnd.nextRound")
            }}
          </button>

          <button
            v-if="!gameState?.is_game_over && hasOpponentLeft"
            type="button"
            @click="$emit('leave-game')"
            class="w-full py-3 bg-rose-600 hover:bg-rose-700 text-white font-bold rounded-xl shadow-lg shadow-rose-500/10 transition-all cursor-pointer text-sm uppercase tracking-wider active:scale-95"
          >
            {{ $t("gameError.leave") }}
          </button>

          <template v-else-if="gameState?.is_game_over">
            <button
              v-if="gameState?.turn_order?.[0] === myID"
              @click="$emit('restart-game')"
              type="button"
              class="w-full py-3 bg-gradient-to-r from-amber-500 to-orange-600 hover:from-amber-600 hover:to-orange-700 text-slate-950 font-black rounded-xl shadow-lg shadow-amber-500/20 transition-all cursor-pointer text-sm uppercase tracking-wider active:scale-95"
            >
              {{ $t("gameEnd.restart") }}
            </button>

            <div
              v-else-if="!hasOpponentLeft"
              class="text-xs text-slate-400 animate-pulse py-2"
            >
              {{ $t("gameEnd.waitingForHost") }}
            </div>

            <button
              type="button"
              @click="$emit('leave-game')"
              class="w-full py-3 bg-slate-800 hover:bg-slate-700 border border-slate-700 text-slate-200 font-bold rounded-xl shadow-lg transition-all cursor-pointer text-sm uppercase tracking-wider active:scale-95"
            >
              {{ $t("gameEnd.leaveLobby") }}
            </button>
          </template>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { computed } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

const props = defineProps({
  isOpen: { type: Boolean, required: true },
  gameState: { type: Object, required: true },
  myID: { type: String, required: true },
  isSinglePlayer: { type: Boolean, default: false },
});

defineEmits(["next-round", "restart-game", "leave-game"]);

const isAmIWinner = computed(() => {
  return props.gameState?.winner_id === props.myID;
});

const winnerName = computed(() => {
  const wID = props.gameState?.winner_id;
  if (!wID || !props.gameState?.players) return t("common.nobody");

  const pList = props.gameState.players;
  if (Array.isArray(pList)) {
    const found = pList.find((p) => p && p.id === wID);
    return found ? found.username : t("common.opponent");
  }
  return pList[wID]?.username || t("common.opponent");
});

const sortedPlayers = computed(() => {
  const playersData = props.gameState?.players;
  if (!playersData) return [];

  let list = [];
  if (Array.isArray(playersData)) {
    list = playersData.filter((p) => p !== null);
  } else {
    list = Object.keys(playersData).map((id) => ({
      id: id,
      ...playersData[id],
    }));
  }

  return [...list].sort((a, b) => (b.score || 0) - (a.score || 0));
});

const hasOpponentLeft = computed(() => {
  const playersData = props.gameState?.players;
  if (!playersData) return true;

  const totalConnected = Array.isArray(playersData)
    ? playersData.filter((p) => p !== null).length
    : Object.keys(playersData).length;

  return totalConnected < 2;
});
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
</style>
