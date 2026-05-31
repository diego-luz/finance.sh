import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type CookieConsent = 'accepted' | 'rejected';

interface ConsentState {
  /** null = the visitor hasn't made a choice yet (banner should show). */
  cookieConsent: CookieConsent | null;
  /** Persist an explicit choice and dismiss the banner. */
  setCookieConsent: (value: CookieConsent) => void;
}

export const useConsentStore = create<ConsentState>()(
  persist(
    (set) => ({
      cookieConsent: null,
      setCookieConsent: (value) => set({ cookieConsent: value }),
    }),
    {
      name: 'finance-sh-consent',
      partialize: (s) => ({ cookieConsent: s.cookieConsent }),
    },
  ),
);
