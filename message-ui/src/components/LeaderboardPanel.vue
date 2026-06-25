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
        class="relative flex items-start justify-between p-3 pb-8 rounded-lg transition-all gap-4 min-w-0"
        :class="[
          player.username === currentUsername
            ? 'bg-amber-500/10 border border-amber-500/30'
            : 'bg-brand-bg-dark/40 border border-slate-800/40 hover:border-slate-700/60',
        ]"
      >
        <div class="flex items-start gap-3 min-w-0 flex-1">
          <span
            class="w-5 text-center font-mono font-black text-xs flex-shrink-0 mt-[14px]"
            :class="getRankClass(index)"
          >
            {{ index + 1 }}
          </span>

          <div
            class="w-12 h-12 rounded-full bg-brand-bg-dark border border-slate-700 overflow-hidden flex-shrink-0 shadow-inner"
          >
            <img
              :src="getAvatarUrl(player.avatar_seed || 'default_seed')"
              :alt="$t('desktop.myAvatarAlt')"
              class="w-full h-full object-cover rounded-full"
            />
          </div>

          <div class="min-w-0 flex-1 pt-0.5">
            <span
              class="font-mono text-xm font-medium text-left block truncate leading-none"
              :class="
                player.username === currentUsername ? 'text-amber-400' : 'text-slate-200'
              "
              :title="player.username"
            >
              {{ player.username }}
            </span>
          </div>
        </div>

        <div class="text-right flex-shrink-0 pt-0.5">
          <div class="font-mono text-xl font-bold text-amber-500 leading-none">
            {{ player.user_stat?.total_score ?? 0 }}
          </div>
        </div>

        <div
          class="absolute bottom-2.5 right-3 text-xm text-brand-text-subtle font-mono whitespace-nowrap opacity-80"
        >
          Total:{{ player.user_stat?.games_played ?? 0 }} · Won:{{
            player.user_stat?.games_won ?? 0
          }}
          · Spy:{{ player.user_stat?.spy_bonuses ?? 0 }}
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

defineProps({
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
    return "text-amber-400 drop-shadow-[0_0_4px_var(--color-brand-accent)] text-sm font-extrabold";
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
    leaders.value = data && Array.isArray(data) ? data : data?.leaderboard || [];
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
