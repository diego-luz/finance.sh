import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

import ptBR from './locales/pt-BR/translation.json';
import en from './locales/en/translation.json';
import es from './locales/es/translation.json';

/** Languages the UI supports. pt-BR is the complete base / fallback. */
export const SUPPORTED_LANGUAGES = ['pt-BR', 'en', 'es'] as const;
export type SupportedLanguage = (typeof SUPPORTED_LANGUAGES)[number];

/** localStorage key the detector caches the chosen language under. */
export const LANGUAGE_STORAGE_KEY = 'finance-sh:lang';

const resources = {
  'pt-BR': { translation: ptBR },
  en: { translation: en },
  es: { translation: es },
} as const;

void i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'pt-BR',
    supportedLngs: [...SUPPORTED_LANGUAGES],
    // NOTE: do NOT add nonExplicitSupportedLngs + load:'currentOnly' — that combo
    // collapses 'pt-BR' to base 'pt' (no resources) and breaks every pt-BR lookup.
    // The default behaviour already resolves regional variants (en-US → en).
    defaultNS: 'translation',
    ns: ['translation'],
    interpolation: {
      // React already escapes values, so i18next must not double-escape.
      escapeValue: false,
    },
    detection: {
      order: ['localStorage', 'navigator'],
      caches: ['localStorage'],
      lookupLocalStorage: LANGUAGE_STORAGE_KEY,
    },
    // Resources are bundled, so init is synchronous and there is nothing to
    // load lazily — disabling Suspense avoids needing a <Suspense> boundary.
    react: {
      useSuspense: false,
    },
    returnNull: false,
  });

export default i18n;
