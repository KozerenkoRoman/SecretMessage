<template>
  <div
    class="h-app bg-brand-bg text-white p-2 grid grid-cols-1 lg:grid-cols-12 overflow-hidden relative font-sans select-none gap-2 sm:gap-3"
  >
    <div class="lg:col-span-8 flex flex-col h-full min-h-0 overflow-hidden">
      <header
        class="w-full flex justify-between items-center bg-brand-surface/80 backdrop-blur px-3 py-2 rounded-xl border border-brand-border flex-shrink-0 gap-2 sm:gap-4 z-10 min-h-[44px] short:min-h-[40px] tall:min-h-[56px]"
      >
        <div class="flex items-center gap-4">
          <button @click="handleLeaveGame" type="button" class="btn-danger">
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
              {{ $t("board.time")
              }}<span class="text-amber-400 font-mono font-bold text-sm"
                >{{ gameStore.gameState?.seconds_left ?? 0
                }}{{ $t("board.timeUnit") }}</span
              >
            </p>
          </div>
        </div>

        <div class="flex items-center gap-4">
          <div class="flex items-center gap-2">
            <template
              v-if="
                !gameState?.is_started &&
                (!gameState?.turn_order || gameState.turn_order.length === 0)
              "
            >
              <button
                v-if="canStartGame"
                @click="handleStartGame"
                type="button"
                class="btn-primary disabled:opacity-50 disabled:cursor-not-allowed transition-all"
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

      <main
        class="flex-1 min-h-0 flex flex-col my-1 gap-1 sm:gap-2 overflow-hidden w-full mx-auto"
      >
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
                    >
                      {{ $t("board.protection") }}
                    </span>
                  </template>
                  <template v-if="player.is_out"> ({{ $t("status.out") }})</template>
                </span>
                <span
                  class="text-[12px] tall:text-xs retina:text-sm bg-brand-surface-dim text-yellow-400 px-1 py-0.5 rounded font-mono flex-shrink-0"
                >
                  ★{{ player.score || 0 }}
                </span>
              </div>
              <div
                class="flex justify-center items-center w-full relative px-2 sm:px-3 lg:px-5 py-2 flex-1 min-h-0"
              >
                <div
                  class="flex flex-row items-center justify-center relative w-full -space-x-4 flex-shrink-0"
                >
                  <div
                    v-for="cIdx in getOpponentHandCount(player)"
                    :key="cIdx"
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

        <div
          class="flex-1 min-h-0 flex items-center justify-center p-2 sm:p-3 bg-brand-bg-dark/40 rounded-2xl border border-slate-800/60 w-full overflow-hidden"
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
            <div class="flex-scroll flex flex-col justify-start gap-1 pl-1 h-full">
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
                  :data-tooltip="
                    $t('board.discardTooltip', {
                      name: getCardName(entry.type),
                      owner: entry.owner,
                    })
                  "
                >
                  <img
                    v-if="getCardImage(entry.type)"
                    :src="getCardImage(entry.type)"
                    :alt="getCardName(entry.type)"
                    class="w-full h-full object-cover pointer-events-none rounded-[0.4rem]"
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
        </div>
      </main>

      <footer
        class="w-full max-w-2xl xtall:max-w-3xl mx-auto bg-brand-bg-dark/90 backdrop-blur-md p-1.5 sm:p-2 short:p-1 rounded-t-2xl border-t border-x border-slate-800 flex flex-col items-center shadow-2xl flex-shrink-0 z-10 relative"
        style="height: calc(var(--card-primary-h) + 1.5rem)"
      >
        <div class="w-full h-full relative">
          <div
            v-if="myPlayer?.score > 0"
            class="absolute left-1 sm:left-2 top-1/2 -translate-y-1/2 hidden sm:flex flex-col gap-0.1 items-center p-1 sm:p-1.5 max-h-full overflow-hidden"
            :title="$t('common.myChipsTitle')"
          >
            <img
              v-for="n in myPlayer.score"
              :key="`my-chip-${n}`"
              :src="chipImage"
              :alt="$t('common.chipAlt')"
              class="w-12 sm:w-16 lg:w-20 h-auto object-contain animate-fade-in flex-shrink-0"
            />
          </div>
          <div
            v-if="showChancellorPanel"
            class="w-full flex flex-col items-center h-full justify-between"
          >
            <div class="text-center">
              <h4 class="text-amber-400 font-bold text-xs uppercase leading-none">
                {{ $t("board.chancellorPickOwn") }}
              </h4>
            </div>
            <div
              class="flex justify-center gap-2 sm:gap-3 lg:gap-4 items-center flex-1 min-h-0 w-full overflow-hidden"
            >
              <div
                v-for="(cardType, cIdx) in myHandCards"
                :key="`chancellor-card-${cIdx}-${cardType}`"
                @click="handleChancellorClick(cIdx)"
                class="game-card card-primary ring-2 ring-amber-500/50 bg-cover bg-center transition-all duration-200 hover:scale-105 cursor-pointer flex flex-col justify-between overflow-hidden"
                :class="getCardColor(cardType)"
                :style="
                  getCardImage(cardType)
                    ? { backgroundImage: `url(${getCardImage(cardType)})` }
                    : {}
                "
                :data-tooltip="`${getCardName(cardType)}(${getCardValue(
                  cardType
                )}) — ${getCardDesc(cardType)}`"
              >
                <span
                  class="text-[12px] font-bold font-mono text-center block bg-brand-bg-dark/80 p-1 mt-auto z-10 relative text-amber-300 uppercase tracking-wider mx-[-0.5rem] mb-[-0.5rem] rounded-b-xl border-t border-white/5"
                  >{{ getCardName(cardType) }}</span
                >
              </div>
            </div>
          </div>
          <div
            v-else
            class="flex justify-center gap-2 sm:gap-3 lg:gap-4 relative z-10 items-center w-full h-full"
          >
            <div v-if="myHandCards.length === 0" class="text-slate-500 text-xs italic">
              {{ $t("board.waitingForCards") }}
            </div>
            <div
              v-for="(cardType, cIdx) in myHandCards"
              :key="`hand-card-${cIdx}-${cardType}`"
              @click="isMyTurn ? handleCardClick(cardType, cIdx) : null"
              class="game-card card-primary bg-cover bg-center flex flex-col justify-between overflow-hidden"
              :class="[
                getCardColor(cardType),
                isMyTurn ? 'game-card-playable' : 'game-card-disabled',
              ]"
              :style="
                getCardImage(cardType)
                  ? { backgroundImage: `url(${getCardImage(cardType)})` }
                  : {}
              "
              :data-tooltip="`${getCardName(cardType)}(${getCardValue(
                cardType
              )}) — ${getCardDesc(cardType)}`"
            >
              <span
                class="text-[12px] font-bold font-mono text-center block bg-brand-bg-dark/80 p-1 mt-auto z-10 relative text-amber-300 uppercase tracking-wider mx-[-0.5rem] mb-[-0.5rem] rounded-b-xl border-t border-white/5"
                >{{ getCardName(cardType) }}</span
              >
            </div>
            <div
              v-if="myPlayer?.is_protected"
              class="game-card card-secondary border-2 border-yellow-400 shadow-[0_0_10px_rgba(234,179,8,0.5)] bg-cover bg-center animate-fade-in self-center relative flex flex-col justify-between overflow-hidden"
              :style="
                getCardImage(4)
                  ? { backgroundImage: `url(${getCardImage(4)})` }
                  : { backgroundColor: '#1e293b' }
              "
              :data-tooltip="$t('board.protectionTooltip', { name: getCardName(4) })"
            >
              <span
                class="absolute -top-2 -right-1 bg-brand-warning text-slate-950 font-black text-[8px] px-1 rounded shadow uppercase tracking-wider z-20"
                >{{ $t("status.protected") }}</span
              >
              <span
                class="text-[10px] font-bold font-mono text-center block bg-brand-bg-dark/80 p-0.5 mt-auto z-10 relative text-amber-300 uppercase tracking-wider mx-[-0.5rem] mb-[-0.5rem] rounded-b-xl border-t border-white/5"
                >{{ getCardName(4) }}</span
              >
            </div>
          </div>
        </div>
      </footer>
    </div>

    <div
      class="lg:col-span-4 h-full max-h-app overflow-hidden hidden lg:block pt-1 min-h-0"
    >
      <GameLogPanel />
    </div>

    <ActionModal
      :is-open="showActionModal"
      :card-type="String(activePlay.cardType)"
      :hand-index="activePlay.handIndex"
      :players="arrangedPlayers"
      :myID="effectiveMyID"
      @close="showActionModal = false"
      @submit="handleActionModalSubmit"
    />
    <CardRevealModal
      :is-open="!!revealedCardData"
      v-if="revealedCardData"
      :data="revealedCardData"
      :players="props.gameState.players"
      @close="handleCloseRevealModal"
    />
    <ChancellorModal
      :is-open="showChancellorPanel"
      :cards="myHandCards"
      @submit="handleChancellorModalSubmit"
    />
    <GameEndModal
      v-if="!revealedCardData"
      :is-open="showGameEndModal"
      :isSinglePlayer="isOpponentLeft"
      :game-state="props.gameState"
      :myID="effectiveMyID"
      @next-round="handleNextRoundRequest"
      @restart-game="handleRestartGameRequest"
      @leave-game="emit('leave-game')"
    />
    <GameErrorModal :message="gameStore.error" @close="handleClearError" />
    <ConfirmModal
      :is-open="showLeaveConfirm"
      :title="$t('board.leaveConfirm.title')"
      :message="$t('board.leaveConfirm.message')"
      :confirm-text="$t('board.leaveConfirm.confirm')"
      :cancel-text="$t('board.leaveConfirm.cancel')"
      @confirm="handleConfirmLeave"
      @cancel="showLeaveConfirm = false"
    />
  </div>
</template>

<script setup>
import { getCardInfoHelper } from "../constants/cards";
import { ref, computed, watch, onUnmounted } from "vue";
import { useI18n } from "vue-i18n";
import { useGameStore } from "../stores/gameStore";
import { storeToRefs } from "pinia";
import { getAvatarUrl } from "../utils/avatar";
import GameLogPanel from "../components/GameLogPanel.vue";
import ChancellorModal from "./ChancellorModal.vue";
import ActionModal from "./ActionModal.vue";
import CardRevealModal from "./CardRevealModal.vue";
import GameEndModal from "./GameEndModal.vue";
import GameErrorModal from "./GameErrorModal.vue";
import ConfirmModal from "./ConfirmModal.vue";
import deckBackImage from "../assets/deckBack.png";
import chipImage from "../assets/chip.png";

const props = defineProps({
  roomID: { type: String, required: true },
  gameState: { type: Object, required: true },
  myID: { type: String, default: "" },
  isStarting: { type: Boolean, default: false },
});

const emit = defineEmits([
  "play-card",
  "start-game",
  "leave-game",
  "next-round",
  "restart-game",
]);

const { t } = useI18n();
const gameStore = useGameStore();
const { revealedCardData, myID: storeMyID } = storeToRefs(gameStore);

const showLeaveConfirm = ref(false);
const showActionModal = ref(false);
const activePlay = ref({ cardType: "", handIndex: 0, targetID: "", guessCard: "" });
const lastPlayedCardsByPlayer = ref({});
const localSecondsLeft = ref(0);
let localTimerInterval = null;

const discardSequence = ref([]);
const discardSeenLengths = ref({});
let discardSeqCounter = 0;

const resetDiscardTracking = () => {
  discardSequence.value = [];
  discardSeenLengths.value = {};
  discardSeqCounter = 0;
};

const effectiveMyID = computed(() => {
  const fromStore = storeMyID.value;
  if (typeof fromStore === "string" && fromStore.length > 0) return fromStore;
  return typeof props.myID === "string" ? props.myID : "";
});

const totalCardsInHands = computed(() => {
  const playersData = props.gameState?.players;
  if (!playersData) return 0;
  const playersList = Array.isArray(playersData)
    ? playersData
    : Object.values(playersData);
  return playersList.reduce((sum, p) => {
    if (!p) return sum;
    const count =
      typeof p.hand_count === "number"
        ? p.hand_count
        : Array.isArray(p.hand)
        ? p.hand.length
        : 0;
    return sum + count;
  }, 0);
});

const arrangedPlayers = computed(() => {
  const playersData = props.gameState?.players;
  if (!playersData) return [];
  const list = Array.isArray(playersData)
    ? props.gameState.players
        .filter(
          (p) => p && typeof p === "object" && typeof p.id === "string" && p.id.length > 0
        )
        .map((p) => ({ ...p }))
    : Object.keys(playersData)
        .map((id) => {
          const raw = playersData[id];
          if (!raw || typeof raw !== "object") return null;
          return { ...raw, id };
        })
        .filter((p) => p !== null && typeof p.id === "string" && p.id.length > 0);

  const turnOrder = props.gameState?.turn_order;
  if (Array.isArray(turnOrder) && turnOrder.length > 0) {
    const indexOf = new Map(turnOrder.map((id, i) => [id, i]));
    list.sort((a, b) => {
      const ai = indexOf.has(a.id) ? indexOf.get(a.id) : Number.MAX_SAFE_INTEGER;
      const bi = indexOf.has(b.id) ? indexOf.get(b.id) : Number.MAX_SAFE_INTEGER;
      return ai - bi;
    });
  }
  return list;
});

const opponents = computed(() => {
  const me = effectiveMyID.value;
  return arrangedPlayers.value.filter((p) => p.id !== me);
});

const myPlayer = computed(() => {
  const me = effectiveMyID.value;
  if (!me) return null;
  const playerObj = arrangedPlayers.value.find((p) => p.id === me) || null;
  if (playerObj && (!playerObj.avatar_seed || playerObj.avatar_seed === "")) {
    playerObj.avatar_seed = localStorage.getItem("avatar_seed") || "default_seed";
  }
  return playerObj;
});

const myHandCards = computed(() => {
  const me = myPlayer.value;
  if (me && Array.isArray(me.hand)) return me.hand;
  const fallback = props.gameState?.my_hand;
  return Array.isArray(fallback) ? fallback : [];
});

const isMyTurn = computed(() => {
  const me = effectiveMyID.value;
  if (!me) return false;
  const turnOrder = props.gameState?.turn_order;
  const currentTurnIdx = props.gameState?.current_turn;
  if (Array.isArray(turnOrder) && typeof currentTurnIdx === "number") {
    return turnOrder[currentTurnIdx] === me;
  }
  return props.gameState?.current_player_id === me;
});

onUnmounted(() => {
  if (localTimerInterval) clearInterval(localTimerInterval);
});

watch(
  ...[
    () => props.gameState,
    (newGameState, oldGameState) => {
      if (!newGameState || !newGameState.players || !newGameState.turn_order) return;
      const playersData = newGameState.players;
      const turnOrder = newGameState.turn_order;
      const playerById = {};
      if (Array.isArray(playersData)) {
        playersData.forEach((p) => {
          if (p && typeof p.id === "string") playerById[p.id] = p;
        });
      } else {
        Object.keys(playersData).forEach((id) => {
          const raw = playersData[id];
          if (raw && typeof raw === "object") playerById[id] = { ...raw, id };
        });
      }
      const oldPlayers = oldGameState?.players
        ? Array.isArray(oldGameState.players)
          ? oldGameState.players
          : Object.values(oldGameState.players)
        : [];
      const currentActiveID = turnOrder[newGameState.current_turn];
      const idsInTurnOrder = Array.isArray(turnOrder) ? [...turnOrder] : [];
      const allKnownIds = Object.keys(playerById);
      const extraIds = allKnownIds.filter((id) => !idsInTurnOrder.includes(id));
      const orderedIds = [...idsInTurnOrder, ...extraIds];

      orderedIds.forEach((pid) => {
        const p = playerById[pid];
        if (!p) return;
        const oldP = oldPlayers.find((o) => o && o.id === pid);
        const currentDiscardLength = Array.isArray(p.discard_pile)
          ? p.discard_pile.length
          : 0;
        const oldDiscardLength =
          oldP && Array.isArray(oldP.discard_pile) ? oldP.discard_pile.length : 0;
        const seenLen = discardSeenLengths.value[pid] || 0;

        if (currentDiscardLength > seenLen) {
          for (let i = seenLen; i < currentDiscardLength; i++) {
            const rawCard = p.discard_pile[i];
            const isObject = typeof rawCard === "object" && rawCard !== null;
            const cardType = isObject ? rawCard.type : rawCard;
            discardSequence.value.push({
              seq: discardSeqCounter++,
              type: cardType,
              owner: p.username || t("common.player"),
              playerId: pid,
            });
          }
          discardSeenLengths.value[pid] = currentDiscardLength;
        } else if (currentDiscardLength < seenLen) {
          discardSeenLengths.value[pid] = currentDiscardLength;
        }

        if (pid === effectiveMyID.value) return;

        if (currentDiscardLength > oldDiscardLength) {
          const lastCard = p.discard_pile[currentDiscardLength - 1];
          lastPlayedCardsByPlayer.value[pid] = lastCard;
        } else if (pid === currentActiveID) {
          delete lastPlayedCardsByPlayer.value[pid];
        }

        if (currentDiscardLength === 0) {
          delete lastPlayedCardsByPlayer.value[pid];
        }
      });
    },
    { deep: true, immediate: true },
  ]
);

watch(
  () => totalCardsInHands.value,
  (newCount, oldCount) => {
    if (newCount > oldCount) {
      lastPlayedCardsByPlayer.value = {};
      resetDiscardTracking();
    }
  }
);

watch(
  () => props.gameState?.round_number,
  (newRound, oldRound) => {
    if (
      typeof newRound === "number" &&
      typeof oldRound === "number" &&
      newRound > oldRound
    ) {
      handleCloseRevealModal();
      lastPlayedCardsByPlayer.value = {};
      resetDiscardTracking();
    }
  }
);

watch(
  () => props.gameState?.deck?.length,
  (newDeckLength, oldDeckLength) => {
    if (newDeckLength && oldDeckLength && newDeckLength > oldDeckLength) {
      handleCloseRevealModal();
      lastPlayedCardsByPlayer.value = {};
      resetDiscardTracking();
    }
  }
);

watch(
  () => gameStore.gameState?.seconds_left,
  (newSeconds) => {
    if (localTimerInterval) clearInterval(localTimerInterval);
    localSecondsLeft.value = typeof newSeconds === "number" ? newSeconds : 0;
    if (localSecondsLeft.value > 0) {
      localTimerInterval = setInterval(() => {
        if (localSecondsLeft.value > 0) {
          localSecondsLeft.value--;
          if (gameStore.gameState) {
            gameStore.gameState.seconds_left = localSecondsLeft.value;
          }
        } else {
          clearInterval(localTimerInterval);
        }
      }, 1000);
    }
  },
  { immediate: true }
);

const handleCloseRevealModal = () => {
  gameStore.clearRevealedData();
};

const handleClearError = () => {
  gameStore.clearError();
};

const handleStartGame = () => {
  if (
    props.isStarting ||
    props.gameState?.is_started ||
    (props.gameState?.turn_order && props.gameState.turn_order.length > 0)
  ) {
    return;
  }
  emit("start-game");
};
const getCardDesc = (type) =>
  getCardInfoHelper(type, t)?.desc || t("cards.noDescription");
const getCardColor = (type) => getCardInfoHelper(type)?.color || "bg-brand-surface-dim";
const getCardName = (type) => getCardInfoHelper(type, t)?.name || t("cards.unknown");
const getCardImage = (type) => getCardInfoHelper(type)?.image || "";
const getCardValue = (type) => {
  const val = getCardInfoHelper(type)?.value;
  return val !== undefined ? val : "?";
};

const globalDiscardPile = computed(() => {
  const playersData = props.gameState?.players;
  const lookupOwner = (pid) => {
    if (!playersData || !pid) return t("common.player");
    if (Array.isArray(playersData)) {
      const found = playersData.find((p) => p && p.id === pid);
      return found?.username || t("common.player");
    }
    return playersData[pid]?.username || t("common.player");
  };
  return discardSequence.value.map((entry) => ({
    type: entry.type,
    owner: lookupOwner(entry.playerId),
    playerId: entry.playerId,
    seq: entry.seq,
  }));
});

const canStartGame = computed(() => {
  if (props.gameState?.is_started) return false;
  const playersData = props.gameState?.players;
  if (!playersData) return false;
  const count = Array.isArray(playersData)
    ? playersData.filter((p) => p !== null).length
    : Object.keys(playersData).length;
  return count >= 2;
});

const handleLeaveGame = () => {
  showLeaveConfirm.value = true;
};

const handleConfirmLeave = () => {
  showLeaveConfirm.value = false;
  emit("leave-game");
};

const currentTurnPlayerName = computed(() => {
  const activeID = props.gameState?.current_player_id;
  if (!activeID) return t("common.opponent");
  const found = arrangedPlayers.value.find((p) => p.id === activeID);
  return found?.username || t("common.opponent");
});

const getOpponentHandCount = (player) => {
  if (!player) return 0;
  if (typeof player.hand_count === "number") return player.hand_count;
  if (Array.isArray(player.hand)) return player.hand.length;
  return 0;
};

const isTurnOfPlayer = (id) => {
  return props.gameState?.current_player_id === id;
};

const showGameEndModal = computed(() => {
  const phase = props.gameState?.phase;
  const isOver = props.gameState?.is_game_over;
  return phase === "ROUND_END" || phase === "PhaseRoundEnd" || isOver === true;
});

const handleCardClick = (cardType, index) => {
  if (!isMyTurn.value) return;
  activePlay.value = { cardType, handIndex: index, targetID: "", guessCard: "" };
  showActionModal.value = true;
};

const handleActionModalSubmit = ({ handIndex, targetID, guessCardId }) => {
  showActionModal.value = false;
  const me = effectiveMyID.value;
  if (!me) return;
  const actionPayload = {
    player_id: me,
    is_chancellor_type: false,
    hand_index: Number(handIndex),
    target_id: targetID,
    guess_card: guessCardId,
  };
  emit("play-card", actionPayload);
};

const handleNextRoundRequest = () => {
  emit("next-round");
};

const handleRestartGameRequest = () => {
  emit("restart-game");
};

const showChancellorPanel = computed(() => {
  const isChancellorPhase = props.gameState?.phase === "RESOLVE_CHANCELLOR";
  const hasThreeCards = myHandCards.value.length === 3;
  return isChancellorPhase && hasThreeCards;
});

const isOpponentLeft = computed(() => {
  const playersData = props.gameState?.players;
  if (!playersData) return true;
  const playersList = Array.isArray(playersData)
    ? playersData.filter((p) => p !== null)
    : Object.values(playersData);
  return playersList.length < 2;
});

const handleChancellorClick = (index) => {};
const handleChancellorModalSubmit = ({ keepHandIndex, bottomOrder }) => {
  const me = effectiveMyID.value;
  if (!me) return;
  const chancellorPayload = {
    player_id: me,
    keep_hand_index: Number(keepHandIndex),
    bottom_order: bottomOrder.map((card) => Number(card)),
  };
  gameStore.sendWSMessage("CHANCELLOR_RESOLVE", null, null, chancellorPayload);
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
  background: rgba(15, 23, 42, 0.6);
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(245, 158, 11, 0.3);
  border-radius: 2px;
}
</style>
