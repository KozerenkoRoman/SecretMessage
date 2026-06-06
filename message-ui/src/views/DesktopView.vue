<template>
  <div class="min-h-screen bg-slate-950 text-white p-6">
    <div class="max-w-4xl mx-auto">
      <div
        class="lobby-header flex items-center justify-between bg-slate-900 border border-slate-800 p-4 rounded-xl mb-6"
      >
        <div class="flex items-center gap-3">
          <div
            @click="openAvatarModal"
            class="relative w-12 h-12 bg-slate-950 border-2 border-amber-500/80 rounded-full p-0.5 overflow-hidden shadow-md cursor-pointer group transition-transform hover:scale-105"
            title="Налаштування профілю"
          >
            <img
              :src="getAvatarUrl(userAvatarSeed)"
              alt="Мій аватар"
              class="w-full h-full object-cover rounded-full"
            />
            <div
              class="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 flex items-center justify-center transition-opacity rounded-full"
            >
              <span
                class="text-[10px] text-amber-400 font-bold uppercase tracking-tighter"
                >Змінити</span
              >
            </div>
          </div>

          <div>
            <h1 class="text-xl font-bold text-amber-500 font-mono leading-tight">
              Secret Message
            </h1>
            <p class="text-xs text-slate-400">Вітаємо, {{ currentUsername }}</p>
          </div>
        </div>

        <div class="flex items-center gap-3">
          <router-link v-if="authStore.isAdmin" to="/admin" class="btn-admin">
            Панель адміна ⚙
          </router-link>

          <button @click="logout" class="btn-ghost">Вийти</button>

          <button @click="createRoom" class="btn-accent">+ Створити кімнату</button>
        </div>
      </div>

      <div
        v-if="apiError"
        class="p-4 mb-6 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-400 text-sm"
      >
        {{ apiError }}
      </div>

      <h2 class="text-lg font-bold mb-4 font-mono text-slate-300">
        Доступні лобі (в реальному часі):
      </h2>

      <div
        v-if="gameStore.lobbyRooms && gameStore.lobbyRooms.length > 0"
        class="grid grid-cols-1 md:grid-cols-2 gap-4"
      >
        <div
          v-for="room in gameStore.lobbyRooms"
          :key="room.room_id || room.ID"
          class="lobby-card"
        >
          <div>
            <div class="font-bold text-amber-500/90">
              Кімната #{{ room.room_id || room.ID }}
            </div>
            <div class="text-xs text-slate-400 mt-1">
              Гравців: {{ room.player_count ?? room.PlayerCount ?? 0 }}/{{
                room.max_players ?? room.MaxPlayers ?? 4
              }}
            </div>
            <div
              v-if="room.player_names || room.PlayerNames"
              class="text-[10px] text-slate-500 mt-1"
            >
              Учасники: {{ (room.PlayerNames || room.player_names).join(", ") }}
            </div>
          </div>
          <router-link :to="`/room/${room.room_id || room.ID}`" class="btn-enter">
            Увійти
          </router-link>
        </div>
      </div>

      <div v-else class="lobby-empty">
        Активних кімнат немає або вони вже розпочали гру. Створіть першу!
      </div>
    </div>

    <UserProfileModal
      v-if="showAvatarModal"
      :currentUsername="currentUsername"
      :currentSeed="userAvatarSeed"
      :apiUrl="''"
      @close="showAvatarModal = false"
      @updated="handleProfileUpdated"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "../stores/auth";
import { useGameStore } from "../stores/gameStore";
import { getAvatarUrl } from "../utils/avatar";
import UserProfileModal from "./UserProfileModal.vue";

const router = useRouter();
const authStore = useAuthStore();
const gameStore = useGameStore();
const apiError = ref(null);

// Перетворюємо у ref, щоб воно миттєво реагувало на зміни профілю з модалки
const currentUsername = ref(
  localStorage.getItem("username") || authStore.user?.username || "Гравець"
);

// Стан для керування відображенням модалки
const showAvatarModal = ref(false);
const userAvatarSeed = ref(
  localStorage.getItem("avatar_seed") || authStore.user?.avatar_seed || "default_seed"
);

// Функція для генерації випадкового сиду
const generateRandomSeed = () => {
  return (
    Math.random().toString(36).substring(2, 15) +
    Math.random().toString(36).substring(2, 15)
  );
};

// Відкриття вікна редагування
const openAvatarModal = () => {
  showAvatarModal.value = true;
};

const handleProfileUpdated = (updatedData) => {
  userAvatarSeed.value = updatedData.avatar_seed;
  currentUsername.value = updatedData.username;

  if (authStore.user) {
    authStore.user.username = updatedData.username;
    authStore.user.avatar_seed = updatedData.avatar_seed;
  }
};

const fetchRooms = async () => {
  try {
    apiError.value = null;
    const token = authStore.token || localStorage.getItem("token");
    const response = await fetch("/api/rooms", {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
    });

    if (response.status === 401) {
      throw new Error("Сесія застаріла або неавторизована. Будь ласка, перезайдіть.");
    }
    if (!response.ok) {
      throw new Error("Не вдалося завантажити список кімнат");
    }

    const data = await response.json();
    gameStore.lobbyRooms = Array.isArray(data) ? data : data.rooms || [];
  } catch (err) {
    apiError.value = err.message;
    console.error("Помилка завантаження кімнат:", err);
  }
};

const logout = () => {
  if (typeof authStore.clearSession === "function") {
    authStore.clearSession();
  }
  if (gameStore && typeof gameStore.disconnect === "function") {
    gameStore.disconnect();
  }
  router.push("/auth");
};

const createRoom = async () => {
  try {
    const token = authStore.token || localStorage.getItem("token");
    const response = await fetch("/api/rooms", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
    });

    if (response.status === 401) {
      throw new Error("Немає прав для створення кімнати (401)");
    }
    if (!response.ok) throw new Error("Не вдалося створити кімнату");

    const newRoom = await response.json();
    const actualRoomId = newRoom.room_id || newRoom.RoomID || newRoom.id;
    if (actualRoomId) {
      router.push(`/room/${actualRoomId}`);
    }
  } catch (err) {
    alert(err.message);
  }
};

onMounted(() => {
  fetchRooms();
  gameStore.connectToHub();
  const storedSeed = localStorage.getItem("avatar_seed") || authStore.user?.avatar_seed;
  if (storedSeed && storedSeed !== "default_seed") {
    userAvatarSeed.value = storedSeed;
  } else {
    const newSeed = generateRandomSeed();
    userAvatarSeed.value = newSeed;
    localStorage.setItem("avatar_seed", newSeed);

    // Тут в ідеалі зробити швидкий запит на бекенд (POST /api/user),
    // щоб назавжди зберегти цей згенерований сід у базу даних для цього юзера.
  }
});
</script>
