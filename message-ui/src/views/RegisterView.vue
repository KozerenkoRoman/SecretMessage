<template>
  <div class="min-h-screen flex items-center justify-center p-4 bg-brand-bg-dark">
    <div
      class="relative w-full max-w-md bg-brand-surface p-6 rounded-2xl shadow-xl border border-brand-border"
    >
      <div class="absolute top-4 right-4 z-10">
        <LanguageSwitcher />
      </div>

      <h2
        class="text-2xl font-bold text-center mb-6 text-amber-500 font-mono tracking-wide"
      >
        {{ $t("register.title") }}
      </h2>

      <form @submit.prevent="handleRegister" class="space-y-4">
        <div
          v-if="registerError"
          class="p-3 rounded-lg bg-rose-500/20 border border-rose-500/30 text-rose-300 text-xs text-center font-medium"
        >
          {{ registerError }}
        </div>

        <div>
          <label class="block text-sm font-medium mb-1 text-slate-400">{{
            $t("register.username")
          }}</label>
          <input
            v-model="username"
            type="text"
            required
            class="w-full p-2.5 rounded-lg bg-brand-bg border border-brand-border focus:outline-none focus:border-amber-500 text-white font-medium"
          />
        </div>

        <div>
          <label class="block text-sm font-medium mb-1 text-slate-400">{{
            $t("register.email")
          }}</label>
          <input
            v-model="email"
            type="email"
            required
            class="w-full p-2.5 rounded-lg bg-brand-bg border border-brand-border focus:outline-none focus:border-amber-500 text-white font-medium"
            :placeholder="$t('register.emailPlaceholder')"
          />
        </div>

        <div>
          <label class="block text-sm font-medium mb-1 text-slate-400">{{
            $t("register.password")
          }}</label>
          <input
            v-model="password"
            type="password"
            required
            class="w-full p-2.5 rounded-lg bg-brand-bg border border-brand-border focus:outline-none focus:border-amber-500 text-white font-medium"
          />
        </div>

        <button
          type="submit"
          class="w-full py-3.5 bg-gradient-to-r from-amber-500 to-orange-600 hover:from-amber-600 hover:to-orange-700 text-slate-950 font-black rounded-xl shadow-lg shadow-amber-500/10 mt-4 cursor-pointer transition-all active:scale-95 font-mono uppercase text-sm tracking-wider"
        >
          {{ $t("register.submit") }}
        </button>

        <p class="text-center text-xs text-slate-500 mt-4">
          {{ $t("register.loginPrompt") }}
          <router-link to="/auth" class="text-amber-500 hover:underline ml-1">
            {{ $t("register.loginLink") }}
          </router-link>
        </p>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useAuthStore } from "../stores/auth";
import { useRouter } from "vue-router";
import LanguageSwitcher from "../components/LanguageSwitcher.vue";

const { t } = useI18n();
const authStore = useAuthStore();
const router = useRouter();

const username = ref("");
const email = ref("");
const password = ref("");
const registerError = ref(null);

const handleRegister = async () => {
  try {
    registerError.value = null;

    const bodyPayload = {
      username: username.value,
      email: email.value,
      password: password.value,
      avatar_seed: `user_${Math.random().toString(36).substring(2, 11)}`,
    };

    const response = await fetch("/api/register", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(bodyPayload),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.message || t("register.errors.registrationFailed"));
    }

    const data = await response.json();

    if (data.token) {
      authStore.setSession(data.token, {
        id: data.user_id,
        username: data.username,
        role: data.user_role,
        avatar_seed: data.avatar_seed,
      });

      localStorage.setItem("token", data.token);
      localStorage.setItem("user_id", data.user_id || "");
      localStorage.setItem("avatar_seed", data.avatar_seed || "");

      router.push("/desktop");
    } else {
      throw new Error(t("register.errors.noToken"));
    }
  } catch (err) {
    registerError.value = err.message;
    console.error("Реєстрація провалена:", err);
  }
};
</script>
