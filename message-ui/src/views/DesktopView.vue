<template>
  <div class="min-h-screen bg-brand-bg-dark text-white p-3 sm:p-6">
    <div class="max-w-6xl mx-auto">
      <div
        class="lobby-header flex flex-wrap items-center justify-between gap-3 bg-brand-bg border border-slate-800 p-3 sm:p-4 rounded-xl mb-6"
      >
        <div class="flex items-center gap-3">
          <div
            @click="openAvatarModal"
            class="relative w-16 h-16 bg-brand-bg-dark border-2 border-amber-500/80 rounded-full p-0.5 overflow-hidden shadow-md cursor-pointer group transition-transform hover:scale-105"
            :title="$t('desktop.profileTitle')"
          >
            <img
              :src="getAvatarUrl(userAvatarSeed)"
              :alt="$t('desktop.myAvatarAlt')"
              class="w-full h-full object-cover rounded-full"
            />
            <div
              class="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 flex items-center justify-center transition-opacity rounded-full"
            >
              <span
                class="text-[10px] text-amber-400 font-bold uppercase tracking-tighter"
              >
                {{ $t("desktop.avatarChangeHint") }}
              </span>
            </div>
          </div>

          <div>
            <h1 class="text-xl font-bold text-amber-500 font-mono leading-tight">
              {{ $t("desktop.brand") }}
            </h1>
            <p class="text-xm text-slate-400">
              {{ $t("desktop.welcome", { username: currentUsername }) }}
            </p>
          </div>
        </div>

        <div class="flex flex-wrap items-center justify-end gap-2 sm:gap-3 w-full sm:w-auto">
          <button
            @click="showRulesModal = true"
            class="btn-ghost flex items-center gap-1.5 text-slate-300 hover:text-amber-400 transition-colors"
          >
            {{ $t("desktop.rulesButton") }}
          </button>
          <router-link v-if="authStore.isAdmin" to="/admin" class="btn-admin">
            {{ $t("desktop.adminLink") }}
          </router-link>
          <button @click="logout" class="btn-ghost">{{ $t("desktop.logout") }}</button>
          <button @click="createRoom" class="btn-accent">
            {{ $t("desktop.createRoom") }}
          </button>
        </div>
      </div>

      <div
        v-if="apiError"
        class="p-4 mb-6 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-400 text-sm"
      >
        {{ apiError }}
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
        <div class="lg:col-span-2">
          <h2 class="text-lg font-bold mb-4 font-mono text-brand-text-subtle">
            {{ $t("desktop.lobbyHeader") }}
          </h2>

          <div
            v-if="gameStore.lobbyRooms && gameStore.lobbyRooms.length > 0"
            class="grid grid-cols-1 md:grid-cols-2 gap-4"
          >
            <div
              v-for="room in gameStore.lobbyRooms"
              :key="room.room_id || room.ID"
              class="lobby-card flex items-center justify-between p-4 bg-brand-bg border border-slate-800 rounded-xl hover:border-brand-border transition-all"
            >
              <div class="flex items-center gap-3 truncate">
                <div
                  class="w-12 h-12 rounded-full border-2 border-amber-500/40 p-0.5 bg-brand-bg-dark overflow-hidden flex-shrink-0 shadow-md"
                >
                  <img
                    :src="
                      getAvatarUrl(
                        room.host_avatar_seed || room.HostAvatarSeed || 'default_seed'
                      )
                    "
                    :alt="$t('desktop.hostAvatarAlt')"
                    class="w-full h-full object-cover rounded-full"
                  />
                </div>

                <div class="truncate">
                  <div
                    class="font-bold text-amber-500/90 font-mono text-sm flex items-center gap-1.5"
                  >
                    {{ $t("desktop.roomLabel", { id: room.room_id || room.ID }) }}
                  </div>
                  <div class="text-xs text-slate-400 mt-0.5">
                    {{
                      $t("desktop.playersCount", {
                        count: room.player_count ?? room.PlayerCount ?? 0,
                        max: room.max_players ?? room.MaxPlayers ?? 4,
                      })
                    }}
                  </div>
                  <div
                    v-if="room.player_names || room.PlayerNames"
                    class="text-[10px] text-slate-500 mt-0.5 truncate max-w-[200px]"
                    :title="(room.PlayerNames || room.player_names).join(',')"
                  >
                    {{
                      $t("desktop.participants", {
                        names: (room.PlayerNames || room.player_names).join(","),
                      })
                    }}
                  </div>
                </div>
              </div>

              <router-link
                :to="`/room/${room.room_id || room.ID}`"
                class="btn-enter flex-shrink-0"
              >
                {{ $t("desktop.join") }}
              </router-link>
            </div>
          </div>

          <div v-else class="lobby-empty">
            {{ $t("desktop.noRooms") }}
          </div>
        </div>

        <div class="lg:col-span-1">
          <h2
            class="text-lg font-bold mb-4 font-mono text-slate-400 invisible hidden lg:block select-none"
          >
            -
          </h2>
          <LeaderboardPanel :currentUsername="currentUsername" />
        </div>
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

    <GameRulesModal v-if="showRulesModal" @close="showRulesModal = false" />
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { useAuthStore } from "../stores/auth";
import { useGameStore } from "../stores/gameStore";
import { getAvatarUrl } from "../utils/avatar";
import UserProfileModal from "./UserProfileModal.vue";
import LeaderboardPanel from "../components/LeaderboardPanel.vue";
import GameRulesModal from "./GameRulesModal.vue";

const { t } = useI18n();
const router = useRouter();
const authStore = useAuthStore();
const gameStore = useGameStore();

const apiError = ref(null);
const currentUsername = ref(
  localStorage.getItem("username") || authStore.user?.username || t("common.player")
);
const showAvatarModal = ref(false);
const showRulesModal = ref(false);
const userAvatarSeed = ref(
  localStorage.getItem("avatar_seed") || authStore.user?.avatar_seed || "default_seed"
);

const generateRandomSeed = () => {
  return (
    Math.random().toString(36).substring(2, 15) +
    Math.random().toString(36).substring(2, 15)
  );
};

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
    if (response.status === 401) throw new Error(t("desktop.errors.sessionExpired"));
    if (!response.ok) throw new Error(t("desktop.errors.loadRoomsFailed"));
    const data = await response.json();
    gameStore.lobbyRooms = Array.isArray(data) ? data : data.rooms || [];
  } catch (err) {
    apiError.value = err.message;
  }
};

const logout = () => {
  if (typeof authStore.clearSession === "function") authStore.clearSession();
  if (gameStore && typeof gameStore.disconnect === "function") gameStore.disconnect();
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
    if (response.status === 401) throw new Error(t("desktop.errors.noRightsToCreate"));
    if (!response.ok) throw new Error(t("desktop.errors.createFailed"));
    const newRoom = await response.json();
    const actualRoomId = newRoom.room_id || newRoom.RoomID || newRoom.id;
    if (actualRoomId) router.push(`/room/${actualRoomId}`);
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
  }
});
</script>
