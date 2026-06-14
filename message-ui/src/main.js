import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n';
import { uk } from './i18n/uk';
import { en } from './i18n/en';
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

const i18n = createI18n({
    legacy: false,
    globalInjection: true,
    locale: initialLocale,
    fallbackLocale: 'en',
    messages: { uk, en },
    missingWarn: false,
    fallbackWarn: false,
})

if (typeof document !== 'undefined') {
    document.documentElement.lang = initialLocale
}

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(i18n)

app.mount('#app')