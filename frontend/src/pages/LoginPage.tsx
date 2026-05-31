import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Link, Navigate } from 'react-router-dom';
import { Mail, Lock, ShieldCheck, ArrowLeft } from 'lucide-react';
import { AuthLayout } from '@/layouts/AuthLayout';
import { Button, Input } from '@/components/ui';
import { useLogin, useVerifyTwoFactorLogin } from '@/hooks';
import { useSetupStatus } from '@/hooks/useSetupStatus';
import { isMfaChallenge } from '@/types';

const loginSchema = z.object({
  email: z.string().min(1, 'Informe o e-mail').email('E-mail inválido'),
  password: z.string().min(1, 'Informe a senha'),
});

type LoginValues = z.infer<typeof loginSchema>;

const codeSchema = z.object({
  code: z
    .string()
    .min(6, 'Digite o código de 6 dígitos')
    .max(6, 'Digite o código de 6 dígitos')
    .regex(/^\d{6}$/, 'O código deve conter apenas números'),
});

type CodeValues = z.infer<typeof codeSchema>;

export function LoginPage() {
  const { t } = useTranslation();
  const login = useLogin();
  const verify = useVerifyTwoFactorLogin();
  // Direct-URL safety net: even though the SetupGate also enforces this rule,
  // a user pasting /login while the platform is uninitialized must end up at
  // the wizard. Reads the same cached probe — no extra HTTP round-trip.
  const { data: setupStatus } = useSetupStatus();
  // Fail-open: treat undefined / errored response as registration being open.
  const [mfaToken, setMfaToken] = useState<string | null>(null);

  if (setupStatus?.needs_setup === true) {
    return <Navigate to="/setup" replace />;
  }

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginValues>({ resolver: zodResolver(loginSchema) });

  const {
    register: registerCode,
    handleSubmit: handleSubmitCode,
    formState: { errors: codeErrors },
  } = useForm<CodeValues>({ resolver: zodResolver(codeSchema) });

  const onSubmit = (values: LoginValues) =>
    login.mutate(values, {
      onSuccess: (res) => {
        if (isMfaChallenge(res)) {
          setMfaToken(res.mfa_token);
        }
      },
    });

  const onSubmitCode = (values: CodeValues) => {
    if (!mfaToken) return;
    verify.mutate({ mfa_token: mfaToken, code: values.code });
  };

  // -------------------------------------------------------------------------
  // Step 2: 2FA code entry
  // -------------------------------------------------------------------------
  if (mfaToken) {
    return (
      <AuthLayout
        title={t('auth.mfa.title')}
        subtitle={t('auth.mfa.subtitle')}
      >
        <form onSubmit={handleSubmitCode(onSubmitCode)} className="space-y-4" noValidate>
          <div className="flex items-center gap-3 rounded-xl border border-primary/30 bg-primary/5 p-3.5">
            <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
              <ShieldCheck className="h-[18px] w-[18px]" />
            </span>
            <p className="text-sm text-gray-600 dark:text-gray-300">
              {t('auth.mfa.protected')}
            </p>
          </div>

          <Input
            label={t('auth.fields.verificationCode')}
            inputMode="numeric"
            autoComplete="one-time-code"
            autoFocus
            maxLength={6}
            placeholder={t('auth.placeholders.code')}
            className="text-center text-lg tracking-[0.5em]"
            error={codeErrors.code?.message}
            {...registerCode('code')}
          />

          <Button type="submit" size="lg" className="w-full" loading={verify.isPending}>
            {t('auth.mfa.submit')}
          </Button>

          <button
            type="button"
            onClick={() => setMfaToken(null)}
            className="inline-flex w-full items-center justify-center gap-1.5 text-sm font-medium text-gray-500 transition hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
          >
            <ArrowLeft className="h-4 w-4" /> {t('auth.mfa.backToLogin')}
          </button>
        </form>
      </AuthLayout>
    );
  }

  // -------------------------------------------------------------------------
  // Step 1: credentials
  // -------------------------------------------------------------------------
  return (
    <AuthLayout title={t('auth.login.title')} subtitle={t('auth.login.subtitle')}>
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <Input
          label={t('auth.fields.email')}
          type="email"
          autoComplete="email"
          placeholder={t('auth.placeholders.email')}
          leftIcon={<Mail className="h-4 w-4" />}
          error={errors.email?.message}
          {...register('email')}
        />
        <div>
          <div className="mb-1.5 flex items-center justify-between">
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
              {t('auth.fields.password')}
            </span>
            <Link
              to="/forgot-password"
              className="text-xs font-medium text-primary hover:text-primary-600"
            >
              {t('auth.login.forgotPassword')}
            </Link>
          </div>
          <Input
            type="password"
            autoComplete="current-password"
            placeholder={t('auth.placeholders.password')}
            leftIcon={<Lock className="h-4 w-4" />}
            error={errors.password?.message}
            {...register('password')}
          />
        </div>

        <Button type="submit" size="lg" className="w-full" loading={login.isPending}>
          {t('auth.login.submit')}
        </Button>
      </form>

      {/* Link to the (separately deployed) marketing landing, only when its URL
          is configured via VITE_LANDING_URL. Empty → hide the link entirely. */}
      {import.meta.env.VITE_LANDING_URL && (
        <p className="mt-6 text-center text-sm text-gray-500 dark:text-gray-400">
          Quer usar o finance.sh?{' '}
          <a
            href={import.meta.env.VITE_LANDING_URL}
            target="_blank"
            rel="noreferrer"
            className="font-medium text-primary hover:text-primary-600"
          >
            Solicitar acesso →
          </a>
        </p>
      )}
    </AuthLayout>
  );
}
