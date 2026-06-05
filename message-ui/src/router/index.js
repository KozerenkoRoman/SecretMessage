/* ===== FILE: router/index.js ===== */
import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '../stores/auth';
import RoomManager from '../views/RoomManager.vue';

const routes = [
  {
    path: '/',
    redirect: '/auth'
  },
  {
    path: '/auth',
    name: 'Auth',
    component: () => import('../views/AuthView.vue'),
    meta: { guestOnly: true }
  },
  {
    path: '/desktop',
    name: 'Desktop',
    component: () => import('../views/DesktopView.vue'),
    meta: { requiresAuth: true }
  },
  // Наш захищений маршрут для панелі адміністратора
  {
    path: '/admin',
    name: 'Admin',
    component: () => import('../views/AdminView.vue'),
    meta: {
      requiresAuth: true,     // Доступ тільки для авторизованих користувачів
      requiresAdmin: true     // Доступ тільки для користувачів з роллю 'admin'
    }
  },
  {
    path: '/room/:id',
    name: 'Room',
    component: RoomManager,
    meta: { requiresAuth: true }
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

/* ===== FILE: router/index.js (Оновлена логіка гварда) ===== */
router.beforeEach((to, from) => {
  const authStore = useAuthStore();

  console.log(`Гвард: Перехід на ${to.path}. Статус авторизації: ${authStore.isAuthenticated}`);

  // 1. Якщо сторінка вимагає авторизації, а користувач не увійшов — відправляємо на логін
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return { path: '/auth', query: { redirect: to.fullPath } };
  }

  // 2. Якщо користувач УЖЕ авторизований і намагається відкрити екран входу (/auth)
  if (to.meta.guestOnly && authStore.isAuthenticated) {
    // Якщо це адмін, його стартова сторінка /admin, якщо гравець — /desktop
    return authStore.isAdmin ? '/admin' : '/desktop';
  }

  // 3. ПЕРЕВІРКА ДОСТУПУ: Якщо сторінка вимагає роль адміна, а користувач звичайний гравець
  if (to.meta.requiresAdmin && !authStore.isAdmin) {
    console.warn('Спроба несанкціонованого доступу до адмінки користувачем:', authStore.user?.username);
    return '/desktop'; // Звичайного гравця викидає
  }

  // В усіх інших випадках (включаючи перехід адміна на /desktop чи в ігрові кімнати) — дозволяємо рух!
  return true;
});
export default router;