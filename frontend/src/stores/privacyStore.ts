import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface PrivacyState {
  /** When true, all monetary values are blurred/redacted across the app. */
  hidden: boolean;
  toggle: () => void;
  setHidden: (v: boolean) => void;
}

/** Toggle the global `values-hidden` class on <html> immediately. */
function applyPrivacy(hidden: boolean) {
  document.documentElement.classList.toggle('values-hidden', hidden);
}

export const usePrivacyStore = create<PrivacyState>()(
  persist(
    (set, get) => ({
      hidden: false,
      toggle: () => {
        const next = !get().hidden;
        applyPrivacy(next);
        set({ hidden: next });
      },
      setHidden: (v) => {
        applyPrivacy(v);
        set({ hidden: v });
      },
    }),
    {
      name: 'finance-sh:privacy',
      onRehydrateStorage: () => (state) => {
        // Sync the DOM class with the persisted value on load.
        applyPrivacy(state?.hidden ?? false);
      },
    },
  ),
);
