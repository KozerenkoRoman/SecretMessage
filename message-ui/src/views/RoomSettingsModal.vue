<template>
  <Transition name="scale">
    <div
      v-if="isOpen"
      class="fixed inset-0 bg-black/80 backdrop-blur-md flex items-center justify-center p-2 sm:p-4 z-[100] h-app"
    >
      <div
        class="bg-brand-bg border-2 border-amber-500/40 rounded-2xl p-4 sm:p-5 lg:p-6 w-full max-w-md shadow-2xl text-left flex flex-col transform overflow-y-auto max-h-[90dvh] custom-scrollbar"
      >
        <h3
          class="text-amber-400 font-bold text-lg mb-4 flex items-center justify-center gap-2 font-mono uppercase tracking-wide"
        >
          {{ $t("roomSettings.title") }}
        </h3>

        <div
          class="flex items-center justify-between gap-3 bg-brand-bg-dark/60 border border-slate-800 rounded-xl p-3"
        >
          <div class="flex-1">
            <p class="text-sm font-medium text-brand-text-subtle">
              {{ $t("roomSettings.winnerStartsNextRound") }}
            </p>
            <p class="text-xs text-slate-500 mt-1">
              {{ $t("roomSettings.winnerStartsNextRoundHint") }}
            </p>
          </div>
          <button
            type="button"
            role="switch"
            :aria-checked="localWinnerStartsNextRound"
            :disabled="!isHost"
            @click="handleToggle"
            class="relative inline-flex h-6 w-11 flex-shrink-0 items-center rounded-full transition-colors border"
            :class="[
              localWinnerStartsNextRound
                ? 'bg-amber-500 border-amber-500'
                : 'bg-brand-surface border-brand-border',
              isHost ? 'cursor-pointer' : 'cursor-not-allowed opacity-50',
            ]"
          >
            <span
              class="inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform"
              :class="localWinnerStartsNextRound ? 'translate-x-6' : 'translate-x-1'"
            ></span>
          </button>
        </div>

        <p v-if="!isHost" class="text-xs text-slate-500 mt-3 text-center">
          {{ $t("roomSettings.hostOnlyHint") }}
        </p>

        <div class="flex gap-4 items-center justify-center mt-6">
          <button
            @click="$emit('close')"
            type="button"
            class="flex-1 py-2.5 bg-brand-surface hover:bg-brand-surface-dim text-brand-text-muted font-bold rounded-xl text-sm transition-all border border-brand-border/60 active:scale-95 cursor-pointer font-mono uppercase tracking-wider"
          >
            {{ $t("roomSettings.close") }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { ref, watch } from "vue";

const props = defineProps({
  isOpen: { type: Boolean, required: true },
  settings: { type: Object, default: () => ({ winner_starts_next_round: false }) },
  isHost: { type: Boolean, default: false },
});

const emit = defineEmits(["close", "update-settings"]);

const localWinnerStartsNextRound = ref(!!props.settings?.winner_starts_next_round);

watch(
  () => props.settings?.winner_starts_next_round,
  (val) => {
    localWinnerStartsNextRound.value = !!val;
  }
);

const handleToggle = () => {
  if (!props.isHost) return;
  const next = !localWinnerStartsNextRound.value;
  localWinnerStartsNextRound.value = next;
  emit("update-settings", { winner_starts_next_round: next });
};
</script>

<style scoped>
.scale-enter-from,
.scale-leave-to {
  opacity: 0;
}
.scale-enter-from .bg-brand-bg,
.scale-leave-to .bg-brand-bg {
  transform: scale(0.9) translateY(10px);
}
.scale-enter-active,
.scale-leave-active {
  transition: opacity 0.25s ease;
}
.scale-enter-active .bg-brand-bg,
.scale-leave-active .bg-brand-bg {
  transition: transform 0.25s cubic-bezier(0.34, 1.56, 0.64, 1), opacity 0.25s ease;
}
</style>
