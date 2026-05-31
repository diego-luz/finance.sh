import { useEffect, useState } from 'react';
import { Download, X } from 'lucide-react';
import { Button } from '@/components/ui';

/**
 * `beforeinstallprompt` is not part of lib.dom, so we declare the minimal shape
 * we rely on. It extends Event with the deferred-prompt API exposed by Chromium.
 */
interface BeforeInstallPromptEvent extends Event {
  readonly platforms: string[];
  readonly userChoice: Promise<{
    outcome: 'accepted' | 'dismissed';
    platform: string;
  }>;
  prompt(): Promise<void>;
}

const DISMISS_KEY = 'finance-sh:install-prompt-dismissed';

/**
 * Dismissible "Add to Home Screen" banner. Captures the browser's deferred
 * `beforeinstallprompt` event and surfaces a branded install CTA. Hidden once
 * the user installs or dismisses (dismissal persisted in localStorage).
 *
 * Renders OUTSIDE the RouterProvider tree — it uses no router hooks, so a plain
 * element is safe here (mirrors CookieConsentBanner).
 */
export function InstallPrompt() {
  const [deferredPrompt, setDeferredPrompt] =
    useState<BeforeInstallPromptEvent | null>(null);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    // Respect a previous dismissal.
    if (localStorage.getItem(DISMISS_KEY) === '1') return;

    const onBeforeInstall = (event: Event) => {
      // Prevent Chrome's default mini-infobar; stash the event for later.
      event.preventDefault();
      setDeferredPrompt(event as BeforeInstallPromptEvent);
      setVisible(true);
    };

    const onInstalled = () => {
      setVisible(false);
      setDeferredPrompt(null);
    };

    window.addEventListener('beforeinstallprompt', onBeforeInstall);
    window.addEventListener('appinstalled', onInstalled);

    return () => {
      window.removeEventListener('beforeinstallprompt', onBeforeInstall);
      window.removeEventListener('appinstalled', onInstalled);
    };
  }, []);

  const dismiss = () => {
    localStorage.setItem(DISMISS_KEY, '1');
    setVisible(false);
  };

  const install = async () => {
    if (!deferredPrompt) return;
    await deferredPrompt.prompt();
    const { outcome } = await deferredPrompt.userChoice;
    // The prompt can only be used once; drop it regardless of the outcome.
    setDeferredPrompt(null);
    setVisible(false);
    if (outcome === 'dismissed') {
      // Don't nag again this session/device once explicitly declined.
      localStorage.setItem(DISMISS_KEY, '1');
    }
  };

  if (!visible || !deferredPrompt) return null;

  return (
    <div
      role="dialog"
      aria-label="Instalar o finance.sh"
      className="fixed inset-x-0 bottom-0 z-[95] animate-fade-in p-3 sm:p-4"
    >
      <div className="mx-auto flex max-w-md flex-col gap-3 rounded-xl border border-gray-200 bg-white/95 p-4 shadow-card backdrop-blur dark:border-ink-border dark:bg-ink-surface/95 sm:flex-row sm:items-center">
        <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
          <Download className="h-5 w-5" />
        </span>
        <div className="min-w-0 flex-1">
          <p className="text-sm font-semibold text-gray-900 dark:text-gray-100">
            Instalar o finance.sh
          </p>
          <p className="text-sm text-gray-600 dark:text-gray-300">
            Adicione o finance.sh à tela inicial para acesso rápido e offline.
          </p>
        </div>
        <div className="flex shrink-0 gap-2">
          <Button variant="ghost" size="sm" onClick={dismiss}>
            Agora não
          </Button>
          <Button size="sm" onClick={install}>
            <Download className="h-4 w-4" />
            Instalar
          </Button>
        </div>
        <button
          onClick={dismiss}
          aria-label="Fechar"
          className="absolute right-2 top-2 rounded-md p-1 text-gray-400 transition hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-ink-elevated sm:hidden"
        >
          <X className="h-4 w-4" />
        </button>
      </div>
    </div>
  );
}
