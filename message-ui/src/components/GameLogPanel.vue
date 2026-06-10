<template>
  <div
    class="flex flex-col h-full bg-brand-bg-dark/60 border-l border-slate-800/80 overflow-hidden backdrop-blur-sm shadow-inner"
  >
    <div
      class="px-3 py-2 border-b border-slate-800/80 bg-brand-surface/40 flex justify-between items-center flex-shrink-0 h-[6vh]"
    >
      <span
        class="text-[10px] uppercase font-bold tracking-wider text-amber-400 font-mono flex items-center gap-1.5"
      >
        {{ $t("log.history") }}
      </span>
      <span
        class="text-[10px] font-mono text-slate-500 bg-brand-bg-dark px-1.5 py-0.5 rounded"
      >
        {{ gameLog.length }}
      </span>
    </div>

    <div
      ref="logContainer"
      class="flex-1 overflow-y-auto p-2 custom-scrollbar space-y-1.5 text-left text-xs selection:bg-amber-500/30 bg-brand-bg-dark/20"
    >
      <div
        v-if="gameLog.length === 0"
        class="text-slate-500 italic text-center py-4 text-[11px]"
      >
        {{ $t("log.empty") }}
      </div>

      <!-- Додаємо унікальний рядковий префікс до ключа log- -->
      <div
        v-for="(log, index) in gameLog"
        :key="`log-${log.id || index}`"
        class="group flex items-start gap-1.5 hover:bg-white/5 p-1 rounded transition-colors duration-150 animate-fade-in"
      >
        <span
          class="text-[10px] font-mono text-slate-500 bg-slate-900/40 px-1 py-0.5 rounded flex-shrink-0"
        >
          {{ log.timestamp }}
        </span>

        <span class="text-slate-200 leading-normal break-words text-[11px] flex-1">
          <!-- v-if захищає від спроб рендеру у зникаючому з DOM батьківському елементі -->
          <i18n-t v-if="log && log.messageKey" :keypath="log.messageKey" scope="global">
            <template #player>
              <strong class="text-amber-300 font-semibold">
                {{ log.namedArgs?.player || "" }}
              </strong>
            </template>

            <template #target>
              <strong class="text-cyan-400 font-semibold">
                {{ log.namedArgs?.target || "" }}
              </strong>
            </template>

            <template #winner>
              <strong class="text-emerald-400 font-semibold">
                {{ log.namedArgs?.winner || "" }}
              </strong>
            </template>

            <template #loser>
              <strong class="text-rose-400 font-semibold">
                {{ log.namedArgs?.loser || "" }}
              </strong>
            </template>

            <template #card>
              <span
                class="text-yellow-400 font-medium underline decoration-yellow-500/40"
              >
                {{
                  log.namedArgs?.card
                    ? log.namedArgs.card.includes(".")
                      ? $t(log.namedArgs.card)
                      : log.namedArgs.card
                    : ""
                }}
              </span>
            </template>

            <template #guess>
              <span class="text-rose-400 font-medium italic">
                {{
                  log.namedArgs?.guess
                    ? log.namedArgs.guess.includes(".")
                      ? $t(log.namedArgs.guess)
                      : log.namedArgs.guess
                    : ""
                }}
              </span>
            </template>

            <template #points>
              <span class="text-purple-400 font-bold font-mono">
                {{ log.namedArgs?.points ? `+${log.namedArgs.points}` : "" }}
              </span>
            </template>

            <template #reason>
              <span class="text-slate-400 italic text-[10px]">
                {{
                  log.namedArgs?.reason
                    ? log.namedArgs.reason.includes(".")
                      ? $t(log.namedArgs.reason)
                      : log.namedArgs.reason
                    : ""
                }}
              </span>
            </template>
          </i18n-t>
        </span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, nextTick, onMounted } from "vue";
import { storeToRefs } from "pinia";
import { useGameStore } from "../stores/gameStore";

const gameStore = useGameStore();
const { gameLog } = storeToRefs(gameStore);
const logContainer = ref(null);

const scrollToBottom = async () => {
  await nextTick();
  if (logContainer.value) {
    logContainer.value.scrollTop = logContainer.value.scrollHeight;
  }
};

watch(
  () => gameLog.value.length,
  () => {
    scrollToBottom();
  },
  { deep: true }
);

onMounted(() => {
  scrollToBottom();
});
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: rgba(15, 23, 42, 0.4);
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(245, 158, 11, 0.2);
  border-radius: 2px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(245, 158, 11, 0.4);
}
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(2px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
.animate-fade-in {
  animation: fadeIn 0.2s ease-out forwards;
}
</style>
