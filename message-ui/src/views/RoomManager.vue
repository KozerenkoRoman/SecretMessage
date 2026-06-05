<template>
  <div class="room-manager-container min-h-screen bg-slate-900">
    <div
      v-if="loading || !gameState"
      class="flex h-screen items-center justify-center text-white"
    >
      <div class="text-center">
        <div
          class="animate-spin rounded-full h-12 w-12 border-b-2 border-cyan-400 mx-auto mb-4"
        ></div>
        <p class="text-slate-400">Підключення до ігрової кімнати {{ roomID }}...</p>
      </div>
    </div>

    <BoardView
      v-else
      :roomID="roomID"
      :gameState="gameState"
      :myID="myID"
      @play-card="handlePlayCard"
      @start-game="handleStartGameSignal"
      @leave-game="handleLeaveRoom"
      @next-round="triggerNextRound"
      @restart-game="triggerRestartGame"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useGameStore } from "../stores/gameStore";
import { useAuthStore } from "../stores/auth";
import BoardView from "./BoardView.vue";

const gameStore = useGameStore();
const authStore = useAuthStore();

const route = useRoute(); // Поточний стан роуту (параметри)
const router = useRouter(); // Інструмент для навігації (push/replace)
const roomID = computed(() => route.params.id); // 1. ВИПРАВЛЕНО: route замість router
const loading = ref(true);
const gameState = computed(() => gameStore.gameState);
const myID = computed(() => {
  return authStore.user?.id || localStorage.getItem("user_id") || "";
});

watch(
  () => gameStore.gameState,
  (newState) => {
    if (newState) {
      loading.value = false;
    }
  },
  { immediate: true }
);

onMounted(() => {
  if (roomID.value) {
    loading.value = true;
    gameStore.connectToHub(roomID.value);
  }
});

const triggerNextRound = () => {
  console.log("[RoomManager] Надсилаємо запит NEXT_ROUND на бекенд...");
  gameStore.sendWSMessage("NEXT_ROUND", null, null, null);
};

const triggerRestartGame = () => {
  console.log("[RoomManager] Надсилаємо запит START_GAME для перезапуску сесії...");
  gameStore.sendWSMessage("START_GAME", null, null, null);
};

const handleStartGameSignal = () => {
  console.log("[RoomManager] Надсилаємо сигнал START_GAME на сервер...");
  gameStore.sendWSMessage("START_GAME", null, null, null);
};

// Обробка виходу з кімнати
const handleLeaveRoom = () => {
  console.log("[RoomManager] Гравець виходить з кімнати...");
  // Явно викликаємо метод стору для відключення перед зміною сторінки
  gameStore.leaveCurrentRoom();
  router.push("/desktop");
};

// Обробка розіграшу карти
const handlePlayCard = (actionPayload) => {
  // Надсилаємо хід на сервер
  gameStore.sendWSMessage("ACTION", null, actionPayload, null);
};

onBeforeUnmount(() => {
  // Запобіжний вихід з кімнати, якщо компонент розмонтується іншим шляхом
  gameStore.leaveCurrentRoom();
});
</script>
