<template>
  <div
    class="inline-flex items-center gap-0.5 bg-brand-bg/80 backdrop-blur-sm border border-brand-border/60 rounded-xl p-0.5 shadow-md select-none"
    role="group"
    :aria-label="$t('language.switcherAriaLabel')"
  >
    <button
      v-for="lang in SUPPORTED_LOCALES"
      :key="lang.code"
      type="button"
      @click="setLanguage(lang.code)"
      :aria-pressed="locale === lang.code"
      :title="lang.title"
      :class="[
        'px-2.5 py-1 text-[11px] font-mono font-bold uppercase tracking-wider rounded-lg transition-all duration-200 cursor-pointer active:scale-95',
        locale === lang.code
          ? 'bg-brand-accent/15 text-amber-400 border border-amber-500/40 shadow-inner'
          : 'text-slate-400 hover:text-amber-300 border border-transparent',
      ]"
    >
      {{ lang.code.toUpperCase() }}
    </button>
  </div>
</template>

<script setup>
import { useI18n } from "vue-i18n";

const SUPPORTED_LOCALES = [
  { code: "uk", title: "Українська" },
  { code: "en", title: "English" },
];

const STORAGE_KEY = "lang";

const { locale } = useI18n();

const setLanguage = (newLang) => {
  if (!SUPPORTED_LOCALES.some((l) => l.code === newLang)) return;
  if (locale.value === newLang) return;

  locale.value = newLang;
  try {
    localStorage.setItem(STORAGE_KEY, newLang);
  } catch (err) {
    console.warn("[LanguageSwitcher] Не вдалося зберегти мову у localStorage:", err);
  }
  if (typeof document !== "undefined") {
    document.documentElement.lang = newLang;
  }
};
</script>
