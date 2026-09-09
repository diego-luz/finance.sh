import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { registerSW } from 'virtual:pwa-register';

// Fonts are bundled and served by this instance, never fetched from Google.
// A self-hosted app that pulls its typefaces from fonts.gstatic.com sends every
// visitor's IP and User-Agent to a third party on each page load, and renders
// with fallback faces when the instance has no internet access — neither of
// which fits a self-hosted, LGPD-minded product.
//
// Only the latin subsets are imported: the UI ships pt-BR, en and es, so the
// cyrillic/greek/vietnamese faces would be dead weight in the bundle.
// Weights match tailwind.config.js: Inter 400-700 (sans), Manrope 500-800 (heading).
import '@fontsource/inter/latin-400.css';
import '@fontsource/inter/latin-500.css';
import '@fontsource/inter/latin-600.css';
import '@fontsource/inter/latin-700.css';
import '@fontsource/inter/latin-ext-400.css';
import '@fontsource/inter/latin-ext-500.css';
import '@fontsource/inter/latin-ext-600.css';
import '@fontsource/inter/latin-ext-700.css';
import '@fontsource/manrope/latin-500.css';
import '@fontsource/manrope/latin-600.css';
import '@fontsource/manrope/latin-700.css';
import '@fontsource/manrope/latin-800.css';
import '@fontsource/manrope/latin-ext-500.css';
import '@fontsource/manrope/latin-ext-600.css';
import '@fontsource/manrope/latin-ext-700.css';
import '@fontsource/manrope/latin-ext-800.css';

import './i18n';
import App from './App';
import './index.css';

// Register the service worker. With registerType: 'autoUpdate' the plugin
// activates new versions automatically; we reload once a fresh build is ready
// so users never run a stale bundle (financial data must reflect the latest UI).
const updateSW = registerSW({
  immediate: true,
  onNeedRefresh() {
    updateSW(true);
  },
});

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
