<template>
  <div class="min-h-screen bg-slate-950 text-white p-6">
    <div class="max-w-5xl mx-auto">
      <div
        class="flex justify-between items-center mb-8 bg-slate-900 p-4 rounded-xl border border-slate-800"
      >
        <div>
          <h1 class="text-2xl font-bold text-rose-500 font-mono tracking-wide">
            {{ $t("admin.title") }}
          </h1>
          <p class="text-xs text-slate-400">{{ $t("admin.subtitle") }}</p>
        </div>
        <router-link
          to="/desktop"
          class="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-amber-500 hover:text-amber-400 text-xs font-bold uppercase tracking-wider font-mono rounded-xl transition-all border border-slate-700 shadow-md active:scale-95"
        >
          {{ $t("admin.backToDesktop") }}
        </router-link>
      </div>

      <div
        v-if="errorMessage"
        class="p-4 mb-6 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-400 text-sm"
      >
        {{ errorMessage }}
      </div>
      <div
        v-if="successMessage"
        class="p-4 mb-6 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-sm"
      >
        {{ successMessage }}
      </div>

      <div
        class="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden shadow-xl"
      >
        <div class="p-4 border-b border-slate-800 flex justify-between items-center">
          <h2 class="font-mono text-slate-300 font-bold">
            {{ $t("admin.table.title") }}
          </h2>
          <button
            @click="fetchUsers"
            class="px-2.5 py-1 bg-slate-800 hover:bg-slate-700 text-amber-500 hover:text-amber-400 text-[10px] font-bold font-mono rounded-lg border border-slate-700 cursor-pointer transition-all"
          >
            {{ $t("admin.refresh") }}
          </button>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr
                class="bg-slate-950/50 text-slate-400 text-xs uppercase tracking-wider border-b border-slate-800"
              >
                <th class="p-4 font-semibold">{{ $t("admin.table.id") }}</th>
                <th class="p-4 font-semibold">{{ $t("admin.table.username") }}</th>
                <th class="p-4 font-semibold">{{ $t("admin.table.email") }}</th>
                <th class="p-4 font-semibold">{{ $t("admin.table.role") }}</th>
                <th class="p-4 font-semibold text-right">
                  {{ $t("admin.table.action") }}
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800/60 text-sm">
              <tr
                v-for="user in users"
                :key="user.id"
                class="hover:bg-slate-800/30 transition-colors"
              >
                <td
                  class="p-4 font-mono text-xs text-slate-500 max-w-[120px] truncate"
                  :title="user.id"
                >
                  {{ user.id }}
                </td>
                <td class="p-4 font-semibold text-slate-200">{{ user.username }}</td>
                <td class="p-4 text-slate-400">{{ user.email || "—" }}</td>
                <td class="p-4">
                  <span
                    :class="[
                      'px-2 py-0.5 text-xs rounded-md font-medium',
                      user.role === 'admin'
                        ? 'bg-rose-500/10 text-rose-400 border border-rose-500/20'
                        : 'bg-slate-800 text-slate-400',
                    ]"
                  >
                    {{ user.role }}
                  </span>
                </td>
                <td class="p-4 text-right">
                  <button
                    v-if="user.role !== 'admin'"
                    @click="openBlockConfirmation(user.id, user.username)"
                    class="px-3 py-1 bg-rose-950/40 hover:bg-rose-900/60 text-rose-400 font-bold rounded-xl border border-rose-900/50 shadow-md active:scale-95 transition-all cursor-pointer text-xs uppercase tracking-wider font-mono"
                  >
                    {{ $t("admin.block") }}
                  </button>
                  <span v-else class="text-xs text-slate-600 italic">{{
                    $t("admin.blocked")
                  }}</span>
                </td>
              </tr>
              <tr v-if="users.length === 0">
                <td colspan="5" class="p-8 text-center text-slate-500 text-sm italic">
                  {{ $t("admin.emptyUsers") }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <ConfirmModal
      :is-open="showConfirmModal"
      :title="$t('admin.blockConfirm.title')"
      :message="
        $t('admin.blockConfirm.message', {
          username: selectedUser?.username,
        })
      "
      :confirm-text="$t('admin.blockConfirm.confirm')"
      :cancel-text="$t('admin.blockConfirm.cancel')"
      @confirm="handleConfirmBlock"
      @cancel="closeConfirmModal"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import { useAuthStore } from "../stores/auth";
import ConfirmModal from "./ConfirmModal.vue";

const { t } = useI18n();
const authStore = useAuthStore();
const users = ref([]);
const errorMessage = ref(null);
const successMessage = ref(null);
const showConfirmModal = ref(false);
const selectedUser = ref(null);

const fetchUsers = async () => {
  try {
    errorMessage.value = null;
    const token = authStore.token || localStorage.getItem("token");
    const response = await fetch("/api/admin/users", {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
    });
    if (response.status === 403 || response.status === 401) {
      throw new Error(t("admin.errors.noAccess"));
    }
    if (!response.ok) throw new Error(t("admin.errors.loadFailed"));
    const data = await response.json();
    users.value = Array.isArray(data) ? data : data.users || [];
  } catch (err) {
    errorMessage.value = err.message;
    console.error("Помилка адмінки:", err);
  }
};

const openBlockConfirmation = (userId, username) => {
  selectedUser.value = { id: userId, username };
  showConfirmModal.value = true;
};

const closeConfirmModal = () => {
  showConfirmModal.value = false;
  selectedUser.value = null;
};

const handleConfirmBlock = async () => {
  if (!selectedUser.value) return;

  const { id: userId, username } = selectedUser.value;
  closeConfirmModal();

  try {
    errorMessage.value = null;
    successMessage.value = null;
    const token = authStore.token || localStorage.getItem("token");
    const url = `/api/admin/users/${userId}/block`;
    const response = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({}),
    });
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(
        errorData.message || t("admin.errors.blockFailed", { username })
      );
    }
    successMessage.value = t("admin.blockSuccess", { username });
    setTimeout(() => {
      successMessage.value = null;
    }, 4000);
    fetchUsers();
  } catch (err) {
    errorMessage.value = err.message;
  }
};

onMounted(() => {
  fetchUsers();
});
</script>
