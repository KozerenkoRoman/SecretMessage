<template>
  <div
    class="h-screen bg-slate-900 text-white p-2 flex flex-col justify-between overflow-hidden relative font-sans select-none"
  >
    <!-- HEADER -->
    <header
      class="w-full flex justify-between items-center bg-slate-800/80 backdrop-blur px-3 py-2 rounded-xl border border-slate-700 flex-shrink-0 gap-4 z-10 h-[6vh]"
    >
      <div class="flex items-center gap-4">
        <button @click="handleLeaveGame" type="button" class="btn-danger">
          {{ $t("board.leave") }}
        </button>
        <div class="h-6 w-[1px] bg-slate-700"></div>
        <div
          class="flex items-center gap-2 bg-slate-900/60 pl-2 pr-1.5 py-0.5 rounded-lg border border-slate-700/40"
        >
          <span
            class="text-xs font-medium text-slate-300 hidden sm:inline max-w-[80px] truncate"
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
              class="w-full h-full object-cover rounded-full bg-slate-800"
            />
          </div>
        </div>
        <div>
          <h2 class="text-sm font-bold text-yellow-400 font-mono leading-none mb-0.5">
            {{ $t("board.room") }}{{ roomID }}
          </h2>
          <p class="text-xs text-slate-400">
            {{ $t("board.time") }}
            <span class="text-amber-400 font-mono font-bold text-xs">
              {{ gameState?.seconds_left || 0 }}{{ $t("board.timeUnit") }}
            </span>
          </p>
        </div>
      </div>
      <div class="flex items-center gap-4">
        <div class="flex items-center gap-2">
          <template v-if="!gameState?.is_started">
            <button v-if="canStartGame" @click="handleStartGame" class="btn-primary">
              {{ $t("board.start") }}
            </button>
            <div
              v-else
              class="status-pill bg-blue-500/20 text-blue-400 border-blue-500/40 animate-pulse"
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
              class="status-pill bg-slate-700 text-slate-300 border-transparent"
            >
              {{ $t("board.currentTurn") }}{{ currentTurnPlayerName }}
            </div>
          </template>
        </div>
        <div class="h-6 w-[1px] bg-slate-700"></div>
      </div>
    </header>

    <main
      class="flex-1 flex flex-col justify-between my-1 gap-1 overflow-hidden w-full max-w-7xl mx-auto"
    >
      <!-- Блок Опонентів верхній -->
      <div
        class="w-full flex justify-center items-center gap-6 flex-shrink-0 h-[26vh] max-h-[210px] overflow-hidden"
      >
        <template v-for="player in opponents" :key="player.id">
          <div
            :class="[
              isTurnOfPlayer(player.id)
                ? 'border-yellow-400 bg-slate-800 ring-2 ring-yellow-400/30'
                : 'border-slate-700 bg-slate-800/60',
              player.is_out ? 'opacity-40 grayscale-[30%]' : '',
            ]"
            class="flex flex-col items-center p-2 rounded-xl border transition-all duration-300 w-64 h-full justify-between shadow-md relative"
          >
            <!-- Рядок статусу гравця -->
            <div
              class="flex items-center justify-between w-full border-b border-slate-700 pb-1 flex-shrink-0 min-h-[28px]"
            >
              <span
                class="font-bold text-[11px] truncate max-w-[150px] transition-all duration-200 px-1.5 py-0.5 rounded"
                :class="[
                  player.is_out
                    ? 'bg-rose-900 text-rose-100 font-black'
                    : 'text-slate-200',
                ]"
                :title="player.username"
              >
                {{ player.username || $t("common.opponent") }}

                <template v-if="player.is_protected">
                  <span
                    class="font-bold text-[11px] truncate max-w-[150px] transition-all duration-200 px-1.5 py-0.5 rounded"
                    :class="['bg-yellow-500 text-slate-950 font-black shadow']"
                  >
                    {{ $t("board.protection") }}
                  </span>
                </template>

                <template v-if="player.is_out"> ({{ $t("status.out") }})</template>
              </span>
              <span
                class="text-[12px] bg-slate-700 text-yellow-400 px-1 py-0.5 rounded font-mono"
              >
                ★{{ player.score || 0 }}
              </span>
            </div>

            <!-- Карти в руці опонента (Центровані з динамічним накладанням без виходу за межі) -->
            <div
              class="flex justify-center items-center h-36 w-full overflow-hidden relative px-2"
            >
              <div
                class="flex flex-row items-center justify-center relative w-full gap-x-[-1.5rem] -space-x-4"
              >
                <!-- Сорочки карт в руці (можуть злегка перекриватися, flex-shrink дозволяє адаптацію) -->
                <div
                  v-for="cIdx in getOpponentHandCount(player)"
                  :key="cIdx"
                  class="w-24 h-36 sm:w-24 sm:h-36 rounded-lg border border-amber-500 bg-slate-950 p-[1px] shadow-[0_0_6px_rgba(245,158,11,0.4)] overflow-hidden relative transition-transform duration-300 flex-shrink"
                >
                  <img
                    :src="deckBackImage"
                    :alt="$t('common.cardBackAlt')"
                    class="w-full h-full object-cover rounded-sm select-none"
                  />
                </div>

                <!-- Остання зіграна карта (Завжди поверх інших z-20, не стискається flex-shrink-0, показується повністю) -->
                <div
                  v-if="lastPlayedCardsByPlayer[player.id] !== undefined"
                  class="flex flex-col items-center justify-center flex-shrink-0 z-20 ml-2"
                >
                  <div
                    class="w-24 h-36 sm:w-24 sm:h-36 rounded-lg border border-amber-500 bg-slate-950 p-[1px] shadow-[0_0_6px_rgba(245,158,11,0.4)] overflow-hidden relative transition-transform duration-300 flex-shrink"
                    :class="getCardColor(lastPlayedCardsByPlayer[player.id])"
                    :data-tooltip="getCardName(lastPlayedCardsByPlayer[player.id])"
                  >
                    <img
                      v-if="getCardImage(lastPlayedCardsByPlayer[player.id])"
                      :src="getCardImage(lastPlayedCardsByPlayer[player.id])"
                      :alt="getCardName(lastPlayedCardsByPlayer[player.id])"
                      class="w-full h-full rounded-sm pointer-events-none"
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

            <!-- Нижній відступ замість видаленого блоку статусів для збереження пропорцій геометрії -->
            <div class="h-1 w-full flex-shrink-0"></div>
          </div>
        </template>
      </div>

      <!-- Центральна частина: Колода та Стіл відбою -->
      <div
        class="h-[40vh] flex items-center justify-center p-3 bg-slate-950/40 rounded-2xl border border-slate-800/60 w-full overflow-hidden"
      >
        <div
          class="flex gap-6 items-center w-full max-w-none mx-auto justify-between h-full"
        >
          <!-- Колода -->
          <div class="flex flex-col items-center justify-center gap-1 flex-shrink-0">
            <div
              class="relative w-44 h-64 bg-gradient-to-br from-yellow-800 to-indigo-900 rounded-xl border-2 border-amber-500 shadow-lg shadow-amber-950/20 flex flex-col items-center justify-center bg-cover bg-center overflow-hidden p-[2px]"
              :style="{ backgroundImage: `url(${deckBackImage})` }"
            >
              <div
                class="absolute inset-0 bg-slate-950/50 flex flex-col items-center justify-center p-1"
              >
                <span
                  class="text-[12px] font-bold uppercase text-amber-200 tracking-wider bg-slate-950/80 px-2 py-1 rounded backdrop-blur-[1px]"
                >
                  {{ $t("board.deck") }}
                </span>
                <span
                  class="text-2xl font-black text-white leading-none mt-1.5 bg-slate-950/70 px-2.5 py-1 rounded font-mono shadow-md border border-white/5"
                >
                  {{ gameState?.deck ? gameState.deck.length : 0 }}
                </span>
              </div>
            </div>
            <div class="text-[16px] font-bold text-slate-400 font-mono mt-0.5">
              {{ $t("board.discardPile") }}{{ globalDiscardPile.length }}
            </div>
          </div>

          <div class="h-full max-h-[220px] w-[1px] bg-slate-800 flex-shrink-0"></div>

          <!-- Стіл відбою -->
          <div
            class="flex-1 flex flex-col justify-start gap-1 pl-1 h-full overflow-hidden"
          >
            <div
              class="flex justify-between items-center border-b border-slate-800 pb-1 flex-shrink-0"
            >
              <span
                class="text-[10px] uppercase text-amber-400 font-bold font-mono tracking-wider"
              >
                {{ $t("board.table") }}
              </span>
            </div>

            <div
              class="flex flex-wrap gap-1 justify-start items-start content-start w-full overflow-y-auto overflow-x-hidden custom-scrollbar flex-1 px-1 pt-4 pb-1 h-[16rem]"
            >
              <template
                v-for="slotIndex in globalDiscardPile.length"
                :key="`discard-slot-${slotIndex}`"
              >
                <div
                  v-if="globalDiscardPile[slotIndex - 1]"
                  class="game-card game-card-discard w-24 h-36 p-0 border border-slate-700/50 overflow-hidden bg-slate-950 flex-shrink-0"
                  :class="getCardColor(globalDiscardPile[slotIndex - 1].type)"
                  :data-tooltip="
                    $t('board.discardTooltip', {
                      name: getCardName(globalDiscardPile[slotIndex - 1].type),
                      owner: globalDiscardPile[slotIndex - 1].owner,
                    })
                  "
                >
                  <img
                    v-if="getCardImage(globalDiscardPile[slotIndex - 1].type)"
                    :src="getCardImage(globalDiscardPile[slotIndex - 1].type)"
                    :alt="getCardName(globalDiscardPile[slotIndex - 1].type)"
                    class="w-full h-full object-cover pointer-events-none rounded-[0.4rem]"
                  />
                  <div
                    v-else
                    class="w-full h-full flex items-center justify-center text-[10px] text-center p-1"
                  >
                    {{ getCardName(globalDiscardPile[slotIndex - 1].type) }}
                  </div>
                </div>
              </template>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- НИЖНЯ ПАНЕЛЬ (РУКА ВЛАСНОГО ГРАВЦЯ) -->
    <footer
      class="w-full max-w-2xl mx-auto bg-slate-950/90 backdrop-blur-md p-2 rounded-t-2xl border-t border-x border-slate-800 flex flex-col items-center shadow-2xl flex-shrink-0 z-10 h-[28vh] min-h-[280px] relative"
    >
      <div class="w-full h-full relative">
        <div
          v-if="myPlayer?.score > 0"
          class="absolute left-2 top-1/2 -translate-y-1/2 flex flex-col gap-0.1 items-center p-1.5"
          :title="$t('common.myChipsTitle')"
        >
          <img
            v-for="n in myPlayer.score"
            :key="`my-chip-${n}`"
            :src="chipImage"
            :alt="$t('common.chipAlt')"
            class="w-20 h-8 object-contain animate-fade-in"
          />
        </div>

        <!-- CHANCELLOR PANEL -->
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
            class="flex justify-center gap-4 items-center flex-1 w-full overflow-hidden"
          >
            <div
              v-for="(cardType, cIdx) in myHandCards"
              :key="`chancellor-card-${cIdx}-${cardType}`"
              @click="handleChancellorClick(cIdx)"
              class="game-card w-44 ring-2 ring-amber-500/50 bg-cover bg-center transition-all duration-200 hover:scale-105 cursor-pointer flex flex-col justify-between overflow-hidden"
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
                class="text-[12px] font-bold font-mono text-center block bg-slate-950/80 p-1 mt-auto z-10 relative text-amber-300 uppercase tracking-wider mx-[-0.5rem] mb-[-0.5rem] rounded-b-xl border-t border-white/5"
              >
                {{ getCardName(cardType) }}
              </span>
            </div>
          </div>
        </div>

        <!-- Контент руки (Стандартний стан) -->
        <div
          v-else
          class="flex justify-center gap-4 relative z-10 items-center w-full h-full"
        >
          <div v-if="myHandCards.length === 0" class="text-slate-500 text-xs italic">
            {{ $t("board.waitingForCards") }}
          </div>

          <div
            v-for="(cardType, cIdx) in myHandCards"
            :key="`hand-card-${cIdx}-${cardType}`"
            @click="isMyTurn ? handleCardClick(cardType, cIdx) : null"
            class="game-card w-44 bg-cover bg-center flex flex-col justify-between overflow-hidden"
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
              class="text-[12px] font-bold font-mono text-center block bg-slate-950/80 p-1 mt-auto z-10 relative text-amber-300 uppercase tracking-wider mx-[-0.5rem] mb-[-0.5rem] rounded-b-xl border-t border-white/5"
            >
              {{ getCardName(cardType) }}
            </span>
          </div>

          <!-- Карта захисту у себе -->
          <div
            v-if="myPlayer?.is_protected"
            class="game-card w-24 h-36 border-2 border-yellow-400 shadow-[0_0_10px_rgba(234,179,8,0.5)] bg-cover bg-center animate-fade-in self-center relative flex-shrink-0 flex flex-col justify-between overflow-hidden"
            :style="
              getCardImage(4)
                ? { backgroundImage: `url(${getCardImage(4)})` }
                : { backgroundColor: '#1e293b' }
            "
            :data-tooltip="$t('board.protectionTooltip', { name: getCardName(4) })"
          >
            <span
              class="absolute -top-2 -right-1 bg-yellow-500 text-slate-950 font-black text-[8px] px-1 rounded shadow uppercase tracking-wider z-20"
            >
              {{ $t("status.protected") }}
            </span>
            <span
              class="text-[10px] font-bold font-mono text-center block bg-slate-950/80 p-0.5 mt-auto z-10 relative text-amber-300 uppercase tracking-wider mx-[-0.5rem] mb-[-0.5rem] rounded-b-xl border-t border-white/5"
            >
              {{ getCardName(4) }}
            </span>
          </div>
        </div>
      </div>
    </footer>

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
import { ref, computed, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useGameStore } from "../stores/gameStore";
import { storeToRefs } from "pinia";
import { getAvatarUrl } from "../utils/avatar";
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

watch(
  () => props.gameState,
  (newGameState, oldGameState) => {
    if (!newGameState || !newGameState.players || !newGameState.turn_order) return;
    const currentPlayers = Array.isArray(newGameState.players)
      ? newGameState.players
      : Object.values(newGameState.players);

    const oldPlayers = oldGameState?.players
      ? Array.isArray(oldGameState.players)
        ? oldGameState.players
        : Object.values(oldGameState.players)
      : [];
    const currentActiveID = newGameState.turn_order[newGameState.current_turn];
    currentPlayers.forEach((p) => {
      if (!p || !p.id) return;

      if (p.id === effectiveMyID.value) return;

      const oldP = oldPlayers.find((o) => o && o.id === p.id);
      const currentDiscardLength = Array.isArray(p.discard_pile)
        ? p.discard_pile.length
        : 0;
      const oldDiscardLength =
        oldP && Array.isArray(oldP.discard_pile) ? oldP.discard_pile.length : 0;

      if (currentDiscardLength > oldDiscardLength) {
        const lastCard = p.discard_pile[currentDiscardLength - 1];
        lastPlayedCardsByPlayer.value[p.id] = lastCard;
      } else if (p.id === currentActiveID) {
        delete lastPlayedCardsByPlayer.value[p.id];
      }
      if (currentDiscardLength === 0) {
        delete lastPlayedCardsByPlayer.value[p.id];
      }
    });
  },
  { deep: true, immediate: true }
);

watch(
  () => totalCardsInHands.value,
  (newCount, oldCount) => {
    if (newCount > oldCount) {
      lastPlayedCardsByPlayer.value = {};
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
    }
  }
);

watch(
  () => props.gameState?.deck?.length,
  (newDeckLength, oldDeckLength) => {
    if (newDeckLength && oldDeckLength && newDeckLength > oldDeckLength) {
      handleCloseRevealModal();
      lastPlayedCardsByPlayer.value = {};
    }
  }
);

const handleCloseRevealModal = () => {
  gameStore.clearRevealedData();
};

const handleClearError = () => {
  gameStore.clearError();
};

const handleStartGame = () => emit("start-game");

const getCardDesc = (type) => getCardInfoHelper(type)?.desc || t("cards.noDescription");
const getCardColor = (type) => getCardInfoHelper(type)?.color || "bg-slate-700";
const getCardName = (type) => getCardInfoHelper(type)?.name || t("cards.unknown");
const getCardImage = (type) => getCardInfoHelper(type)?.image || "";
const getCardValue = (type) => {
  const val = getCardInfoHelper(type)?.value;
  return val !== undefined ? val : "?";
};

const globalDiscardPile = computed(() => {
  const playersData = props.gameState?.players;
  if (!playersData) return [];
  const allDiscards = [];

  const processPlayerDiscard = (p, id) => {
    if (p && Array.isArray(p.discard_pile)) {
      p.discard_pile.forEach((card, index) => {
        const isObject = typeof card === "object" && card !== null;
        const cardType = isObject ? card.type : card;
        const turnOrder = isObject ? card.turn || card.timestamp || index : index;

        allDiscards.push({
          type: cardType,
          owner: p.username || t("common.player"),
          playerId: id,
          turn: turnOrder,
        });
      });
    }
  };

  if (typeof playersData === "object" && !Array.isArray(playersData)) {
    Object.keys(playersData).forEach((id) => {
      processPlayerDiscard(playersData[id], id);
    });
  } else if (Array.isArray(playersData)) {
    playersData.forEach((p) => {
      if (p) processPlayerDiscard(p, p.id);
    });
  }
  return allDiscards.sort((a, b) => a.turn - b.turn);
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
