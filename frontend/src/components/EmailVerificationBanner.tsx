import { useState } from 'react';
import { MailWarning, X } from 'lucide-react';
import { Button } from '@/components/ui';
import { useAuthStore } from '@/stores/authStore';
import { useResendVerification } from '@/hooks';

/**
 * Dismissible banner shown in the dashboard while the logged user's e-mail is
 * unverified. Offers a one-click resend of the confirmation link.
 */
export function EmailVerificationBanner() {
  const user = useAuthStore((s) => s.user);
  const resend = useResendVerification();
  const [dismissed, setDismissed] = useState(false);

  if (!user || user.email_verified || dismissed) return null;

  return (
    <div className="mb-4 flex flex-col gap-3 rounded-xl border border-amber-300/70 bg-amber-50 px-4 py-3 dark:border-amber-500/30 dark:bg-amber-500/10 sm:flex-row sm:items-center">
      <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-amber-500/15 text-amber-600 dark:text-amber-400">
        <MailWarning className="h-5 w-5" />
      </span>
      <div className="min-w-0 flex-1">
        <p className="text-sm font-semibold text-amber-800 dark:text-amber-200">
          Confirme seu e-mail
        </p>
        <p className="text-sm text-amber-700/90 dark:text-amber-300/80">
          Enviamos um link de confirmação para <strong>{user.email}</strong>. Confirme para
          manter sua conta segura.
        </p>
      </div>
      <div className="flex shrink-0 items-center gap-1.5">
        <Button
          size="sm"
          variant="outline"
          loading={resend.isPending}
          onClick={() => resend.mutate(user.email)}
          className="border-amber-300 text-amber-800 hover:bg-amber-100 dark:border-amber-500/40 dark:text-amber-200 dark:hover:bg-amber-500/15"
        >
          Reenviar
        </Button>
        <button
          onClick={() => setDismissed(true)}
          className="rounded-lg p-1.5 text-amber-600/70 transition hover:bg-amber-100 hover:text-amber-700 dark:text-amber-400/70 dark:hover:bg-amber-500/15"
          aria-label="Dispensar aviso"
        >
          <X className="h-4 w-4" />
        </button>
      </div>
    </div>
  );
}
