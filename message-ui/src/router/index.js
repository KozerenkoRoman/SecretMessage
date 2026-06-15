import { i18n } from '../i18n';
import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '../stores/auth';
import RoomManager from '../views/RoomManager.vue';
import { watch } from 'vue';

const routes = [
  {
    path: '/',
    redirect: '/auth'
  },
  {
    path: '/auth',
    name: 'Auth',
    component: () => import('../views/AuthView.vue'),
    meta: { guestOnly: true, titleKey: 'titles.auth' }
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('../views/RegisterView.vue'),
    meta: { guestOnly: true, titleKey: 'titles.register' }
  },
  {
    path: '/desktop',
    name: 'Desktop',
    component: () => import('../views/DesktopView.vue'),
    meta: { requiresAuth: true, titleKey: 'titles.desktop' }
  },
  {
    path: '/admin',
    name: 'Admin',
    component: () => import('../views/AdminView.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      titleKey: 'titles.admin'
    }
  },
  {
    path: '/room/:id',
    name: 'Room',
    component: RoomManager,
    meta: { requiresAuth: true, titleKey: 'titles.room' }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/desktop'
  }
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

/* ===== Логіка гварда ===== */
router.beforeEach((to, from) => {
  const authStore = useAuthStore();

  console.log(`Гвард: Перехід на ${to.path}. Статус авторизації: ${authStore.isAuthenticated}`);

  // 1. Якщо сторінка вимагає авторизації, а користувач НЕ увійшов
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return { path: '/auth', query: { redirect: to.fullPath } };
  }

  // 2. Якщо користувач УЖЕ авторизований і намагається відкрити екрани для гостей (/auth або /register)
  if (to.meta.guestOnly && authStore.isAuthenticated) {
    // Якщо це адмін, його стартова сторінка /admin, якщо гравець - /desktop
    return authStore.isAdmin ? '/admin' : '/desktop';
  }

  // 3. ПЕРЕВІРКА РОЛІ: Якщо сторінка вимагає адміна, а користувач - звичайний гравець
  if (to.meta.requiresAdmin && !authStore.isAdmin) {
    console.warn('Спроба несанкціонованого доступу до адмінки користувачем:', authStore.user?.username);
    return '/desktop';
  }

  // В усіх інших випадках - дозволяємо рух!
  return true;
});

router.afterEach((to) => {
  const i18nGlobal = i18n.global;

  if (i18nGlobal) {
    const updateTitle = () => {
      const titleKey = to.meta.titleKey || 'titles.default';
      const translated = i18nGlobal.t(titleKey);
      document.title = translated !== titleKey ? translated : 'Таємне Послання';

      console.log(`[Router] Встановлено заголовок вкладки: "${document.title}" (Ключ: ${titleKey})`);
    };

    updateTitle();

    if (!router.titleWatcher) {
      router.titleWatcher = watch(
        () => i18nGlobal.locale.value,
        () => {
          const currentTitleKey = router.currentRoute.value.meta.titleKey || 'titles.default';
          document.title = i18nGlobal.t(currentTitleKey);
          console.log(`[Router] Мова змінилася! Новий заголовок: "${document.title}"`);
        }
      );
    }
  } else {
    console.error("[Router Error] Не вдалося отримати доступ до i18n.global. Перевірте імпорт файлу i18n.");
  }
});

export default router;