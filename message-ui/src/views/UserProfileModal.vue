<template>
  <div
    class="fixed inset-0 bg-brand-bg-dark/80 backdrop-blur-sm flex items-center justify-center z-50 p-4 overflow-y-auto"
  >
    <div
      class="bg-brand-bg border border-slate-800 rounded-2xl w-full max-w-md p-6 shadow-2xl relative my-8"
    >
      <h3
        class="text-lg font-bold text-yellow-400 font-mono mb-2 text-center uppercase tracking-wider"
      >
        {{ $t("profile.title") }}
      </h3>
      <p class="text-xs text-slate-400 text-center mb-6">
        {{ $t("profile.subtitle") }}
      </p>

      <form @submit.prevent="saveProfile" class="space-y-5">
        <div
          class="flex flex-col items-center gap-4 bg-brand-bg-dark/40 p-4 rounded-xl border border-slate-800/60"
        >
          <div
            class="w-28 h-28 bg-brand-bg-dark border-2 border-amber-500 rounded-full p-1 overflow-hidden shadow-xl shadow-amber-950/20"
          >
            <img
              :src="avatarUrl"
              :alt="$t('profile.avatarAlt')"
              class="w-full h-full object-cover rounded-full"
            />
          </div>

          <button
            @click="randomizeAvatar"
            type="button"
            class="px-4 py-1.5 bg-brand-surface text-brand-text-muted border border-brand-border font-bold text-xs uppercase rounded-xl hover:bg-brand-surface-dim transition"
          >
            🎲 {{ $t("profile.avatarChange") }}
          </button>
        </div>

        <div>
          <label
            class="block text-xs font-bold font-mono text-slate-400 uppercase mb-1.5 tracking-wider"
          >
            {{ $t("profile.usernameLabel") }}
          </label>
          <input
            v-model="form.username"
            type="text"
            required
            :placeholder="$t('profile.usernamePlaceholder')"
            class="w-full px-4 py-2.5 bg-brand-bg-dark border border-slate-800 rounded-xl text-brand-text-primary text-sm focus:outline-none focus:border-amber-500 transition"
          />
        </div>

        <div class="relative flex py-2 items-center">
          <div class="flex-grow border-t border-slate-800"></div>
          <span
            class="flex-shrink mx-4 text-[10px] font-mono uppercase text-slate-500 tracking-widest"
            >{{ $t("profile.passwordSectionDivider") }}</span
          >
          <div class="flex-grow border-t border-slate-800"></div>
        </div>

        <div>
          <label
            class="block text-xs font-bold font-mono text-slate-400 uppercase mb-1.5 tracking-wider"
          >
            {{ $t("profile.currentPassword") }}
            <span v-if="form.password" class="text-rose-500">*</span>
          </label>
          <input
            v-model="form.password_old"
            type="password"
            :required="!!form.password"
            :placeholder="$t('profile.currentPasswordPlaceholder')"
            class="w-full px-4 py-2.5 bg-brand-bg-dark border border-slate-800 rounded-xl text-brand-text-primary text-sm focus:outline-none focus:border-amber-500 transition"
          />
        </div>

        <div>
          <label
            class="block text-xs font-bold font-mono text-slate-400 uppercase mb-1.5 tracking-wider"
          >
            {{ $t("profile.newPassword") }}
          </label>
          <input
            v-model="form.password"
            type="password"
            :placeholder="$t('profile.newPasswordPlaceholder')"
            class="w-full px-4 py-2.5 bg-brand-bg-dark border border-slate-800 rounded-xl text-brand-text-primary text-sm focus:outline-none focus:border-amber-500 transition"
          />
        </div>

        <div
          v-if="localError"
          class="text-xs text-rose-400 bg-rose-500/10 border border-rose-500/20 p-2.5 rounded-xl"
        >
          ⚠️ {{ localError }}
        </div>

        <div class="flex gap-3 pt-2">
          <button
            @click="$emit('close')"
            type="button"
            class="flex-1 px-4 py-2.5 bg-brand-surface text-brand-text-subtle font-bold text-xs uppercase rounded-xl hover:bg-brand-surface-dim transition"
          >
            {{ $t("profile.cancel") }}
          </button>
          <button
            type="submit"
            :disabled="isSaving"
            class="flex-1 btn-primary py-2.5 font-bold text-xs uppercase rounded-xl transition"
          >
            {{ isSaving ? $t("profile.saving") : $t("profile.save") }}
          </button>
        </div>
      </form>

      <button
        @click="$emit('close')"
        class="absolute top-4 right-4 text-slate-400 hover:text-white font-bold text-sm"
      >
        ✕
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from "vue";
import { useI18n } from "vue-i18n";
import { useAuthStore } from "../stores/auth";
import { getAvatarUrl, generateRandomSeed } from "../utils/avatar";

const { t } = useI18n();

const props = defineProps({
  currentUsername: { type: String, default: "" },
  currentSeed: { type: String, default: "" },
  apiUrl: { type: String, required: true },
});

const emit = defineEmits(["close", "updated"]);

const authStore = useAuthStore();

const currentSeedState = ref(props.currentSeed || generateRandomSeed());
const isSaving = ref(false);
const localError = ref(null);

const form = ref({
  username: props.currentUsername || localStorage.getItem("username") || "",
  avatar_seed: currentSeedState.value,
  password: "",
  password_old: "",
});

const avatarUrl = computed(() => getAvatarUrl(currentSeedState.value));

const randomizeAvatar = () => {
  currentSeedState.value = generateRandomSeed();
  form.value.avatar_seed = currentSeedState.value;
};

const saveProfile = async () => {
  localError.value = null;

  if (form.value.password && !form.value.password_old) {
    localError.value = t("profile.errors.currentRequired");
    return;
  }

  isSaving.value = true;
  try {
    const token = authStore?.token || localStorage.getItem("token");

    const response = await fetch(`${props.apiUrl}/api/user`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        username: form.value.username,
        avatar_seed: form.value.avatar_seed,
        password: form.value.password || undefined,
        password_old: form.value.password_old || undefined,
      }),
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error || t("profile.errors.saveFailed"));
    }

    if (data.status === "no_changes") {
      form.value.password = "";
      form.value.password_old = "";
      emit("close");
      return;
    }

    // Якщо статус "success" — фіксуємо оновлені дані у локальних сховищах
    localStorage.setItem("username", data.username || form.value.username);
    localStorage.setItem("avatar_seed", data.avatar_seed || form.value.avatar_seed);

    if (authStore?.user) {
      authStore.user.username = data.username || form.value.username;
      authStore.user.avatar_seed = data.avatar_seed || form.value.avatar_seed;
    }

    // Передаємо батьківському компоненту підтверджені бекендом дані
    emit("updated", {
      username: data.username || form.value.username,
      avatar_seed: data.avatar_seed || form.value.avatar_seed,
    });

    // Очищуємо чутливі дані форми
    form.value.password = "";
    form.value.password_old = "";

    emit("close");
  } catch (error) {
    console.error("Profile save error:", error);
    localError.value = error.message || t("profile.errors.networkError");
  } finally {
    isSaving.value = false;
  }
};
</script>
