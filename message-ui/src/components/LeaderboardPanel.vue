<template>
  <div
    class="bg-brand-bg border border-slate-800 rounded-xl p-4 h-full flex flex-col shadow-lg"
  >
    <div class="flex items-center gap-2 mb-4 pb-2 border-b border-slate-800/60">
      <h3 class="font-mono font-bold text-md text-amber-500 uppercase tracking-wider">
        {{ $t("desktop.leaderboard.title", "Top Players") }}
      </h3>
    </div>

    <div
      v-if="loading"
      class="flex-1 flex items-center justify-center py-12 text-slate-500 font-mono text-xs animate-pulse"
    >
      {{ $t("desktop.leaderboard.loading", "Syncing database...") }}
    </div>

    <div
      v-else-if="error"
      class="text-rose-400 text-xs font-mono p-2 rounded bg-rose-500/5 border border-rose-500/10 text-center my-4"
    >
      {{ error }}
    </div>

    <div v-else class="flex-1 overflow-y-auto space-y-2 pr-1 custom-scrollbar">
      <div
        v-for="(player, index) in leaders"
        :key="player.username"
        class="flex items-center justify-between p-2 rounded-lg transition-all"
        :class="[
          player.username === currentUsername
            ? 'bg-amber-500/10 border border-amber-500/30'
            : 'bg-brand-bg-dark/40 border border-slate-800/40 hover:border-slate-700/60',
        ]"
      >
        <div class="flex items-center gap-2.5 min-w-0">
          <span
            class="w-5 text-center font-mono font-black text-xs"
            :class="getRankClass(index)"
          >
            {{ index + 1 }}
          </span>

          <div
            class="w-16 h-16 rounded-full bg-brand-bg-dark border border-slate-700 overflow-hidden flex-shrink-0 shadow-inner"
          >
            <img
              :src="
                getAvatarUrl(player.avatar_seed || player.AvatarSeed || 'default_seed')
              "
              :alt="$t('desktop.myAvatarAlt')"
              class="w-full h-full object-cover rounded-full"
            />
          </div>

          <span
            class="font-mono text-sm truncate font-medium"
            :class="
              player.username === currentUsername ? 'text-amber-400' : 'text-slate-200'
            "
          >
            {{ player.username }}
          </span>
        </div>

        <div class="text-right flex-shrink-0 pl-2">
          <div class="font-mono text-xl font-bold text-amber-500">
            {{ player.total_score }}
          </div>
          <div class="text-xs text-brand-text-subtle font-mono">
            Wan:{{ player.games_won }} / Spy:{{ player.games_played - player.games_won }}
          </div>
        </div>
      </div>

      <div
        v-if="leaders.length === 0"
        class="text-center py-8 text-slate-600 font-mono text-xs"
      >
        {{ $t("desktop.leaderboard.empty", "No records found") }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useAuthStore } from "../stores/auth";
import { getAvatarUrl } from "../utils/avatar";

const props = defineProps({
  currentUsername: {
    type: String,
    required: true,
  },
});

const authStore = useAuthStore();
const leaders = ref([]);
const loading = ref(true);
const error = ref(null);

const getRankClass = (index) => {
  if (index === 0)
    return "text-amber-400 drop-shadow-[0_0_4px_rgba(245,158,11,0.5)] text-sm font-extrabold";
  if (index === 1) return "text-slate-300 text-sm";
  if (index === 2) return "text-amber-700 text-sm";
  return "text-slate-500";
};

const fetchLeaderboard = async () => {
  try {
    loading.value = true;
    error.value = null;
    const token = authStore.token || localStorage.getItem("token");

    const response = await fetch("/api/leaderboard?limit=10", {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
    });

    if (!response.ok) throw new Error("Failed to load global rating");

    const data = await response.json();
    leaders.value = Array.isArray(data) ? data : data.leaderboard || [];
  } catch (err) {
    error.value = err.message;
    console.error("Leaderboard error:", err);
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  fetchLeaderboard();
});
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #1e293b;
  border-radius: 2px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #334155;
}
</style>
