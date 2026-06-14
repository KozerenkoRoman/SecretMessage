import { createI18n } from 'vue-i18n';
import { uk } from './i18n/uk';
import { en } from './i18n/en';

const SUPPORTED_LOCALES = ['uk', 'en'];
const storedLang = localStorage.getItem('lang');
const initialLocale = SUPPORTED_LOCALES.includes(storedLang) ? storedLang : 'uk';

export const i18n = createI18n({
    legacy: false,
    globalInjection: true,
    locale: initialLocale,
    fallbackLocale: 'en',
    messages: { uk, en },
    missingWarn: false,
    fallbackWarn: false,
});

if (typeof document !== 'undefined') {
    document.documentElement.lang = initialLocale;
}