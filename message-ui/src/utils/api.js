// =============================================================================
// utils/api.js
//
// ЦЕНТРАЛІЗОВАНИЙ FETCH-WRAPPER із глобальним перехопленням протермінованих сесій.
//
// Навіщо:
//   У проекті немає axios (див. AGENTS.md) — усі REST-виклики йдуть через
//   нативний fetch. Раніше кожен view сам перевіряв response.status === 401.
//   Тут ми зводимо це в ОДНЕ місце: apiFetch автоматично
//     • підставляє відносний /api-URL (проксі Vite/nginx маршрутизує на Go),
//     • додає Authorization: Bearer <jwt> та Content-Type,
//     • ловить 401 Unauthorized / 403 Forbidden і запускає єдиний обробник
//       протермінованої сесії (очищення стану + редірект на /login).
//
// Розв'язання циклічних імпортів:
//   Цей модуль — pure util і НЕ імпортує ані Pinia-стор, ані vue-router
//   (інакше отримаємо цикл util → store → router → ... і краш на mount).
//   Замість цього router/index.js під час старту РЕЄСТРУЄ callback через
//   registerSessionExpiredHandler(). До реєстрації fallback просто робить
//   hard-redirect на /login, тож безпека не залежить від порядку модулів.
// =============================================================================

// Активний обробник протермінованої сесії. Встановлюється з router/index.js.
let sessionExpiredHandler = null;

// Гарантія, що при шквалі паралельних 401 (кілька запитів разом) ми
// виконаємо очищення+редірект РІВНО один раз, а не N разів.
let isHandlingExpiry = false;

/**
 * registerSessionExpiredHandler — прив'язує глобальний обробник, який
 * викликається при 401/403. Приймає { redirect } — поточний шлях, куди
 * повернути користувача після повторного входу.
 *
 * @param {(opts: { redirect?: string }) => void} handler
 */
export function registerSessionExpiredHandler(handler) {
  sessionExpiredHandler = typeof handler === 'function' ? handler : null;
}

/**
 * triggerSessionExpired — внутрішньо викликається при 401/403.
 * Debounce через isHandlingExpiry, щоб уникнути гонок і подвійних редіректів.
 */
function triggerSessionExpired() {
  if (isHandlingExpiry) return;
  isHandlingExpiry = true;

  const redirect =
    typeof window !== 'undefined'
      ? window.location.pathname + window.location.search
      : undefined;

  try {
    if (sessionExpiredHandler) {
      sessionExpiredHandler({ redirect });
    } else if (typeof window !== 'undefined') {
      // Fallback: обробник ще не зареєстровано (дуже ранній виклик) —
      // виконуємо жорсткий редірект, щоб не лишити користувача на
      // захищеній сторінці з протермінованим токеном.
      localStorage.removeItem('token');
      const q = redirect ? `?redirect=${encodeURIComponent(redirect)}` : '';
      window.location.assign(`/login${q}`);
    }
  } finally {
    // Даємо навігації відпрацювати, потім знімаємо прапорець, щоб майбутні
    // (нові) сесії теж могли обробити протермінування.
    setTimeout(() => {
      isHandlingExpiry = false;
    }, 1000);
  }
}

function readToken() {
  try {
    return localStorage.getItem('token') || '';
  } catch {
    return '';
  }
}

/**
 * apiFetch — обгортка над fetch для всіх викликів до бекенду.
 *
 * @param {string} path      Відносний шлях (напр., "/api/rooms") або повний URL.
 * @param {RequestInit & { auth?: boolean, skipAuthRedirect?: boolean }} [options]
 *        - auth (default true): додати Authorization-заголовок з JWT.
 *        - skipAuthRedirect (default false): НЕ запускати глобальний обробник
 *          протермінування при 401/403 (для ендпоінтів логіну/реєстрації,
 *          де 401 — це "невірний пароль", а не "сесія протермінована").
 * @returns {Promise<Response>} Оригінальний Response (виклик сам читає .json()).
 */
export async function apiFetch(path, options = {}) {
  const { auth = true, skipAuthRedirect = false, headers = {}, ...rest } = options;

  // Для FormData (напр., multipart-завантаження файлів) НЕ виставляємо
  // Content-Type самі - браузер має сам згенерувати заголовок з коректним
  // multipart boundary. Якщо виставити 'application/json' (чи будь-яке
  // фіксоване значення) тут, boundary загубиться і бекенд не розпарсить тіло.
  const isFormData = typeof FormData !== 'undefined' && rest.body instanceof FormData;

  const finalHeaders = {
    ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
    ...headers,
  };

  if (auth) {
    const token = readToken();
    if (token && !finalHeaders.Authorization) {
      finalHeaders.Authorization = `Bearer ${token}`;
    }
  }

  const response = await fetch(path, { ...rest, headers: finalHeaders });

  // Глобальне перехоплення протермінованої сесії.
  if (!skipAuthRedirect && (response.status === 401 || response.status === 403)) {
    triggerSessionExpired();
  }

  return response;
}
