<template>
  <div class="min-h-screen flex items-center justify-center p-4 bg-slate-950">
    <div
      class="w-full max-w-md bg-slate-800 p-6 rounded-2xl shadow-xl border border-slate-700"
    >
      <h2
        class="text-2xl font-bold text-center mb-6 text-amber-500 font-mono tracking-wide"
      >
        {{ isRegister ? "Реєстрація" : "Вхід" }}
      </h2>

      <form @submit.prevent="handleSubmit" class="space-y-4">
        <div
          v-if="loginError"
          class="p-3 rounded-lg bg-rose-500/20 border border-rose-500/30 text-rose-300 text-xs text-center font-medium"
        >
          {{ loginError }}
        </div>

        <div>
          <label class="block text-sm font-medium mb-1 text-slate-400"
            >Логін / Нікнейм</label
          >
          <input
            v-model="username"
            type="text"
            required
            class="w-full p-2.5 rounded-lg bg-slate-900 border border-slate-700 focus:outline-none focus:border-amber-500 text-white font-medium"
          />
        </div>

        <div v-if="isRegister">
          <label class="block text-sm font-medium mb-1 text-slate-400"
            >Email-адреса</label
          >
          <input
            v-model="email"
            type="email"
            required
            class="w-full p-2.5 rounded-lg bg-slate-900 border border-slate-700 focus:outline-none focus:border-amber-500 text-white font-medium"
            placeholder="example@domain.com"
          />
        </div>

        <div>
          <label class="block text-sm font-medium mb-1 text-slate-400">Пароль</label>
          <input
            v-model="password"
            type="password"
            required
            class="w-full p-2.5 rounded-lg bg-slate-900 border border-slate-700 focus:outline-none focus:border-amber-500 text-white font-medium"
          />
        </div>

        <button
          type="submit"
          class="w-full py-3 bg-amber-500 hover:bg-amber-600 text-slate-950 font-bold rounded-lg transition-colors shadow-lg shadow-amber-500/10 mt-4 cursor-pointer"
        >
          {{ isRegister ? "Зареєструватися" : "Увійти до гри" }}
        </button>

        <p class="text-center text-xs text-slate-500 mt-4">
          {{ isRegister ? "Вже маєте акаунт?" : "Ще немає акаунта?" }}
          <span
            @click="toggleMode"
            class="text-amber-500 cursor-pointer hover:underline ml-1"
          >
            {{ isRegister ? "Увійти" : "Зареєструватися" }}
          </span>
        </p>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useAuthStore } from "../stores/auth";
import { useRouter } from "vue-router";

const authStore = useAuthStore();
const router = useRouter();

const isRegister = ref(false);
const username = ref("");
const email = ref(""); // Нове реактивне поле для email
const password = ref("");
const loginError = ref(null);

const toggleMode = () => {
  isRegister.value = !isRegister.value;
  loginError.value = null;
  email.value = "";
};

const handleSubmit = async () => {
  try {
    loginError.value = null;
    const endpoint = isRegister.value ? "/api/register" : "/api/auth";

    const bodyPayload = {
      username: username.value,
      password: password.value,
    };

    if (isRegister.value) {
      bodyPayload.email = email.value;
    }

    const response = await fetch(endpoint, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(bodyPayload),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(
        errorData.message ||
          (isRegister.value ? "Помилка реєстрації" : "Неправильний логін або пароль")
      );
    }

    const data = await response.json();
    console.log("Відповідь від сервера Go:", data);

    if (data.token) {
      // 1. Зберігаємо сесію та мапимо роль
      authStore.setSession(data.token, {
        username: data.username,
        role: data.user_role, // перетворюємо "user_role" від Go на "role" для Vue
      });

      localStorage.setItem("user_id", data.username);
      console.log("Успішний вхід. Роль користувача:", authStore.user?.role);

      // 2. АВТОМАТИЧНИЙ РЕДІРЕКТ НА ОСНОВІ РОЛІ:
      if (authStore.isAdmin) {
        console.log("Користувач є адміном. Перенаправлення на /admin");
        router.push("/admin");
      } else {
        console.log("Звичайний користувач. Перенаправлення на /desktop");
        const redirectPath = router.currentRoute.value.query.redirect || "/desktop";
        router.push(redirectPath);
      }
    } else {
      throw new Error("Сервер не повернув JWT-токен");
    }
  } catch (err) {
    loginError.value = err.message;
    console.error("Помилка автентифікації:", err);
  }
};
</script>
