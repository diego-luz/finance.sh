import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useSearchParams } from 'react-router-dom';
import { CheckCircle2, XCircle, Loader2, Mail } from 'lucide-react';
import { AuthLayout } from '@/layouts/AuthLayout';
import { useVerifyEmail } from '@/hooks';
import { useAuthStore } from '@/stores/authStore';

type Status = 'verifying' | 'success' | 'error' | 'missing';

export function VerifyEmailPage() {
  const { t } = useTranslation();
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token');
  const verifyEmail = useVerifyEmail();
  const isAuthenticated = useAuthStore((s) => Boolean(s.accessToken));

  const [status, setStatus] = useState<Status>(token ? 'verifying' : 'missing');
  const [message, setMessage] = useState('');
  // Guard against double-invocation (React 18 StrictMode mounts twice in dev).
  const started = useRef(false);

  useEffect(() => {
    if (!token || started.current) return;
    started.current = true;
    verifyEmail.mutate(token, {
      onSuccess: () => setStatus('success'),
      onError: (err) => {
        setStatus('error');
        setMessage(err.message || t('auth.verify.errorFallback'));
      },
    });
    // mutate is stable; run once for the token.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  const content = {
    verifying: {
      icon: <Loader2 className="h-8 w-8 animate-spin text-primary" />,
      tone: 'bg-primary/10',
      title: t('auth.verify.verifyingTitle'),
      text: t('auth.verify.verifyingText'),
    },
    success: {
      icon: <CheckCircle2 className="h-8 w-8 text-primary" />,
      tone: 'bg-primary/10',
      title: t('auth.verify.successTitle'),
      text: t('auth.verify.successText'),
    },
    error: {
      icon: <XCircle className="h-8 w-8 text-red-500" />,
      tone: 'bg-red-500/10',
      title: t('auth.verify.errorTitle'),
      text: message || t('auth.verify.errorText'),
    },
    missing: {
      icon: <Mail className="h-8 w-8 text-amber-500" />,
      tone: 'bg-amber-500/10',
      title: t('auth.verify.missingTitle'),
      text: t('auth.verify.missingText'),
    },
  }[status];

  return (
    <AuthLayout
      title={t('auth.verify.title')}
      subtitle={t('auth.verify.subtitle')}
    >
      <div className="flex flex-col items-center text-center">
        <span
          className={`flex h-16 w-16 items-center justify-center rounded-2xl ${content.tone}`}
        >
          {content.icon}
        </span>
        <h3 className="mt-5 font-heading text-xl font-bold text-gray-900 dark:text-white">
          {content.title}
        </h3>
        <p className="mt-2 max-w-sm text-sm text-gray-500 dark:text-gray-400">{content.text}</p>

        {status !== 'verifying' && (
          <div className="mt-7 w-full">
            <Link
              to={isAuthenticated ? '/' : '/login'}
              className="inline-flex h-12 w-full items-center justify-center rounded-lg bg-primary px-6 text-base font-medium text-white shadow-sm transition-colors hover:bg-primary-600 focus:outline-none focus:ring-2 focus:ring-primary/40"
            >
              {status === 'success'
                ? isAuthenticated
                  ? t('auth.verify.goToDashboard')
                  : t('auth.verify.signIn')
                : isAuthenticated
                  ? t('auth.verify.backToDashboard')
                  : t('auth.verify.backToLogin')}
            </Link>
          </div>
        )}
      </div>
    </AuthLayout>
  );
}
