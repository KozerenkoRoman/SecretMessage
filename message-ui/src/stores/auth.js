import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { jwtDecode } from 'jwt-decode';

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || null);
  const user = ref(null);
  // Прапорець "сесію завершено через протермінування" — щоб екран логіну
  // міг показати відповідне повідомлення. Скидається при новому setSession.
  const sessionExpired = ref(false);

  const initUserFromToken = (jwtToken) => {
    if (!jwtToken) return;
    try {
      const decoded = jwtDecode(jwtToken);
      user.value = {
        id: decoded.user_id,
        username: decoded.username,
        role: decoded.user_role,
        avatar_seed: decoded.avatar_seed,
        // avatar_url НЕ зберігається в JWT (щоб не роздувати токен), тож
        // читаємо його з localStorage - синхронізується там при кожному
        // логіні/реєстрації/оновленні профілю (див. AuthView/RegisterView/
        // UserProfileModal).
        avatar_url: localStorage.getItem('avatar_url') || '',
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

  // setSession — ЄДИНА точка входу для встановлення активної сесії
  // (логін/реєстрація/оновлення профілю). userData, якщо переданий,
  // ПОВНІСТЮ перезаписує user.value — жодних злиттів зі старими даними,
  // щоб дані попереднього користувача на цьому ж пристрої не "протікали"
  // в сесію нового (див. баг з localStorage["username"] нижче).
  const setSession = (newToken, userData = null) => {
    token.value = newToken;
    sessionExpired.value = false;
    localStorage.setItem('token', newToken);

    if (userData) {
      user.value = { ...userData };
      // Синхронізуємо застарілі окремі ключі localStorage, які досі читають
      // деякі view (DesktopView, UserProfileModal), щоб вони НЕ показували
      // дані попереднього користувача, залишені в цьому ж браузері.
      localStorage.setItem('username', userData.username || '');
      localStorage.setItem('user_id', userData.id || '');
      localStorage.setItem('avatar_seed', userData.avatar_seed || '');
      localStorage.setItem('avatar_url', userData.avatar_url || '');
    } else {
      initUserFromToken(newToken);
    }
  };

  const clearSession = () => {
    token.value = null;
    user.value = null;
    localStorage.removeItem('token');
    localStorage.removeItem('username');
    localStorage.removeItem('user_id');
    localStorage.removeItem('avatar_seed');
    localStorage.removeItem('avatar_url');
    // Прибираємо будь-який кешований стан сесії і з sessionStorage.
    try {
      sessionStorage.removeItem('token');
      sessionStorage.removeItem('username');
      sessionStorage.removeItem('user_id');
      sessionStorage.removeItem('avatar_seed');
      sessionStorage.removeItem('avatar_url');
    } catch {
      /* sessionStorage може бути недоступний — не критично */
    }
    console.log("[Pinia] Сесію користувача успішно очищено");
  };

  /**
   * handleSessionExpired — єдина точка обробки протермінованої сесії
   * (викликається глобальним fetch-перехоплювачем при 401/403).
   * Повністю чистить стан автентифікації та піднімає прапорець sessionExpired,
   * щоб екран логіну показав відповідне повідомлення.
   */
  const handleSessionExpired = () => {
    clearSession();
    sessionExpired.value = true;
    console.warn('[Auth] Сесію протерміновано — стан автентифікації скинуто.');
  };

  const acknowledgeSessionExpired = () => {
    sessionExpired.value = false;
  };

  return {
    token,
    user,
    isAuthenticated,
    isAdmin,
    sessionExpired,
    setSession,
    clearSession, // експортуємо саме clearSession
    handleSessionExpired,
    acknowledgeSessionExpired
  };
});