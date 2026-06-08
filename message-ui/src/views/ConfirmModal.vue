<template>
  <Transition name="scale">
    <div
      v-if="isOpen"
      class="fixed inset-0 bg-black/80 backdrop-blur-md flex items-center justify-center p-4 z-[100]"
    >
      <div
        class="bg-slate-900 border-2 border-amber-500/40 rounded-2xl p-6 w-full max-w-md shadow-2xl text-center flex flex-col transform"
      >
        <h3
          class="text-amber-400 font-bold text-lg mb-3 flex items-center justify-center gap-2 font-mono uppercase tracking-wide"
        >
          {{ $t("modal.confirm.title") }}
        </h3>

        <p class="text-sm text-slate-300 font-sans leading-relaxed px-2 mb-6">
          {{ $t("modal.confirm.message") }}
        </p>

        <div class="flex gap-4 items-center justify-center">
          <button
            @click="handleCancel"
            type="button"
            class="flex-1 py-2.5 bg-slate-800 hover:bg-slate-700 text-slate-200 font-bold rounded-xl text-sm transition-all border border-slate-700/60 active:scale-95 cursor-pointer font-mono uppercase tracking-wider"
          >
            {{ $t("modal.confirm.cancel") }}
          </button>

          <button
            @click="handleConfirm"
            type="button"
            class="flex-1 py-2.5 bg-gradient-to-r from-rose-500 to-red-600 hover:from-rose-600 hover:to-red-700 text-white font-bold rounded-xl text-sm transition-all shadow-lg shadow-red-500/10 active:scale-95 cursor-pointer font-mono uppercase tracking-wider"
          >
            {{ $t("modal.confirm.ok") }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
defineProps({
  isOpen: {
    type: Boolean,
    required: true,
  },
  title: {
    type: String,
    default: "Підтвердження",
  },
  message: {
    type: String,
    default: "Ви впевнені, що хочете виконати цю дію?",
  },
  confirmText: {
    type: String,
    default: "Підтвердити",
  },
  cancelText: {
    type: String,
    default: "Скасувати",
  },
});

const emit = defineEmits(["confirm", "cancel"]);

const handleConfirm = () => emit("confirm");
const handleCancel = () => emit("cancel");
</script>

<style scoped>
.scale-enter-from,
.scale-leave-to {
  opacity: 0;
}
.scale-enter-from .bg-slate-900,
.scale-leave-to .bg-slate-900 {
  transform: scale(0.9) translateY(10px);
}
.scale-enter-active,
.scale-leave-active {
  transition: opacity 0.25s ease;
}
.scale-enter-active .bg-slate-900,
.scale-leave-active .bg-slate-900 {
  transition: transform 0.25s cubic-bezier(0.34, 1.56, 0.64, 1), opacity 0.25s ease;
}
</style>
