import { Cookie } from 'lucide-react';
import { Button } from '@/components/ui';
import { useConsentStore } from '@/stores/consentStore';

/**
 * Bottom cookie/consent bar shown until the visitor makes a choice. The choice
 * is persisted in localStorage (zustand), so it appears once until dismissed.
 */
export function CookieConsentBanner() {
  const cookieConsent = useConsentStore((s) => s.cookieConsent);
  const setCookieConsent = useConsentStore((s) => s.setCookieConsent);

  if (cookieConsent !== null) return null;

  return (
    <div
      role="dialog"
      aria-label="Aviso de cookies"
      className="fixed inset-x-0 bottom-0 z-[90] animate-fade-in p-3 sm:p-4"
    >
      <div className="mx-auto flex max-w-4xl flex-col gap-3 rounded-xl border border-gray-200 bg-white/95 p-4 shadow-card backdrop-blur dark:border-ink-border dark:bg-ink-surface/95 sm:flex-row sm:items-center">
        <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
          <Cookie className="h-5 w-5" />
        </span>
        <p className="flex-1 text-sm text-gray-600 dark:text-gray-300">
          Utilizamos cookies para autenticação e para melhorar sua experiência. Você pode aceitar
          ou recusar os cookies opcionais. Saiba mais na nossa{' '}
          {/* Plain anchor: this banner renders OUTSIDE the RouterProvider tree,
              so a react-router <Link> would crash (null router context). */}
          <a
            href="/privacidade"
            className="font-medium text-primary hover:text-primary-600"
          >
            Política de Privacidade
          </a>
          .
        </p>
        <div className="flex shrink-0 gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setCookieConsent('rejected')}
          >
            Recusar
          </Button>
          <Button size="sm" onClick={() => setCookieConsent('accepted')}>
            Aceitar
          </Button>
        </div>
      </div>
    </div>
  );
}
