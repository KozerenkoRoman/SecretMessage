import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { i18n } from './i18n';
import App from './App.vue'
import router from './router'
import './style.css'
import faviconUrl from './assets/favicon.ico?url'

const SUPPORTED_LOCALES = ['uk', 'en']
const storedLang = localStorage.getItem('lang')
const initialLocale = SUPPORTED_LOCALES.includes(storedLang) ? storedLang : 'uk'

const link = document.querySelector("link[rel~='icon']") || document.createElement('link');
link.type = 'image/x-icon';
link.rel = 'icon'; link.href = faviconUrl;
document.getElementsByTagName('head')[0].appendChild(link);

if (typeof document !== 'undefined') {
    document.documentElement.lang = initialLocale
}

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)
app.use(router)
app.use(i18n)

app.mount('#app')

console.log(
    `%c MESSAGE UI v${__APP_VERSION__} %c Production`,
    'background: #1e1e2f; color: #00ffcc; padding: 3px 5px; border-radius: 3px; font-weight: bold;',
    'color: #888;'
);