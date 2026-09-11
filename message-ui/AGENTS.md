# AGENTS.md

Instructions for AI assistants working in this repository.

## Project at a glance

- **Name:** `secret-message` (`message-ui`) — Vue 3 SPA for the *Secret Message* (Love Letter–style) card game.
- **Stack:** Vue 3.5 (Composition API, `<script setup>` SFCs) + Vite 8 + Pinia 3 + Vue Router + vue-i18n 9
  (`legacy: false`) + Tailwind CSS 4 (`@tailwindcss/vite`) + `jwt-decode`. Native `WebSocket` (no socket library).
- **Language mix:** JavaScript by default; **TypeScript is opt-in** and currently used only for stable
  domain contracts in `src/types/` and i18n message dictionaries in `src/i18n/`. Do **not** convert existing `.js`
  files to `.ts` without an explicit task.
- **Package manager:** npm. Lockfile committed at the repo root and inside `message-ui/`.
- **Entry points:** `index.html` → `src/main.js` → `App.vue` (`<router-view />` only). Dev server is Vite
  (default port 5173 unless overridden). The Vite dev server proxies `/api` and `/ws` to `http://localhost:3000`
  (the Go backend in `../message-backend/`).
- **Backend contract:** REST under `/api/*` and a single WebSocket at `/ws?token=<JWT>`. Auth is a JWT stored in
  `localStorage` under key `token`. Server payloads are `snake_case`; convert at the boundary.
- **Build-time globals:** `__APP_VERSION__` is injected from `package.json` via `vite.config.js` `define`.

## Commands you should run

| Task                       | Command         |
|----------------------------|-----------------|
| Install deps               | `npm install`   |
| Dev server (HMR)           | `npm run dev`   |
| Production build           | `npm run build` |
| Preview production build   | `npm run preview` |

There is currently **no lint, no formatter, no type-checker, and no test runner configured**. Do **not** add
any of these tools without an explicit user request — adding them would change CI assumptions and conflict
with the existing untyped `.js` codebase. If you genuinely need to verify a change, run `npm run build` (it
fails on syntax errors and broken imports) and reload the dev server in the browser.

## Repository layout

```
message-ui/
├── index.html                # Vite entry; mounts #app
├── vite.config.js            # Vite + @vitejs/plugin-vue + @tailwindcss/vite, /api & /ws proxy
├── tailwind.config.js        # Semantic `brand-*` palette
├── package.json              # Scripts: dev / build / preview
└── src/
    ├── main.js               # createApp, Pinia, Router, i18n, favicon, version banner
    ├── App.vue               # <router-view /> only — do not add layout here
    ├── i18n.js               # createI18n({ legacy: false, ... })
    ├── style.css             # Global Tailwind import + @custom-variant + @theme tokens + utility classes
    ├── assets/               # Static images (cards, chip, deckBack, favicon)
    ├── components/           # Reusable cross-view components (GameLogPanel, LeaderboardPanel, LanguageSwitcher)
    ├── constants/            # Static data (cards.js — CARD_INFO_NUMBERS, etc.)
    ├── i18n/                 # Locale dictionaries (en.ts, uk.ts) + errorMessages.ts
    ├── router/index.js       # Vue Router with beforeEach guard + afterEach title sync
    ├── stores/               # Pinia stores (auth.js, gameStore.js)
    ├── types/                # TypeScript domain contracts (events.ts, errors.ts) — server↔client truth
    ├── utils/                # Pure helpers (avatar.js, merge.js — in-place reactive state merge)
    └── views/                # Route-level views + modals; subfolder `board/` for in-game UI
```

Routes are declared in `src/router/index.js` and lazy-loaded with dynamic `import()`. Route components live
in `src/views/`. Modals are also `.vue` files in `src/views/` (e.g. `ConfirmModal.vue`, `GameEndModal.vue`).

## File-naming conventions (mirror existing code; do not invent new rules)

- **`.vue` files: PascalCase.** `BoardView.vue`, `RoomManager.vue`, `LanguageSwitcher.vue`. Both views and
  components follow this rule.
- **`.js` files: camelCase.** `gameStore.js`, `auth.js`, `avatar.js`, `merge.js`. Lowercase single-word names
  also fine (`auth.js`, `cards.js`).
- **`.ts` files: camelCase.** `events.ts`, `errors.ts`, `errorMessages.ts`, `en.ts`, `uk.ts`.
- **Folders: lowercase, single word where possible.** `views/`, `stores/`, `components/`, `views/board/`.
- **Routes:** the `name` field on a route is PascalCase or lowercase — match the surrounding entries; do not
  introduce a new style.
- **Pinia store ids:** kebab/camelCase string matching the store filename without `.js`
  (`'auth'`, `'gameStore'`).

## Coding conventions

### Code design (KISS, DRY, data-driven)

- **KISS.** Write the simplest thing that works. Don't introduce a constant, helper, or abstraction
  used in only one place.
- **Minimal, DRY changes.** Make the smallest change that solves the task. No opportunistic refactors of
  untouched code.
- **Data-driven over copy-paste.** Two or more similar branches over a known set of values become a
  module-level config array, not duplicated `if`s — adding a new case should be one array entry. Example
  from `LanguageSwitcher.vue`:
  ```js
  const SUPPORTED_LOCALES = [
    { code: 'uk', title: 'Українська' },
    { code: 'en', title: 'English' },
  ];
  ```
- **Code locality.** A helper used by a single feature lives next to it. Promote into `src/utils/` only
  when 2+ unrelated places need it (e.g. `merge.js`, `avatar.js`).
- **Lift the condition to the call site.** Keep the component pure and gate it where it is used —
  `<MyModal v-if="canShow" />` — rather than baking the condition into the component.

### Vue / Composition API

- **`<script setup>` only** for new components. Do not use the Options API; do not mix Options API with
  `<script setup>` in a single file.
- **`legacy: false` i18n.** Use `useI18n()` in `<script setup>` to read `t` / `locale`; in templates use
  `$t('key')` and `<i18n-t keypath="..." scope="global">` for slot-aware messages (see
  `GameLogPanel.vue`).
- **Reactivity primitives:**
  - `ref()` for scalar / object state where the consumer expects a `.value` handle.
  - `computed()` for derived state — never duplicate computation in a `watch` callback that just writes
    back to another `ref`.
  - `watch` / `watchEffect` only for *side effects* (DOM, network, log). Do not use `watch` to derive
    a value from another ref — that's what `computed` is for.
  - **Prefer derived state over `watch`.** If a value can be computed from refs/store state during render,
    use `computed`.
- **`v-for` always with stable `:key`.** When the list is rebuilt from server state, prefer a stable UID
  (`log.id`, `slot.uid`) over array index. The hand-slot pattern in `gameStore.js` (`handSlots`,
  `_nextHandUid()`) is the canonical example — copy it when you face the same problem.
- **Lifecycle order:** import lifecycle hooks (`onMounted`, `onBeforeUnmount`, etc.) and always pair every
  subscription, interval, timeout, or `addEventListener` with a teardown in `onBeforeUnmount`. The
  existing `gameStore` owns the WebSocket; views must call `gameStore.disconnect()` /
  `gameStore.leaveCurrentRoom()` instead of opening their own sockets.
- **Props down, events up.** Children declare `defineProps` and `defineEmits`. Do not mutate props.
- **Slots over prop-drilling render bits.** Use named slots for layout flexibility (`<i18n-t>` template
  slots are the in-repo pattern).
- **Async setup with caution.** Top-level `await` in `<script setup>` makes the component a Suspense
  boundary; we don't use `<Suspense>` anywhere — keep async work inside actions called from `onMounted`
  or in store actions.

### Performance

- **Stable references matter for Vue reactivity.** When merging server state into local state, do **not**
  re-assign whole arrays/objects on every WebSocket packet — it triggers re-renders and DOM rebuilds and
  has previously caused card-flicker bugs. Use `mergeState` / `mergeArray` from `src/utils/merge.js` for
  in-place merges. Don't `JSON.parse(JSON.stringify(...))` server payloads.
- **Hold `:key` stability for list items whose identity didn't change.** If you mutate hand slots or
  discard piles, mirror the `handSlots` UID pattern in `gameStore.js`.
- **Heavy derived state goes in the store, not in `watch({ deep: true })`.** Compute board-derived fields
  (`discardSequence`, `lastPlayedCardsByPlayer`, `handSlots`) **once per packet** in the store, not in
  every component. See `_updateBoardDerived` in `gameStore.js` — replicate that approach.
- **Lazy-load route components** with dynamic `import()` (see `router/index.js`). Eagerly import only the
  components that are needed before the first paint of the initial route.
- **Don't recreate large objects per render.** Hoist constants (icon maps, color tables) to module scope
  or wrap them in `computed`.
- **Local timers must mutate local refs only.** A 1 Hz tick that mutates a field of `gameState` would
  re-trigger every consumer of `gameState` — `gameStore` keeps `secondsLeft` separate for exactly this
  reason. Mirror that pattern.

### State management (Pinia)

- **Setup-style stores only.** `defineStore('id', () => { ... })` with `ref` / `computed` and returned
  actions. Mirror `src/stores/auth.js` and `src/stores/gameStore.js`.
- **One store = one concern.** `auth.js` holds session/JWT/user. `gameStore.js` holds the live game and
  the WebSocket. Don't merge them; don't add a third "everything" store.
- **Derived state via `computed`.** Expose getters as `computed` refs (e.g. `isAuthenticated`,
  `isAdmin`, `isGameStarted`).
- **Persistence:** the only thing persisted to `localStorage` today is `token` and `lang`. Do **not**
  persist game state or unbounded structures. If a new persisted preference is needed, add it under a
  documented, stable key and read it once at store init.
- **Don't access stores from utils** (`src/utils/`). Utils stay pure. Stores can call utils; utils never
  call stores.
- **Components consume stores via the `useXStore()` hook**, then either destructure with `storeToRefs` or
  use `computed(() => store.field)` to keep reactivity. Don't destructure plain values out of a Pinia
  store with `const { field } = useStore()` — you'll lose reactivity.

### TypeScript (where it exists)

- TypeScript is used **only** in `src/types/` and `src/i18n/`. These files are the single source of truth
  for backend↔frontend wire contracts (`events.ts`, `errors.ts`).
- **Do not convert `.js` → `.ts`** opportunistically. Adding TS coverage is a separate, scoped task
  (would require setting up `tsconfig.json` and `vue-tsc`).
- When editing `src/types/events.ts`:
  - Field names are `snake_case` to match the Go wire format. Do not camelCase them.
  - Numeric enums (`CardType`) mirror Go `iota`. Order is significant — change only in lockstep with
    `message-backend/cmd/engine/types.go`.
  - Discriminated unions (`GameEvent`) are how UI code narrows event payloads. Add new events to
    `EventType`, `EventTypeMap`, and the union — all three.
- Prefer **string-literal union types** over new enums for new code. Keep existing enums (`CardType`,
  `EliminationReason`) as they are.
- Avoid `any`. Prefer `unknown` + narrowing; document why if `any` is unavoidable.

### API & WebSocket layer

- **Single WebSocket** is owned by `useGameStore`. Connect via `gameStore.connectToHub(roomID)`,
  send via `gameStore.sendWSMessage(type, ...)`, and disconnect via `gameStore.disconnect()`. Do not
  open additional `WebSocket` instances elsewhere.
- **REST calls** today use `fetch` directly (see `addBotToRoom` in `gameStore.js`). All requests must
  attach the JWT as `Authorization: Bearer ${token}` and use a relative URL under `/api/...` so Vite's
  dev proxy (and nginx in prod) can route to the Go backend. Do not hard-code `http://localhost:3000`
  in components.
- **Server payload shape:** `snake_case`. Convert/normalize at the boundary (inside the store action that
  receives the message) — UI code reads camelCase or whatever the store re-exposes.
- **Errors** flow through `gameStore.error` (an object `{ code, message, details, requestId? }`).
  `GameErrorModal.vue` renders this; do not invent a parallel error channel.
- **Reconnect** is automatic in `gameStore` with a 3 s backoff unless `isIntentionallyClosed` is set.
  Don't add a second reconnect loop in views.
- **Auth refresh:** `gameStore.refreshAuthToken()` re-reads `token` from `localStorage`. Call it after a
  login flow before opening the socket.

### Routing & guards

- Routes live in `src/router/index.js`. Lazy-load via dynamic `import()`. The home view (`RoomManager`) is
  the **only** statically imported route component — match this default for new routes (lazy unless
  there's a measured reason).
- Each route declares `meta`:
  - `requiresAuth: true` — gated by `authStore.isAuthenticated`.
  - `guestOnly: true` — redirects authed users to `/admin` (admin) or `/desktop` (player).
  - `requiresAdmin: true` — gated by `authStore.isAdmin`.
  - `titleKey: 'titles.<x>'` — i18n key used by the `afterEach` hook to set `document.title` and update
    it on locale change. **Always supply `titleKey`** for a new route.
- The catch-all `:pathMatch(.*)*` redirects to `/desktop`. Don't change this without checking auth flow.

### i18n

- All user-facing strings go through `vue-i18n`. Add the key to **both** `src/i18n/uk.ts` and
  `src/i18n/en.ts` (uk is the primary; en is fallback). Missing-key warnings are silenced
  (`missingWarn: false`), so add tests by eyeballing the UI in both locales after a change.
- Locale storage key is `'lang'` in `localStorage`; supported set is `['uk', 'en']` — change in **all
  three places** if extended (`main.js`, `i18n.js`, `LanguageSwitcher.vue`).
- For interpolated content with markup, use `<i18n-t keypath="..." scope="global">` with named template
  slots, as in `GameLogPanel.vue` (`#player`, `#target`, `#winner`, `#loser`).
- Error code → message mapping lives in `src/i18n/errorMessages.ts`. New backend error codes are added
  there.

### Styling (Tailwind 4)

- **Tailwind 4** is consumed via `@tailwindcss/vite` and a single `@import "tailwindcss";` in
  `src/style.css`. There is **no `postcss.config.js`** and the legacy `tailwind.config.js` is kept only
  for the `content` glob and the semantic `brand-*` palette referenced by templates.
- **Use semantic colors.** Always reach for `brand-bg`, `brand-surface`, `brand-accent`,
  `brand-text-primary`, etc. Do **not** hard-code hex / `rgb(...)` in templates. Card-specific colors
  are `brand-card-*`.
- **Custom responsive variants** (`short`, `tall`, `xtall`, `square`, `portrait`, `wide`, `retina`) are
  defined in `src/style.css` via `@custom-variant`. Use them instead of arbitrary `@media` queries.
- **Card sizing** is governed by the `--card-primary-h` / `--card-secondary-h` CSS custom properties in
  `src/style.css`. Never hard-code card pixel sizes in component classes — set the CSS variable on the
  closest container if a panel needs a non-default size.
- **Custom scrollbar / animations** (`custom-scrollbar`, `animate-fade-in`, etc.) are global utilities
  defined in `src/style.css`. Reuse them rather than inlining `style="..."` rules.
- **Don't add a CSS framework or CSS-in-JS lib.** Tailwind + a single global CSS file is the contract.
- **Per-component scoped styles** (`<style scoped>`) are allowed but rare. Prefer Tailwind utilities;
  use `<style scoped>` only for animations or selectors Tailwind can't express.

### Comments

- **Do not add comments unless explicitly requested or the logic is genuinely non-obvious.**
- The repo contains many Ukrainian-language comments in store internals (e.g. `gameStore.js`). When
  editing such a file, **match the surrounding language**. Do not translate existing comments to English
  as a side-quest.

### Choosing a library

Before `npm install <lib>`:

1. **Check if Vue / Vite / Pinia / vue-router already provides it.** Most "small lib" needs (date,
   uuid, deep-merge) are either in a `crypto` global, in `Intl`, or already solved by `src/utils/`.
2. **npm last publish.** More than 1–2 years without a release → look for an alternative.
3. **License.** MIT / Apache-2.0 / BSD are fine; GPL / AGPL → ask the user before adding.
4. **Bundle size matters** for an SPA. Check `bundlephobia` for anything > 30 kB gzipped before
   adding.
5. **Vue 3 / Composition-API compatibility is non-negotiable.** A library that only ships a Vue 2
   `Vue.use(...)` plugin is a no-go.

### Bounded inputs

Any path where the client accepts an externally-unbounded amount of data needs an explicit barrier
(file imports, paste, server lists without pagination, WebSocket bursts).

- Define a named `MAX_*` constant next to the call site that enforces it.
- Exceeding the limit must surface an explicit user-facing error (route through `gameStore.error` for
  game-related limits, or a local `ref` for view-local ones) — never silent truncation, never a hang.
- The current `gameLog` in `gameStore.js` is **unbounded** today; if you add a new feature that pushes
  into a store array on every server packet, cap it.

### DO NOT (in this project)

- **No React, Angular, Svelte, or Solid.** This is a Vue 3 SPA.
- **No Options API in new components.** Composition API + `<script setup>` only.
- **No Vuex.** Pinia is the state library.
- **No SSR / Nuxt APIs.** Pure Vite SPA — `import.meta.env`, `window`, and `document` are always
  available at runtime; no `useNuxtApp`, no `definePageMeta`, no server components.
- **No `axios` or other HTTP client libs.** Use native `fetch` and stay consistent with `gameStore.js`.
- **No socket.io / SockJS.** Native `WebSocket` is the contract.
- **No new CSS framework / CSS-in-JS.** Tailwind 4 + `src/style.css` only.
- **No direct `localStorage` writes for app state** beyond the documented `token` and `lang` keys.
  Route session state through `useAuthStore`.
- **No hard-coded backend URLs** like `http://localhost:3000` in views. Use relative `/api/...` and
  rely on the Vite proxy / nginx.
- **No emojis in source files** (templates, scripts, comments). The version banner in `main.js` uses
  styled console output — that pattern is fine.
- **No conversion of `.js` → `.ts`** without an explicit task that includes setting up `tsconfig.json`
  and `vue-tsc` (currently absent on purpose).
- **No deep watchers (`watch(x, fn, { deep: true })`)** over server state. Compute derived fields once
  per packet in the store; see `_updateBoardDerived` in `gameStore.js`.

## Workflow rules for AI

1. **Explore before editing.** Read the closest existing view/component/store and copy its structure.
   Most tasks here are "make it like X" — `BoardView.vue` and `gameStore.js` are the largest references.
2. **Build after edits.** Run `npm run build` after non-trivial changes; it catches Vite resolution
   errors and template syntax issues. Reload the dev server in the browser to confirm the runtime path.
3. **Keep `App.vue` minimal.** It is `<router-view />` only — do not add layout, providers, or
   wrappers there. Add layout in views or in a shared layout component imported by views.
4. **Don't touch unrelated files.** Especially: don't edit `package-lock.json` by hand, don't edit
   `node_modules`, `dist`, `.git`, `.idea`, or `.vite`.
5. **Don't commit, push, branch, amend, or rewrite history unless explicitly told to.** When asked to
   commit, follow the existing concise commit-message style (see `git log --oneline`); confirm the
   branch and the staged files before running `git commit`.
6. **Don't add README/MD docs proactively.** This file is the exception, edited on user request.
7. **Mirror existing patterns over inventing new ones.** When the same problem is solved differently
   in two places (e.g. error handling in a view vs. in `gameStore`), prefer the store-side solution.
8. **When unsure about a backend contract, permission name, or event payload, ask** — do not guess
   wire formats. The truth lives in `src/types/events.ts` and the Go code under `../message-backend/`.
9. **Match the surrounding language for comments** (Ukrainian in `gameStore.js` and parts of
   `router/index.js`; English elsewhere). Don't mass-translate.

## Useful starting points when adding things

- **New route + page:** add a lazy entry in `src/router/index.js` (with `meta.titleKey`), create the
  `.vue` file in `src/views/`, add the title key to `src/i18n/uk.ts` and `src/i18n/en.ts`.
- **New modal:** copy `src/views/ConfirmModal.vue`. Modals are imported by their parent view, not
  routed.
- **New WebSocket message handling:** add a `case` in the `socket.value.onmessage` switch in
  `gameStore.js`, mutate the relevant `ref` (or use `mergeState` for nested updates), and if it's a
  domain event, also add the type to `src/types/events.ts` (`EventType`, `EventTypeMap`, `GameEvent`).
- **New REST endpoint call:** add an `async function` inside `gameStore.js` (or a new dedicated store
  for cross-cutting concerns), use `fetch('/api/...')` with `Authorization: Bearer ${token}`, and
  surface failures by writing to `gameStore.error`.
- **New Pinia store:** copy `src/stores/auth.js` for a small store, or `src/stores/gameStore.js` for a
  store that owns long-running resources.
- **New global utility:** drop a small pure module under `src/utils/` only if 2+ callers need it.

# Agent runtime tips

## Dev environment

- The Vite dev server is started with `npm run dev`. If a dev server is already running, do **not**
  kill or restart it — just reload the browser tab after edits (HMR handles most cases; full reload
  is needed for router or i18n config changes).
- The Go backend must be running on `http://localhost:3000` for `/api` and `/ws` to work. If the
  backend is offline, the WebSocket reconnect loop in `gameStore.js` will keep retrying every 3 s —
  expected, not a bug.

## Login (do this first if you hit `/auth`)

- The dev login flow expects a registered user. If you don't have one, register at `/register`.
- After login, you should be redirected to `/desktop` (player) or `/admin` (admin role). Confirm this
  before proceeding.

## What "verify" means here

After any change, before reporting done:

1. Run `npm run build` (catches build-time errors, template parse failures, missing imports).
2. Reload the affected view in the browser.
3. Open the browser console and report **all** errors and warnings (Vue prop warnings, missing-key
   warnings, unhandled promise rejections).
4. Confirm the WebSocket connected (look for `[WS] Сокет успішно відкрито з бекендом.` in the
   console). If you see repeated reconnect attempts, the backend is down — say so explicitly.
5. If the change is visual (colors, layout, animations, modal positioning), describe the visual result
   or attach a screenshot. The DOM/accessibility tree is not pixel-aware.
6. If the change touches i18n, verify the affected text in both `uk` and `en` via the language
   switcher in the UI.

## Selectors for in-browser checks

- Prefer visible text or `aria-*` attributes for interactive elements. Most icon-only buttons in this
  project rely on `:title` and `aria-label` — use them as selectors.
- If you can't target a UI element reliably, say so instead of guessing.
