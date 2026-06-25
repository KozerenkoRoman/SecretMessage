<template>
  <div
    class="h-app bg-brand-bg text-white p-2 grid grid-cols-1 lg:grid-cols-12 overflow-hidden relative font-sans select-none gap-2 sm:gap-3"
  >
    <div class="lg:col-span-8 flex flex-col h-full min-h-0 overflow-hidden">
      <BoardHeader
        :roomID="roomID"
        :gameState="gameState"
        :myPlayer="myPlayer"
        :secondsLeft="localSecondsLeft"
        :canStartGame="canStartGame"
        :playersCount="arrangedPlayers.length"
        :isBotSubmitting="isBotSubmitting"
        :isStarting="isStarting"
        :showChancellorPanel="showChancellorPanel"
        :isMyTurn="isMyTurn"
        :currentTurnPlayerName="currentTurnPlayerName"
        @leave="handleLeaveGame"
        @add-bot="handleDelayAndAddBot"
        @start-game="handleStartGame"
      />

      <main
        class="flex-1 min-h-0 flex flex-col my-1 gap-1 sm:gap-2 overflow-hidden w-full mx-auto"
      >
        <BoardOpponents
          :opponents="opponents"
          :lastPlayedCardsByPlayer="lastPlayedCardsByPlayer"
          :deckBackImage="deckBackImage"
          :isTurnOfPlayer="isTurnOfPlayer"
          :getOpponentHandCount="getOpponentHandCount"
          :getCardColor="getCardColor"
          :getCardName="getCardName"
          :getCardImage="getCardImage"
        />
        <BoardTable
          :gameState="gameState"
          :globalDiscardPile="globalDiscardPile"
          :deckBackImage="deckBackImage"
          :getCardColor="getCardColor"
          :getCardName="getCardName"
          :getCardImage="getCardImage"
          :getCardValue="getCardValue"
          :latest-turn-alert="latestTurnAlert"
        />
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
                v-for="slot in localHandSlots"
                :key="slot.uid"
                @click="handleChancellorClick(slot.index)"
                class="game-card card-primary ring-2 ring-amber-500/50 bg-cover bg-center transition-all duration-200 hover:scale-105 cursor-pointer flex flex-col justify-between overflow-hidden"
                :class="getCardColor(slot.cardType)"
                :style="
                  getCardImage(slot.cardType)
                    ? { backgroundImage: `url(${getCardImage(slot.cardType)})` }
                    : {}
                "
                :data-tooltip="`${getCardName(slot.cardType)}(${getCardValue(
                  slot.cardType
                )}) - ${getCardDesc(slot.cardType)}`"
              >
                <span
                  class="text-[12px] font-bold font-mono text-center block bg-brand-bg-dark/80 p-1 mt-auto z-10 relative text-amber-300 uppercase tracking-wider mx-[-0.5rem] mb-[-0.5rem] rounded-b-xl border-t border-white/5"
                  >{{ getCardName(slot.cardType) }}</span
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
              v-for="slot in localHandSlots"
              :key="slot.uid"
              @click="isMyTurn ? handleCardClick(slot.cardType, slot.index) : null"
              class="game-card card-primary bg-cover bg-center flex flex-col justify-between overflow-hidden"
              :class="[
                getCardColor(slot.cardType),
                isMyTurn ? 'game-card-playable' : 'game-card-disabled',
              ]"
              :style="
                getCardImage(slot.cardType)
                  ? { backgroundImage: `url(${getCardImage(slot.cardType)})` }
                  : {}
              "
              :data-tooltip="`${getCardName(slot.cardType)}(${getCardValue(
                slot.cardType
              )}) - ${getCardDesc(slot.cardType)}`"
            >
              <span
                class="text-[12px] font-bold font-mono text-center block bg-brand-bg-dark/80 p-1 mt-auto z-10 relative text-amber-300 uppercase tracking-wider mx-[-0.5rem] mb-[-0.5rem] rounded-b-xl border-t border-white/5"
                >{{ getCardName(slot.cardType) }}</span
              >
            </div>

            <div
              v-if="myPlayer?.is_protected"
              class="game-card card-secondary border-2 border-yellow-400 shadow-[0_0_10px_color-mix(in_srgb,var(--color-brand-warning)_50%,transparent)] bg-cover bg-center animate-fade-in self-center relative flex flex-col justify-between overflow-hidden"
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
      :players="props.gameState?.players"
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
import { getCardInfoHelper } from "../../constants/cards";
import { ref, computed, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useGameStore } from "../../stores/gameStore";
import { storeToRefs } from "pinia";
import BoardHeader from "./BoardHeader.vue";
import BoardOpponents from "./BoardOpponents.vue";
import BoardTable from "./BoardTable.vue";
import GameLogPanel from "../../components/GameLogPanel.vue";
import ChancellorModal from "../ChancellorModal.vue";
import ActionModal from "../ActionModal.vue";
import CardRevealModal from "../CardRevealModal.vue";
import GameEndModal from "../GameEndModal.vue";
import GameErrorModal from "../GameErrorModal.vue";
import ConfirmModal from "../ConfirmModal.vue";
import deckBackImage from "../../assets/deckBack.png";
import chipImage from "../../assets/chip.png";

const props = defineProps({
  roomID: { type: String, required: true },
  gameState: { type: Object, required: true },
  myID: { type: String, default: "" },
  isStarting: { type: [Boolean, String], default: false },
});

const emit = defineEmits([
  "play-card",
  "start-game",
  "leave-game",
  "next-round",
  "restart-game",
  "add-bot",
]);

const { t } = useI18n();
const gameStore = useGameStore();

// Тільки стабільні (завжди присутні) поля беремо через storeToRefs.
const { revealedCardData, myID: storeMyID } = storeToRefs(gameStore);

/*
  Похідні поля дошки (handSlots, lastPlayedCardsByPlayer, discardSequence,
  secondsLeft) живуть у сторі, але ми НЕ деструктуримо їх через storeToRefs.
  Причина: під час Vite HMR модуль компонента може перезавантажитись
  раніше ніж модуль стора, тимчасово даючи undefined для новододаних
  ref'ів - це валило render із "Cannot read properties of undefined".
  Замість цього робимо тонкі computed-обгортки з optional-chaining;
  Pinia сама забезпечує реактивність на доступ до власних полів.
*/
const lastPlayedCardsByPlayer = computed(() => gameStore.lastPlayedCardsByPlayer ?? {});
const discardSequence = computed(() => gameStore.discardSequence ?? []);
const localSecondsLeft = computed(() => gameStore.secondsLeft ?? 0);
const latestTurnAlert = computed(() => gameStore.latestTurnAlert);

const showLeaveConfirm = ref(false);
const showActionModal = ref(false);
const isBotSubmitting = ref(false);
const activePlay = ref({ cardType: "", handIndex: 0, targetID: "", guessCard: "" });

const effectiveMyID = computed(() => {
  const fromStore = storeMyID.value;
  if (typeof fromStore === "string" && fromStore.length > 0) return fromStore;
  return typeof props.myID === "string" ? props.myID : "";
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

// myHandCards залишаємо як зручний доступ для логіки (наприклад,
// showChancellorPanel перевіряє довжину). Стабільні UID-и беремо
// з myHandSlots зі стору.
const myHandCards = computed(() => {
  const me = myPlayer.value;
  if (me && Array.isArray(me.hand)) return me.hand;
  const fallback = props.gameState?.my_hand;
  return Array.isArray(fallback) ? fallback : [];
});

// Створення стабільних слотів карт для усунення миготіння у v-for
const localHandSlots = computed(() => {
  return myHandCards.value.map((card, idx) => ({
    uid: `hand-slot-${idx}-${card}`,
    index: idx,
    cardType: card,
  }));
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

/*
  Уся логіка discard-tracking, lastPlayedCardsByPlayer та локального
  тікера секунд тепер живе у Pinia сторі (gameStore._updateBoardDerived
  та startLocalTimer). Компонент - чистий споживач.

  Що ми тут НЕ робимо більше:
    - watch({ deep: true }) по props.gameState (re-render storm)
    - setInterval що мутує gameStore.gameState.seconds_left
      (feedback-loop з merge)
    - локальний скид по round_number / deck.length
      (стор сам це зробить у обробнику ROOM_UPDATED)
    - onUnmounted з clearInterval (нема локального таймера)
*/

const handleCloseRevealModal = () => gameStore.clearRevealedData();
const handleClearError = () => gameStore.clearError();

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

// discardSequence приходить зі стору вже як [{seq, type, playerId}].
// BoardTable читає лише seq + type, тому додаткового мепу не потрібно.
// owner резолвимо ліниво у тулі/дев-консолі через arrangedPlayers, якщо
// колись знадобиться, без ремепу всього масиву на кожен render.
const globalDiscardPile = computed(() => discardSequence.value);

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

const isTurnOfPlayer = (id) => props.gameState?.current_player_id === id;

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
    target_id: targetID || "",
    guess_card: guessCardId ? Number(guessCardId) : 0,
  };
  emit("play-card", actionPayload);
};

const handleNextRoundRequest = () => {
  emit("next-round");
};
const handleRestartGameRequest = () => {
  emit("restart-game");
};

const handleDelayAndAddBot = async () => {
  if (isBotSubmitting.value) return;
  isBotSubmitting.value = true;
  try {
    emit("add-bot");
  } finally {
    setTimeout(() => {
      isBotSubmitting.value = false;
    }, 600);
  }
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
    bottom_order: bottomOrder,
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

/* Стилі та анімації для банера останнього ходу */
.slide-banner-enter-from {
  opacity: 0;
  transform: translateY(-40px) scale(0.95);
}
.slide-banner-leave-to {
  opacity: 0;
  transform: translateY(20px) scale(0.9);
}
.slide-banner-enter-active,
.slide-banner-leave-active {
  transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes pulseSubtle {
  0%,
  100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.02);
    shadow: 0 0 40px rgba(245, 158, 11, 0.6);
  }
}
.animate-pulse-subtle {
  animation: pulseSubtle 2s infinite ease-in-out;
}
</style>
