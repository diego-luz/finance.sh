import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { registerSW } from 'virtual:pwa-register';
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
