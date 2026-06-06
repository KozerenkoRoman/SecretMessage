<template>
  <div
    class="h-screen bg-slate-900 text-white p-2 flex flex-col justify-between overflow-hidden relative font-sans select-none"
  >
    <header
      class="w-full flex justify-between items-center bg-slate-800/80 backdrop-blur px-3 py-2 rounded-xl border border-slate-700 flex-shrink-0 gap-4 z-10 h-[6vh]"
    >
      <div class="flex items-center gap-4">
        <button
          @click="handleLeaveGame"
          type="button"
          class="px-4 py-1.5 bg-rose-950/40 hover:bg-rose-900/60 text-rose-400 font-bold rounded-xl border border-rose-900/50 shadow-md active:scale-95 transition-all text-xs uppercase tracking-wider font-mono cursor-pointer"
        >
          Вийти
        </button>
        <div class="h-6 w-[1px] bg-slate-700"></div>
        <div>
          <h2 class="text-sm font-bold text-yellow-400 font-mono leading-none mb-0.5">
            Кімната: {{ roomID }}
          </h2>
          <p class="text-[10px] text-slate-400">
            Час:
            <span class="text-amber-400 font-mono font-bold"
              >{{ gameState?.seconds_left || 0 }}с</span
            >
          </p>
        </div>
      </div>
      <div class="flex items-center gap-3">
        <div
          v-if="!gameState?.is_started"
          class="status-pill bg-blue-500/20 text-blue-400 border-blue-500/40 animate-pulse"
        >
          Очікування...
        </div>
        <button
          v-if="canStartGame"
          @click="handleStartGame"
          class="px-4 py-1.5 bg-gradient-to-r from-amber-500 to-orange-600 hover:from-amber-600 hover:to-orange-700 text-slate-950 font-black rounded-xl shadow-lg shadow-amber-950/20 active:scale-95 transition-all text-xs uppercase tracking-wider font-mono cursor-pointer"
        >
          Почати ⚔
        </button>
        <div
          v-else-if="showChancellorPanel"
          class="status-pill bg-gradient-to-r from-amber-500/20 to-orange-500/20 text-amber-400 border-yellow-500/40 animate-pulse"
        >
          Вибір Канцлера!
        </div>
        <div
          v-else-if="isMyTurn"
          class="status-pill bg-emerald-500/20 text-emerald-400 border-emerald-500/40 animate-pulse"
        >
          Ваш хід!
        </div>
        <div v-else class="status-pill bg-slate-700 text-slate-300 border-transparent">
          Ходить: {{ currentTurnPlayerName }}
        </div>
      </div>
    </header>

    <main
      class="flex-1 flex flex-col justify-between my-1 gap-1 overflow-hidden w-full max-w-6xl mx-auto"
    >
      <div
        class="w-full flex justify-center items-center gap-6 flex-shrink-0 h-[20vh] max-h-[140px] overflow-hidden"
      >
        <template v-for="(player, index) in opponents" :key="player.id">
          <div
            :class="[
              isTurnOfPlayer(player.id)
                ? 'border-yellow-400 bg-slate-800 ring-2 ring-yellow-400/30'
                : 'border-slate-700 bg-slate-800/60',
            ]"
            class="flex flex-col items-center p-2 rounded-xl border transition-all duration-300 w-60 h-full justify-between shadow-md"
          >
            <div
              class="flex items-center justify-between w-full border-b border-slate-700 pb-0.5 flex-shrink-0"
            >
              <span
                class="font-bold text-[11px] truncate max-w-[130px] text-slate-200"
                :title="player.username"
                >{{ player.username || "Опонент" }}</span
              >
              <span
                class="text-[9px] bg-slate-700 text-yellow-400 px-1 py-0.5 rounded font-mono"
                >★ {{ player.score || 0 }}</span
              >
            </div>
            <div
              class="flex justify-center items-center h-16 w-full overflow-hidden pl-2"
            >
              <div
                v-for="cIdx in getOpponentHandCount(player)"
                :key="cIdx"
                class="w-10 h-14 bg-gradient-to-br from-indigo-950 to-slate-900 rounded-lg border border-indigo-400/40 shadow flex items-center justify-center -mr-2"
              >
                <span class="text-[8px] font-mono text-indigo-400/50">L.H.</span>
              </div>
            </div>
            <div class="w-full text-center text-[9px] flex-shrink-0">
              <span
                v-if="player.is_protected"
                class="text-yellow-400 font-bold uppercase text-[8px]"
                >Захист</span
              >
              <span
                v-else-if="player.is_out"
                class="text-rose-400 font-bold uppercase text-[8px]"
                >Вибув</span
              >
              <span v-else class="text-slate-500 italic">У грі</span>
            </div>
          </div>
        </template>
      </div>

      <div
        class="h-[50vh] flex items-center justify-center p-3 bg-slate-950/40 rounded-2xl border border-slate-800/60 w-full overflow-hidden"
      >
        <div
          class="flex gap-6 items-center w-full max-w-5xl mx-auto justify-between h-full"
        >
          <div class="flex flex-col items-center justify-center gap-1 flex-shrink-0">
            <div
              class="relative w-44 h-64 bg-gradient-to-br from-yellow-800 to-indigo-900 rounded-xl border-2 border-yellow-500/40 shadow-lg shadow-amber-950/20 flex flex-col items-center justify-center bg-cover bg-center overflow-hidden"
              :style="{ backgroundImage: `url(${deckBackImage})` }"
            >
              <div
                class="absolute inset-0 bg-slate-950/50 flex flex-col items-center justify-center p-1"
              >
                <span
                  class="text-[12px] font-bold uppercase text-amber-200 tracking-wider bg-slate-950/80 px-2 py-0.5 rounded backdrop-blur-[1px]"
                  >Колода</span
                >
                <span
                  class="text-2xl font-black text-white leading-none mt-1.5 bg-slate-950/70 px-2.5 py-1 rounded font-mono shadow-md border border-white/5"
                  >{{ gameState?.deck ? gameState.deck.length : 0 }}</span
                >
              </div>
              <div
                class="absolute inset-0 border border-yellow-500/20 rounded-xl translate-x-[2px] translate-y-[2px] -z-10 bg-slate-900"
              ></div>
            </div>
            <div class="text-[16px] font-bold text-slate-400 font-mono mt-0.5">
              Відбій: {{ globalDiscardPile.length }}
            </div>
          </div>
          <div class="h-full max-h-[220px] w-[1px] bg-slate-800 flex-shrink-0"></div>
          <div class="flex-1 flex flex-col justify-between pl-1 h-full overflow-hidden">
            <div
              class="flex justify-between items-center border-b border-slate-800 pb-1 mb-1 flex-shrink-0"
            >
              <span
                class="text-[10px] uppercase text-amber-400 font-bold font-mono tracking-wider"
                >Стіл відбою</span
              >
            </div>
            <div
              class="flex flex-wrap gap-2 justify-start w-full overflow-y-auto overflow-x-hidden custom-scrollbar content-start flex-1 px-3 pt-2 pb-1"
            >
              <template
                v-for="slotIndex in globalDiscardPile.length"
                :key="`discard-slot-${slotIndex}`"
              >
                <div
                  v-if="globalDiscardPile[slotIndex - 1]"
                  class="game-card game-card-discard w-[5.5rem] aspect-[2/3] bg-cover bg-center transition-all duration-200 hover:scale-105 hover:z-20 flex-shrink-0 relative"
                  :class="getCardColor(globalDiscardPile[slotIndex - 1].type)"
                  :style="
                    getCardImage(globalDiscardPile[slotIndex - 1].type)
                      ? {
                          backgroundImage: `url(${getCardImage(
                            globalDiscardPile[slotIndex - 1].type
                          )})`,
                        }
                      : {}
                  "
                ></div>
              </template>
            </div>
          </div>
        </div>
      </div>

      <div class="w-full flex justify-center items-center flex-shrink-0 h-[4vh]">
        <div
          class="bg-slate-800/50 px-4 py-0.5 rounded-full border border-slate-700/60 flex items-center gap-3 text-xs"
        >
          <span class="font-bold text-yellow-400"
            >Ви: <span class="text-white">{{ myPlayer?.username || "Гість" }}</span></span
          >
          <span class="text-slate-500">|</span>
          <span class="text-amber-400 font-mono"
            >Перемоги: {{ myPlayer?.score || 0 }}★</span
          >
          <span class="text-slate-500">|</span>
          <span
            v-if="myPlayer?.is_protected"
            class="text-yellow-400 font-bold uppercase text-[9px]"
            >Під захистом</span
          >
          <span
            v-else-if="myPlayer?.is_out"
            class="text-rose-400 font-bold uppercase text-[9px]"
            >Ви вибули</span
          >
          <span v-else class="text-emerald-400 text-[11px]">У грі</span>
        </div>
      </div>
    </main>

    <footer
      class="w-full max-w-2xl mx-auto bg-slate-950/90 backdrop-blur-md p-2 rounded-t-2xl border-t border-x border-slate-800 flex flex-col items-center shadow-2xl flex-shrink-0 z-10 h-[28vh] min-h-[280px]"
    >
      <div
        v-if="showChancellorPanel"
        class="w-full flex flex-col items-center h-full justify-between"
      >
        <div class="text-center">
          <h4 class="text-amber-400 font-bold text-xs uppercase leading-none">
            🔮 Оберіть карту собі
          </h4>
        </div>
        <div class="flex justify-center gap-4 items-center flex-1 w-full overflow-hidden">
          <div
            v-for="(cardType, cIdx) in myHandCards"
            :key="`chancellor-card-${cIdx}-${cardType}`"
            @click="handleChancellorSelect(cIdx)"
            class="game-card w-44 h-64 cursor-pointer ring-2 ring-amber-500/50 bg-cover bg-center transition-all duration-200 hover:scale-105"
            :class="getCardColor(cardType)"
            :style="
              getCardImage(cardType)
                ? { backgroundImage: `url(${getCardImage(cardType)})` }
                : {}
            "
            :data-tooltip="`${getCardName(cardType)}(${getCardValue(
              cardType
            )}) — ${getCardDesc(cardType)}`"
          ></div>
        </div>
      </div>
      <div
        v-else
        class="flex justify-center gap-4 relative z-10 items-center flex-1 w-full h-full"
      >
        <div v-if="myHandCards.length === 0" class="text-slate-500 text-xs italic">
          Очікування карт...
        </div>
        <div
          v-for="(cardType, cIdx) in myHandCards"
          :key="`hand-card-${cIdx}-${cardType}`"
          @click="isMyTurn ? handleCardClick(cardType, cIdx) : null"
          class="game-card w-44 h-64 bg-cover bg-center transition-all duration-200"
          :class="[
            getCardColor(cardType),
            isMyTurn
              ? 'game-card-playable hover:-translate-y-4 hover:scale-105 cursor-pointer shadow-lg shadow-yellow-500/10'
              : 'game-card-disabled opacity-60 cursor-not-allowed',
          ]"
          :style="
            getCardImage(cardType)
              ? { backgroundImage: `url(${getCardImage(cardType)})` }
              : {}
          "
          :data-tooltip="`${getCardName(cardType)}(${getCardValue(
            cardType
          )}) — ${getCardDesc(cardType)}`"
        ></div>
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
      v-if="revealedCardData"
      :data="revealedCardData"
      :players="props.gameState.players"
      @close="handleCloseRevealModal"
    />
    <ChancellorModal
      :is-open="showChancellorPanel"
      :cards="myHandCards"
      @submit="
        ({ keepHandIndex, bottomOrder }) =>
          handleChancellorSelect({ keepIndex: keepHandIndex, bottomOrder })
      "
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
  </div>
</template>

<script setup>
import { CARD_INFO } from "../constants/cards";
import { ref, computed, watch } from "vue";
import { useGameStore } from "../stores/gameStore";
import { storeToRefs } from "pinia";
import ChancellorModal from "./ChancellorModal.vue";
import ActionModal from "./ActionModal.vue";
import CardRevealModal from "./CardRevealModal.vue";
import GameEndModal from "./GameEndModal.vue";
import GameErrorModal from "./GameErrorModal.vue";
import deckBackImage from "../assets/deckBack.png";

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
const gameStore = useGameStore();
const { revealedCardData, myID: storeMyID } = storeToRefs(gameStore);

watch(
  () => props.gameState?.round_number,
  (newRound, oldRound) => {
    if (
      typeof newRound === "number" &&
      typeof oldRound === "number" &&
      newRound > oldRound
    ) {
      console.log(
        `[BoardView] Раунд збільшився з ${oldRound} на ${newRound}. Закриваємо вікно дуелі.`
      );
      handleCloseRevealModal();
    }
  }
);

watch(
  () => props.gameState?.deck?.length,
  (newDeckLength, oldDeckLength) => {
    if (newDeckLength && oldDeckLength && newDeckLength > oldDeckLength) {
      console.log(
        "[BoardView] Колоду перетасовано для нового раунду. Закриваємо вікно дуелі."
      );
      handleCloseRevealModal();
    }
  }
);

const effectiveMyID = computed(() => {
  const fromStore = storeMyID.value;
  if (typeof fromStore === "string" && fromStore.length > 0) return fromStore;
  return typeof props.myID === "string" ? props.myID : "";
});

const handleCloseRevealModal = () => {
  gameStore.clearRevealedData();
};
const handleClearError = () => {
  gameStore.clearError();
};

const showActionModal = ref(false);
const handleStartGame = () => emit("start-game");

const getCardDesc = (type) => CARD_INFO[type]?.desc || "Опис відсутній";
const getCardColor = (type) => CARD_INFO[type]?.color || "bg-slate-700";
const getCardName = (type) => CARD_INFO[type]?.name || "Невідома карта";
const getCardValue = (type) =>
  CARD_INFO[type]?.value !== undefined ? CARD_INFO[type].value : "?";
const getCardImage = (type) => CARD_INFO[type]?.image || "";

const activePlay = ref({ cardType: "", handIndex: 0, targetID: "", guessCard: "" });

const globalDiscardPile = computed(() => {
  const playersData = props.gameState?.players;
  if (!playersData) return [];
  const allDiscards = [];
  if (typeof playersData === "object" && !Array.isArray(playersData)) {
    Object.keys(playersData).forEach((id) => {
      const p = playersData[id];
      if (p && Array.isArray(p.discard_pile)) {
        p.discard_pile.forEach((cardType) => {
          allDiscards.push({
            type: cardType,
            owner: p.username || "Гравець",
            playerId: id,
          });
        });
      }
    });
  } else if (Array.isArray(playersData)) {
    playersData.forEach((p) => {
      if (p && Array.isArray(p.discard_pile)) {
        p.discard_pile.forEach((cardType) => {
          allDiscards.push({
            type: cardType,
            owner: p.username || "Гравець",
            playerId: p.id,
          });
        });
      }
    });
  }
  return allDiscards;
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
  if (confirm("Ви впевнені, що хочете покинути поточну гру та повернутися в десктоп?")) {
    emit("leave-game");
  }
};

const arrangedPlayers = computed(() => {
  const playersData = props.gameState?.players;
  if (!playersData) return [];
  const list = Array.isArray(playersData)
    ? playersData
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
  return arrangedPlayers.value.find((p) => p.id === me) || null;
});

const myHandCards = computed(() => {
  const me = myPlayer.value;
  if (me && Array.isArray(me.hand)) return me.hand;
  const fallback = props.gameState?.my_hand;
  return Array.isArray(fallback) ? fallback : [];
});

const currentTurnPlayerName = computed(() => {
  const activeID = props.gameState?.current_player_id;
  if (!activeID) return "Опонент";
  const found = arrangedPlayers.value.find((p) => p.id === activeID);
  return found?.username || "Опонент";
});

const getOpponentHandCount = (player) => {
  if (!player) return 0;
  if (typeof player.hand_count === "number") return player.hand_count;
  if (Array.isArray(player.hand)) return player.hand.length;
  return 0;
};

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

const handleChancellorSelect = ({ keepIndex, bottomOrder }) => {
  const me = effectiveMyID.value;
  if (!me) return;
  const chancellorPayload = {
    player_id: me,
    keep_hand_index: Number(keepIndex),
    bottom_order: bottomOrder.map((card) => Number(card)),
  };
  gameStore.sendWSMessage("CHANCELLOR_RESOLVE", null, null, chancellorPayload);
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
  background: rgba(15, 23, 42, 0.3);
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(51, 65, 85, 0.8);
  border-radius: 4px;
}
</style>
