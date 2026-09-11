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
            class="relative w-28 h-28 bg-brand-bg-dark border-2 border-amber-500 rounded-full p-1 overflow-hidden shadow-xl shadow-amber-950/20"
          >
            <img
              :src="avatarUrl"
              :alt="$t('profile.avatarAlt')"
              class="w-full h-full object-cover rounded-full"
            />
            <div
              v-if="isUploadingAvatar"
              class="absolute inset-0 rounded-full bg-black/60 flex items-center justify-center"
            >
              <span
                class="w-6 h-6 rounded-full border-2 border-amber-400 border-t-transparent animate-spin"
              ></span>
            </div>
          </div>

          <div class="flex flex-wrap items-center justify-center gap-2">
            <button
              @click="triggerFilePicker"
              type="button"
              :disabled="isUploadingAvatar"
              class="px-4 py-1.5 bg-brand-surface text-brand-text-muted border border-brand-border font-bold text-xs uppercase rounded-xl hover:bg-brand-surface-dim transition disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ $t("profile.uploadPhoto") }}
            </button>
            <button
              @click="randomizeAvatar"
              type="button"
              :disabled="isUploadingAvatar"
              class="px-4 py-1.5 bg-brand-surface text-brand-text-muted border border-brand-border font-bold text-xs uppercase rounded-xl hover:bg-brand-surface-dim transition disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ $t("profile.avatarChange") }}
            </button>
          </div>

          <input
            ref="fileInputRef"
            type="file"
            :accept="ALLOWED_AVATAR_TYPES.join(',')"
            class="hidden"
            @change="handleFileSelected"
          />

          <p class="text-[10px] text-slate-500 text-center">
            {{ $t("profile.uploadHint", { maxMb: MAX_AVATAR_SIZE_MB }) }}
          </p>

          <div
            v-if="avatarError"
            class="w-full text-xs text-rose-400 bg-rose-500/10 border border-rose-500/20 p-2.5 rounded-xl text-center"
          >
            {{ avatarError }}
          </div>
          <div
            v-if="avatarSuccess"
            class="w-full text-xs text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 p-2.5 rounded-xl text-center"
          >
            {{ avatarSuccess }}
          </div>
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
import { apiFetch } from "../utils/api";

const { t } = useI18n();

const props = defineProps({
  currentUsername: { type: String, default: "" },
  currentSeed: { type: String, default: "" },
  currentAvatarUrl: { type: String, default: "" },
  apiUrl: { type: String, required: true },
});

const emit = defineEmits(["close", "updated"]);

const authStore = useAuthStore();

const currentSeedState = ref(props.currentSeed || generateRandomSeed());
const uploadedAvatarUrl = ref(props.currentAvatarUrl || "");
const previewUrl = ref("");
const isSaving = ref(false);
const localError = ref(null);

const form = ref({
  username: props.currentUsername || localStorage.getItem("username") || "",
  avatar_seed: currentSeedState.value,
  password: "",
  password_old: "",
});

// Клієнтські межі валідації файлу аватара - визначені поруч з місцем
// використання (правило AGENTS.md щодо необмежених вхідних даних).
const MAX_AVATAR_SIZE_MB = 5;
const ALLOWED_AVATAR_TYPES = ["image/png", "image/jpeg", "image/webp", "image/svg+xml"];

const fileInputRef = ref(null);
const isUploadingAvatar = ref(false);
const avatarError = ref("");
const avatarSuccess = ref("");

const avatarUrl = computed(() => {
  if (previewUrl.value) return previewUrl.value;
  if (uploadedAvatarUrl.value) return uploadedAvatarUrl.value;
  return getAvatarUrl(currentSeedState.value);
});

const randomizeAvatar = () => {
  currentSeedState.value = generateRandomSeed();
  form.value.avatar_seed = currentSeedState.value;
  // Випадковий DiceBear-аватар скасовує будь-яке раніше завантажене фото.
  uploadedAvatarUrl.value = "";
  previewUrl.value = "";
  avatarError.value = "";
  avatarSuccess.value = "";
};

const triggerFilePicker = () => {
  avatarError.value = "";
  avatarSuccess.value = "";
  fileInputRef.value?.click();
};

const handleFileSelected = async (event) => {
  const file = event.target.files?.[0];
  event.target.value = "";
  if (!file) return;

  avatarError.value = "";
  avatarSuccess.value = "";

  if (!ALLOWED_AVATAR_TYPES.includes(file.type)) {
    avatarError.value = t("profile.errors.invalidType");
    return;
  }
  if (file.size > MAX_AVATAR_SIZE_MB * 1024 * 1024) {
    avatarError.value = t("profile.errors.tooLarge", { maxMb: MAX_AVATAR_SIZE_MB });
    return;
  }

  // Локальний preview одразу, до відповіді сервера.
  const localPreview = URL.createObjectURL(file);
  previewUrl.value = localPreview;

  isUploadingAvatar.value = true;
  try {
    const formData = new FormData();
    formData.append("avatar", file);

    const response = await apiFetch(`${props.apiUrl}/api/user/avatar`, {
      method: "POST",
      body: formData,
    });

    if (response.status === 401 || response.status === 403) return;

    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(data.error || t("profile.errors.uploadFailed"));
    }

    uploadedAvatarUrl.value = data.avatar_url || "";
    previewUrl.value = "";
    avatarSuccess.value = t("profile.uploadSuccess");

    localStorage.setItem("avatar_url", uploadedAvatarUrl.value);
    if (authStore?.user) {
      authStore.user.avatar_url = uploadedAvatarUrl.value;
    }

    emit("updated", {
      username: form.value.username,
      avatar_seed: form.value.avatar_seed,
      avatar_url: uploadedAvatarUrl.value,
    });
  } catch (error) {
    console.error("Avatar upload error:", error);
    avatarError.value = error.message || t("profile.errors.uploadFailed");
    previewUrl.value = "";
  } finally {
    isUploadingAvatar.value = false;
    URL.revokeObjectURL(localPreview);
  }
};

const saveProfile = async () => {
  localError.value = null;

  if (form.value.password && !form.value.password_old) {
    localError.value = t("profile.errors.currentRequired");
    return;
  }

  isSaving.value = true;
  try {
    const response = await apiFetch(`${props.apiUrl}/api/user`, {
      method: "POST",
      body: JSON.stringify({
        username: form.value.username,
        avatar_seed: form.value.avatar_seed,
        password: form.value.password || undefined,
        password_old: form.value.password_old || undefined,
      }),
    });

    // 401/403 глобально обробляється apiFetch (редірект на /login).
    if (response.status === 401 || response.status === 403) return;

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

    // Якщо статус "success" - фіксуємо оновлені дані у локальних сховищах
    localStorage.setItem("username", data.username || form.value.username);
    localStorage.setItem("avatar_seed", data.avatar_seed || form.value.avatar_seed);
    localStorage.setItem("avatar_url", uploadedAvatarUrl.value);

    if (authStore?.user) {
      authStore.user.username = data.username || form.value.username;
      authStore.user.avatar_seed = data.avatar_seed || form.value.avatar_seed;
      authStore.user.avatar_url = uploadedAvatarUrl.value;
    }

    // Передаємо батьківському компоненту підтверджені бекендом дані
    emit("updated", {
      username: data.username || form.value.username,
      avatar_seed: data.avatar_seed || form.value.avatar_seed,
      avatar_url: uploadedAvatarUrl.value,
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
