import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n';
import { uk } from './i18n/uk';
import { en } from './i18n/en';
import App from './App.vue'
import router from './router'
import './style.css'

const SUPPORTED_LOCALES = ['uk', 'en']
const storedLang = localStorage.getItem('lang')
const initialLocale = SUPPORTED_LOCALES.includes(storedLang) ? storedLang : 'uk'

const i18n = createI18n({
    legacy: false,
    globalInjection: true,
    locale: initialLocale,
    fallbackLocale: 'en',
    messages: { uk, en },
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