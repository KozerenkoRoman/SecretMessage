<template>
  <div class="log-panel-container">
    <div class="log-panel-header">
      <span
        class="text-[12px] uppercase font-bold tracking-wider text-amber-400 font-mono flex items-center gap-1.5"
      >
        {{ $t("log.history") }}
      </span>
      <span
        class="text-[12px] font-mono text-slate-500 bg-brand-bg-dark px-1.5 py-0.5 rounded"
      >
        {{ gameLog.length }}
      </span>
    </div>

    <div ref="logContainer" class="log-panel-scroll custom-scrollbar space-y-1.5">
      <div
        v-if="gameLog.length === 0"
        class="text-slate-500 italic text-center py-4 text-[11px]"
      >
        {{ $t("log.empty") }}
      </div>

      <div
        v-for="(log, index) in gameLog"
        :key="`log-${log.id || index}`"
        class="group log-item-row animate-fade-in"
      >
        <span
          class="text-[12px] font-mono text-slate-500 bg-slate-900/40 px-1 py-0.5 rounded flex-shrink-0"
        >
          {{ log.timestamp }}
        </span>

        <span class="text-slate-200 leading-normal break-words text-[12px] flex-1">
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

            <template #discarded_card>
              <span
                class="text-orange-400 font-medium underline decoration-orange-500/40"
              >
                {{
                  log.namedArgs?.discarded_card
                    ? log.namedArgs.discarded_card.includes(".")
                      ? $t(log.namedArgs.discarded_card)
                      : log.namedArgs.discarded_card
                    : ""
                }}
              </span>
            </template>

            <template #winner_card>
              <span class="text-emerald-400 font-medium italic">
                {{
                  log.namedArgs?.winner_card
                    ? log.namedArgs.winner_card.includes(".")
                      ? $t(log.namedArgs.winner_card)
                      : log.namedArgs.winner_card
                    : ""
                }}
              </span>
            </template>

            <template #loser_card>
              <span class="text-rose-400 font-medium italic">
                {{
                  log.namedArgs?.loser_card
                    ? log.namedArgs.loser_card.includes(".")
                      ? $t(log.namedArgs.loser_card)
                      : log.namedArgs.loser_card
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
              <span class="text-slate-400 italic text-[11px]">
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
