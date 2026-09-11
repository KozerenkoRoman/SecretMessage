<template>
  <div class="min-h-screen bg-brand-bg-dark text-white p-6">
    <div class="max-w-6xl mx-auto">
      <!-- Header -->
      <div
        class="flex justify-between items-center mb-8 bg-brand-bg p-4 rounded-xl border border-slate-800"
      >
        <div>
          <h1 class="text-2xl font-bold text-cyan-400 font-mono tracking-wide">
            {{ $t("simulation.title") }}
          </h1>
          <p class="text-xs text-slate-400">{{ $t("simulation.subtitle") }}</p>
        </div>
        <router-link
          to="/admin"
          class="px-4 py-2 bg-brand-surface hover:bg-brand-surface-dim text-amber-500 hover:text-amber-400 text-xs font-bold uppercase tracking-wider font-mono rounded-xl transition-all border border-brand-border shadow-md active:scale-95"
        >
          {{ $t("simulation.backToAdmin") }}
        </router-link>
      </div>

      <div
        v-if="errorMessage"
        class="p-4 mb-6 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-400 text-sm"
      >
        {{ errorMessage }}
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Config panel -->
        <div
          class="lg:col-span-1 bg-brand-bg border border-slate-800 rounded-xl p-5 shadow-xl h-fit"
        >
          <h2 class="font-mono text-brand-text-subtle font-bold mb-4">
            {{ $t("simulation.config.title") }}
          </h2>

          <label class="block text-xs text-slate-400 mb-1 font-mono uppercase">
            {{ $t("simulation.config.count") }}
          </label>
          <input
            v-model.number="form.count"
            type="number"
            min="1"
            max="5000"
            class="w-full mb-4 px-3 py-2 bg-brand-bg-dark border border-slate-700 rounded-lg text-sm focus:outline-none focus:border-cyan-500"
          />

          <label class="block text-xs text-slate-400 mb-1 font-mono uppercase">
            {{ $t("simulation.config.seed") }}
          </label>
          <input
            v-model.number="form.seed"
            type="number"
            min="0"
            class="w-full mb-4 px-3 py-2 bg-brand-bg-dark border border-slate-700 rounded-lg text-sm focus:outline-none focus:border-cyan-500"
          />

          <label class="block text-xs text-slate-400 mb-2 font-mono uppercase">
            {{ $t("simulation.config.slots") }}
          </label>
          <div
            v-for="(slot, idx) in form.slots"
            :key="idx"
            class="flex items-center gap-2 mb-2"
          >
            <span class="text-xs text-slate-500 w-6 font-mono">#{{ idx }}</span>
            <select
              v-model="slot.strategy"
              class="flex-1 px-2 py-1.5 bg-brand-bg-dark border border-slate-700 rounded-lg text-xs focus:outline-none focus:border-cyan-500"
            >
              <option value="smart">{{ $t("simulation.config.strategies.smart") }}</option>
              <option value="random">
                {{ $t("simulation.config.strategies.random") }}
              </option>
            </select>
            <button
              v-if="form.slots.length > 2"
              @click="removeSlot(idx)"
              class="text-rose-400 hover:text-rose-300 text-xs font-mono px-2"
              :title="$t('simulation.config.removeSlot')"
            >
              ✕
            </button>
          </div>
          <button
            v-if="form.slots.length < 4"
            @click="addSlot"
            class="text-cyan-400 hover:text-cyan-300 text-xs font-mono mb-4"
          >
            {{ $t("simulation.config.addSlot") }}
          </button>

          <label class="block text-xs text-slate-400 mb-1 mt-2 font-mono uppercase">
            {{ $t("simulation.config.firstMove") }}
          </label>
          <select
            v-model="form.firstMoveBotId"
            class="w-full mb-5 px-3 py-2 bg-brand-bg-dark border border-slate-700 rounded-lg text-sm focus:outline-none focus:border-cyan-500"
          >
            <option value="">{{ $t("simulation.config.firstMoveNone") }}</option>
            <option v-for="(slot, idx) in form.slots" :key="idx" :value="slotId(idx)">
              #{{ idx }} — {{ slot.strategy }}
            </option>
          </select>

          <button
            @click="launch"
            :disabled="isLaunching"
            class="w-full py-3 bg-gradient-to-r from-cyan-500 to-blue-600 hover:from-cyan-600 hover:to-blue-700 text-slate-950 font-black rounded-xl shadow-lg transition-all cursor-pointer text-sm uppercase tracking-wider active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ isLaunching ? $t("simulation.config.launching") : $t("simulation.config.launch") }}
          </button>
        </div>

        <!-- Progress + results -->
        <div class="lg:col-span-2 space-y-6">
          <!-- Progress -->
          <div class="bg-brand-bg border border-slate-800 rounded-xl p-5 shadow-xl">
            <div class="flex justify-between items-center mb-4">
              <h2 class="font-mono text-brand-text-subtle font-bold">
                {{ $t("simulation.progress.title") }}
              </h2>
              <button
                v-if="progress && progress.status === 'running'"
                @click="cancelBatch"
                class="px-3 py-1 bg-rose-950/40 hover:bg-rose-900/60 text-rose-400 text-[10px] font-bold font-mono rounded-lg border border-rose-900/50"
              >
                {{ $t("simulation.progress.cancel") }}
              </button>
            </div>

            <div v-if="!progress" class="text-slate-500 text-sm italic py-6 text-center">
              {{ $t("simulation.progress.idle") }}
            </div>

            <div v-else>
              <div class="flex items-center justify-between mb-2 text-sm">
                <span class="font-mono">
                  {{ $t("simulation.progress.status") }}:
                  <span :class="statusClass">{{ statusLabel }}</span>
                </span>
                <span class="font-mono text-slate-400">
                  {{ $t("simulation.progress.games") }}:
                  {{ progress.played_games }} / {{ progress.total_games }}
                </span>
              </div>

              <!-- Progress bar -->
              <div class="w-full h-3 bg-brand-bg-dark rounded-full overflow-hidden mb-4 border border-slate-800">
                <div
                  class="h-full bg-gradient-to-r from-cyan-500 to-blue-500 transition-all duration-300"
                  :style="{ width: progressPct + '%' }"
                ></div>
              </div>

              <!-- Summary stat tiles -->
              <div class="grid grid-cols-2 sm:grid-cols-3 gap-3 text-center">
                <div class="bg-brand-bg-dark/60 rounded-lg p-2 border border-slate-800">
                  <div class="text-lg font-mono font-bold text-cyan-400">
                    {{ fmt(summary.avg_turns) }}
                  </div>
                  <div class="text-[10px] text-slate-500 uppercase">
                    {{ $t("simulation.progress.avgTurns") }}
                  </div>
                </div>
                <div class="bg-brand-bg-dark/60 rounded-lg p-2 border border-slate-800">
                  <div class="text-lg font-mono font-bold text-cyan-400">
                    {{ fmt(summary.avg_rounds) }}
                  </div>
                  <div class="text-[10px] text-slate-500 uppercase">
                    {{ $t("simulation.progress.avgRounds") }}
                  </div>
                </div>
                <div class="bg-brand-bg-dark/60 rounded-lg p-2 border border-slate-800">
                  <div class="text-lg font-mono font-bold text-cyan-400">
                    {{ fmt(summary.avg_duration_ms) }}ms
                  </div>
                  <div class="text-[10px] text-slate-500 uppercase">
                    {{ $t("simulation.progress.avgDuration") }}
                  </div>
                </div>
                <div class="bg-brand-bg-dark/60 rounded-lg p-2 border border-slate-800">
                  <div class="text-lg font-mono font-bold text-slate-300">
                    {{ summary.draw_games || 0 }}
                  </div>
                  <div class="text-[10px] text-slate-500 uppercase">
                    {{ $t("simulation.progress.draws") }}
                  </div>
                </div>
                <div class="bg-brand-bg-dark/60 rounded-lg p-2 border border-slate-800">
                  <div class="text-lg font-mono font-bold text-amber-400">
                    {{ summary.total_invalid_moves || 0 }}
                  </div>
                  <div class="text-[10px] text-slate-500 uppercase">
                    {{ $t("simulation.progress.invalid") }}
                  </div>
                </div>
                <div class="bg-brand-bg-dark/60 rounded-lg p-2 border border-slate-800">
                  <div class="text-lg font-mono font-bold text-amber-400">
                    {{ summary.total_fallback_moves || 0 }}
                  </div>
                  <div class="text-[10px] text-slate-500 uppercase">
                    {{ $t("simulation.progress.fallback") }}
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Win rate chart -->
          <div
            v-if="strategyList.length"
            class="bg-brand-bg border border-slate-800 rounded-xl p-5 shadow-xl"
          >
            <h2 class="font-mono text-brand-text-subtle font-bold mb-4">
              {{ $t("simulation.results.winRates") }}
            </h2>
            <div v-for="st in strategyList" :key="st.strategy" class="mb-3">
              <div class="flex justify-between text-xs mb-1 font-mono">
                <span class="text-slate-300">{{ st.strategy }}</span>
                <span class="text-cyan-400">{{ pct(st.win_rate) }}%</span>
              </div>
              <div class="w-full h-4 bg-brand-bg-dark rounded overflow-hidden border border-slate-800">
                <div
                  class="h-full bg-gradient-to-r from-emerald-500 to-cyan-500 transition-all duration-300"
                  :style="{ width: pct(st.win_rate) + '%' }"
                ></div>
              </div>
            </div>
          </div>

          <!-- Card usage chart -->
          <div
            v-if="cardEntries.length"
            class="bg-brand-bg border border-slate-800 rounded-xl p-5 shadow-xl"
          >
            <h2 class="font-mono text-brand-text-subtle font-bold mb-4">
              {{ $t("simulation.results.cardUsage") }}
            </h2>
            <div v-for="[card, count] in cardEntries" :key="card" class="mb-2">
              <div class="flex justify-between text-xs mb-0.5 font-mono">
                <span class="text-slate-300">{{ card }}</span>
                <span class="text-slate-400">{{ count }}</span>
              </div>
              <div class="w-full h-2.5 bg-brand-bg-dark rounded overflow-hidden">
                <div
                  class="h-full bg-gradient-to-r from-violet-500 to-fuchsia-500"
                  :style="{ width: cardBarPct(count) + '%' }"
                ></div>
              </div>
            </div>
          </div>

          <!-- Strategy breakdown table -->
          <div
            v-if="strategyList.length"
            class="bg-brand-bg border border-slate-800 rounded-xl overflow-hidden shadow-xl"
          >
            <div class="p-4 border-b border-slate-800">
              <h2 class="font-mono text-brand-text-subtle font-bold">
                {{ $t("simulation.results.strategyTable") }}
              </h2>
            </div>
            <div class="overflow-x-auto">
              <table class="w-full text-left text-xs">
                <thead>
                  <tr class="bg-brand-bg-dark/50 text-slate-400 uppercase tracking-wider border-b border-slate-800">
                    <th class="p-3">{{ $t("simulation.results.colStrategy") }}</th>
                    <th class="p-3 text-right">{{ $t("simulation.results.colGames") }}</th>
                    <th class="p-3 text-right">{{ $t("simulation.results.colWins") }}</th>
                    <th class="p-3 text-right">{{ $t("simulation.results.colWinRate") }}</th>
                    <th class="p-3 text-right">{{ $t("simulation.results.colSpy") }}</th>
                    <th class="p-3 text-right">{{ $t("simulation.results.colMoves") }}</th>
                    <th class="p-3 text-right">{{ $t("simulation.results.colInvalid") }}</th>
                    <th class="p-3 text-right">{{ $t("simulation.results.colFirstMoveWin") }}</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-800/60 font-mono">
                  <tr v-for="st in strategyList" :key="st.strategy" class="hover:bg-brand-surface/20">
                    <td class="p-3 font-bold text-cyan-300">{{ st.strategy }}</td>
                    <td class="p-3 text-right">{{ st.games_played }}</td>
                    <td class="p-3 text-right">{{ st.wins }}</td>
                    <td class="p-3 text-right text-emerald-400">{{ pct(st.win_rate) }}%</td>
                    <td class="p-3 text-right">{{ st.spy_bonuses }}</td>
                    <td class="p-3 text-right">{{ st.total_moves }}</td>
                    <td class="p-3 text-right text-amber-400">{{ st.invalid_moves }}</td>
                    <td class="p-3 text-right">{{ firstMoveWinPct(st) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onBeforeUnmount } from "vue";
import { useI18n } from "vue-i18n";
import { apiFetch } from "../utils/api";

const { t } = useI18n();

const form = reactive({
  count: 100,
  seed: 0,
  slots: [{ strategy: "smart" }, { strategy: "random" }],
  firstMoveBotId: "",
});

const isLaunching = ref(false);
const errorMessage = ref(null);
const progress = ref(null);
let pollTimer = null;

const summary = computed(() => progress.value?.summary || {});

const slotId = (idx) => `sim-bot-${idx}`;

const addSlot = () => {
  if (form.slots.length < 4) form.slots.push({ strategy: "smart" });
};
const removeSlot = (idx) => {
  form.slots.splice(idx, 1);
  // Скидаємо first-move override, якщо він указував на видалений слот.
  if (form.firstMoveBotId && !form.slots.some((_, i) => slotId(i) === form.firstMoveBotId)) {
    form.firstMoveBotId = "";
  }
};

const progressPct = computed(() => {
  const p = progress.value;
  if (!p || !p.total_games) return 0;
  return Math.round((p.played_games / p.total_games) * 100);
});

const statusLabel = computed(() => {
  const s = progress.value?.status;
  if (!s) return "";
  return t(`simulation.progress.${s}`, s);
});

const statusClass = computed(() => {
  switch (progress.value?.status) {
    case "running":
      return "text-cyan-400";
    case "completed":
      return "text-emerald-400";
    case "failed":
      return "text-rose-400";
    case "cancelled":
      return "text-amber-400";
    default:
      return "text-slate-400";
  }
});

const strategyList = computed(() => {
  const per = summary.value?.per_strategy;
  if (!per) return [];
  return Object.values(per).sort((a, b) => (b.win_rate || 0) - (a.win_rate || 0));
});

const cardEntries = computed(() => {
  const totals = summary.value?.card_play_totals;
  if (!totals) return [];
  return Object.entries(totals).sort((a, b) => b[1] - a[1]);
});

const maxCardCount = computed(() => {
  return cardEntries.value.reduce((m, [, c]) => Math.max(m, c), 0);
});

const pct = (rate) => Math.round((rate || 0) * 1000) / 10;
const fmt = (n) => (n == null ? "0" : Math.round(n * 10) / 10);
const cardBarPct = (count) =>
  maxCardCount.value ? Math.round((count / maxCardCount.value) * 100) : 0;
const firstMoveWinPct = (st) => {
  if (!st.first_move_games) return "-";
  return Math.round((st.first_move_wins / st.first_move_games) * 1000) / 10 + "%";
};

const buildConfig = () => ({
  count: form.count,
  seed: form.seed,
  first_move_bot_id: form.firstMoveBotId,
  persist: true,
  slots: form.slots.map((s, idx) => ({
    bot_id: slotId(idx),
    strategy: s.strategy,
    username: `Bot-${idx}`,
  })),
});

const launch = async () => {
  errorMessage.value = null;
  if (form.slots.length < 2) {
    errorMessage.value = t("simulation.errors.minSlots");
    return;
  }
  isLaunching.value = true;
  try {
    const response = await apiFetch("/api/admin/simulations", {
      method: "POST",
      body: JSON.stringify(buildConfig()),
    });
    if (response.status === 401 || response.status === 403) return;
    if (!response.ok) {
      const data = await response.json().catch(() => ({}));
      throw new Error(data.error || t("simulation.errors.launchFailed"));
    }
    const data = await response.json();
    startPolling(data.batch_id);
  } catch (err) {
    errorMessage.value = err.message;
  } finally {
    isLaunching.value = false;
  }
};

const fetchProgress = async (batchId) => {
  try {
    const response = await apiFetch(`/api/admin/simulations/${batchId}`, { method: "GET" });
    if (response.status === 401 || response.status === 403) return;
    if (!response.ok) throw new Error(t("simulation.errors.loadFailed"));
    progress.value = await response.json();
    if (progress.value.status !== "running") {
      stopPolling();
    }
  } catch (err) {
    errorMessage.value = err.message;
    stopPolling();
  }
};

const startPolling = (batchId) => {
  stopPolling();
  fetchProgress(batchId);
  pollTimer = setInterval(() => fetchProgress(batchId), 700);
};

const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
};

const cancelBatch = async () => {
  const batchId = progress.value?.batch_id;
  if (!batchId) return;
  await apiFetch(`/api/admin/simulations/${batchId}/cancel`, { method: "POST" });
};

onBeforeUnmount(stopPolling);
</script>
