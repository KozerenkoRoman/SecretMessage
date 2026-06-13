/* ===== FILE: stores/auth.js ===== */
import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { jwtDecode } from 'jwt-decode';

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || null);
  const user = ref(null);

  const initUserFromToken = (jwtToken) => {
    if (!jwtToken) return;
    try {
      const decoded = jwtDecode(jwtToken);
      user.value = {
        id: decoded.user_id,
        username: decoded.username,
        role: decoded.user_role,
        avatar_seed: decoded.avatar_seed
      };
    } catch (e) {
      console.error("Помилка декодування JWT:", e);
      clearSession();
    }
  };

  if (token.value) {
    initUserFromToken(token.value);
  }

  const isAuthenticated = computed(() => !!token.value);
  const isAdmin = computed(() => user.value?.role === 'admin');

  const setSession = (newToken, userData = null) => {
    token.value = newToken;
    localStorage.setItem('token', newToken);

    if (userData) {
      user.value = userData;
    } else {
      initUserFromToken(newToken);
    }
  };

  const clearSession = () => {
    token.value = null;
    user.value = null;
    localStorage.removeItem('token');
    localStorage.removeItem('user_id');
    localStorage.removeItem('avatar_seed');
    console.log("[Pinia] Сесію користувача успішно очищено");
  };

  return {
    token,
    user,
    isAuthenticated,
    isAdmin,
    setSession,
    clearSession // експортуємо саме clearSession
  };
});