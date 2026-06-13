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
    path: '/register',
    name: 'register',
    component: () => import('../views/RegisterView.vue'),
    meta: { guestOnly: true }
  },
  {
    path: '/desktop',
    name: 'Desktop',
    component: () => import('../views/DesktopView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/admin',
    name: 'Admin',
    component: () => import('../views/AdminView.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: true
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
    // Якщо це адмін, його стартова сторінка /admin, якщо гравець — /desktop
    return authStore.isAdmin ? '/admin' : '/desktop';
  }

  // 3. ПЕРЕВІРКА РОЛІ: Якщо сторінка вимагає адміна, а користувач — звичайний гравець
  if (to.meta.requiresAdmin && !authStore.isAdmin) {
    console.warn('Спроба несанкціонованого доступу до адмінки користувачем:', authStore.user?.username);
    return '/desktop';
  }

  // В усіх інших випадках — дозволяємо рух!
  return true;
});

export default router;