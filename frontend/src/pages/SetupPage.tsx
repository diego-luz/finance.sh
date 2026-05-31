import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useNavigate } from 'react-router-dom';
import {
  Mail,
  Lock,
  User as UserIcon,
  Building2,
  Wand2,
  Copy,
  Check,
  ArrowRight,
  ArrowLeft,
  Sparkles,
  Rocket,
  CircleCheck,
} from 'lucide-react';
import { AuthLayout } from '@/layouts/AuthLayout';
import { Button, Input } from '@/components/ui';
import { CurrencySelect } from '@/components/CurrencySelect';
import { useInitializeSetup } from '@/hooks/useInitializeSetup';
import { useSetupStatus } from '@/hooks/useSetupStatus';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';
import { suggestPassword } from '@/lib/password';
import { copyToClipboard } from '@/lib/clipboard';
import { SUPPORTED_CURRENCIES } from '@/lib/currency';
import { ApiRequestError } from '@/lib/axios';
import { cn } from '@/lib/cn';

const currencyCodes = SUPPORTED_CURRENCIES.map((c) => c.code) as [
  string,
  ...string[],
];

// ---------------------------------------------------------------------------
// Schemas — split by step so per-step validation gates the "Next" button.
// ---------------------------------------------------------------------------
const userStepSchema = z
  .object({
    name: z.string().min(2, 'auth.errors.nameMin'),
    email: z.string().min(1, 'auth.errors.emailRequired').email('auth.errors.emailInvalid'),
    password: z.string().min(8, 'auth.errors.passwordMin'),
    confirm_password: z.string().min(1, 'auth.errors.confirmRequired'),
  })
  .refine((v) => v.password === v.confirm_password, {
    message: 'auth.errors.passwordsMismatch',
    path: ['confirm_password'],
  });

const orgStepSchema = z.object({
  organization_name: z.string().min(2, 'auth.errors.orgNameMin'),
  currency: z.enum(currencyCodes),
});

const fullSchema = z
  .object({
    name: z.string().min(2),
    email: z.string().email(),
    password: z.string().min(8),
    confirm_password: z.string(),
    organization_name: z.string().min(2),
    currency: z.enum(currencyCodes),
  })
  .refine((v) => v.password === v.confirm_password, {
    path: ['confirm_password'],
  });

type FormValues = z.infer<typeof fullSchema>;

type Step = 1 | 2 | 3;

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Display label for a currency code, falling back to the bare code. */
function currencyLabel(code: string): string {
  const c = SUPPORTED_CURRENCIES.find((x) => x.code === code);
  return c ? `${c.code} — ${c.name}` : code;
}

/**
 * Resolve the error message coming back from RHF — the schema stores i18n
 * keys, so translate when present, otherwise show the raw string.
 */
function useFieldError() {
  const { t } = useTranslation();
  return (msg: string | undefined) => {
    if (!msg) return undefined;
    return msg.includes('.') ? t(msg) : msg;
  };
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------
export function SetupPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const toast = useToast();
  const accessToken = useAuthStore((s) => s.accessToken);
  const { data: status } = useSetupStatus();
  const initialize = useInitializeSetup();
  const fieldError = useFieldError();

  const [step, setStep] = useState<Step>(1);
  const [copied, setCopied] = useState(false);

  // Already initialized: bounce to login. Authenticated visitor: bounce home.
  useEffect(() => {
    if (accessToken) {
      navigate('/', { replace: true });
      return;
    }
    if (status && status.needs_setup === false) {
      navigate('/login', { replace: true });
    }
  }, [accessToken, status, navigate]);

  const {
    register,
    handleSubmit,
    trigger,
    getValues,
    setValue,
    watch,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(fullSchema),
    defaultValues: {
      name: '',
      email: '',
      password: '',
      confirm_password: '',
      organization_name: '',
      currency: 'BRL',
    },
    mode: 'onBlur',
  });

  const values = watch();

  // -------------------------------------------------------------------------
  // Password helpers (reuse the same UX as the admin reset modal).
  // -------------------------------------------------------------------------
  const suggest = () => {
    const pw = suggestPassword(16);
    setValue('password', pw, { shouldValidate: true, shouldDirty: true });
    setValue('confirm_password', pw, { shouldValidate: true, shouldDirty: true });
  };

  const copy = async () => {
    const pw = getValues('password');
    if (!pw) {
      toast.error(t('setup.errors.emptyPassword'));
      return;
    }
    const ok = await copyToClipboard(pw);
    if (ok) {
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } else {
      toast.error(t('setup.errors.copyFailed'));
    }
  };

  // -------------------------------------------------------------------------
  // Step navigation: validate ONLY the current step's fields before advancing.
  // -------------------------------------------------------------------------
  const goNext = async () => {
    if (step === 1) {
      // Run partial validation against the user step schema so we don't
      // surface organization-step errors prematurely.
      const partial = userStepSchema.safeParse({
        name: values.name,
        email: values.email,
        password: values.password,
        confirm_password: values.confirm_password,
      });
      // Always trigger RHF so the inline error messages render.
      const ok = await trigger(['name', 'email', 'password', 'confirm_password']);
      if (!ok || !partial.success) return;
      setStep(2);
      return;
    }
    if (step === 2) {
      const partial = orgStepSchema.safeParse({
        organization_name: values.organization_name,
        currency: values.currency,
      });
      const ok = await trigger(['organization_name', 'currency']);
      if (!ok || !partial.success) return;
      setStep(3);
      return;
    }
  };

  const goPrev = () => {
    if (step === 1) return;
    setStep((step - 1) as Step);
  };

  // -------------------------------------------------------------------------
  // Submission — wired to step 3 only. Maps RHF values to the backend payload.
  // -------------------------------------------------------------------------
  const onSubmit = (v: FormValues) =>
    initialize.mutate(
      {
        user: { name: v.name, email: v.email, password: v.password },
        organization: { name: v.organization_name, currency: v.currency },
      },
      {
        onError: (err: ApiRequestError) => {
          if (err.status === 409) {
            toast.error(t('setup.errors.already_initialized'));
            navigate('/login', { replace: true });
            return;
          }
          if (err.status === 400) {
            toast.error(err.message || t('setup.errors.generic'));
            return;
          }
          toast.error(err.message || t('setup.errors.generic'));
        },
      },
    );

  // -------------------------------------------------------------------------
  // Render
  // -------------------------------------------------------------------------
  return (
    <AuthLayout title={t('setup.title')} subtitle={t('setup.subtitle')}>
      <ProgressBar step={step} />

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        {step === 1 && (
          <>
            <Input
              label={t('setup.steps.user.fields.name')}
              autoComplete="name"
              placeholder={t('auth.placeholders.name')}
              leftIcon={<UserIcon className="h-4 w-4" />}
              error={fieldError(errors.name?.message)}
              {...register('name')}
            />
            <Input
              label={t('setup.steps.user.fields.email')}
              type="email"
              autoComplete="email"
              placeholder={t('auth.placeholders.email')}
              leftIcon={<Mail className="h-4 w-4" />}
              error={fieldError(errors.email?.message)}
              {...register('email')}
            />

            <div className="flex items-center justify-between gap-2 pt-1">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={suggest}
              >
                <Wand2 className="h-4 w-4" /> {t('setup.actions.suggestPassword')}
              </Button>
              <Button type="button" variant="ghost" size="sm" onClick={copy}>
                {copied ? (
                  <Check className="h-4 w-4 text-primary" />
                ) : (
                  <Copy className="h-4 w-4" />
                )}
                {copied ? t('setup.actions.copied') : t('setup.actions.copy')}
              </Button>
            </div>

            <Input
              label={t('setup.steps.user.fields.password')}
              type="text"
              autoComplete="new-password"
              placeholder={t('auth.placeholders.passwordMin')}
              leftIcon={<Lock className="h-4 w-4" />}
              error={fieldError(errors.password?.message)}
              hint={t('setup.steps.user.passwordHint')}
              {...register('password')}
            />
            <Input
              label={t('setup.steps.user.fields.confirmPassword')}
              type="password"
              autoComplete="new-password"
              leftIcon={<Lock className="h-4 w-4" />}
              error={fieldError(errors.confirm_password?.message)}
              {...register('confirm_password')}
            />
          </>
        )}

        {step === 2 && (
          <>
            <Input
              label={t('setup.steps.org.fields.name')}
              placeholder={t('auth.placeholders.organizationName')}
              leftIcon={<Building2 className="h-4 w-4" />}
              error={fieldError(errors.organization_name?.message)}
              {...register('organization_name')}
            />
            <CurrencySelect
              label={t('setup.steps.org.fields.currency')}
              error={fieldError(errors.currency?.message)}
              {...register('currency')}
            />
            <p className="rounded-xl border border-primary/20 bg-primary/5 p-3.5 text-sm text-gray-600 dark:text-gray-300">
              <Sparkles className="mr-1.5 inline h-3.5 w-3.5 text-primary" />
              {t('setup.steps.org.hint')}
            </p>
          </>
        )}

        {step === 3 && (
          <div className="space-y-4">
            <ReviewBlock
              title={t('setup.steps.review.adminTitle')}
              rows={[
                { label: t('setup.steps.user.fields.name'), value: values.name },
                { label: t('setup.steps.user.fields.email'), value: values.email },
                {
                  label: t('setup.steps.user.fields.password'),
                  value: '•'.repeat(Math.min(values.password.length || 8, 16)),
                },
              ]}
            />
            <ReviewBlock
              title={t('setup.steps.review.orgTitle')}
              rows={[
                {
                  label: t('setup.steps.org.fields.name'),
                  value: values.organization_name,
                },
                {
                  label: t('setup.steps.org.fields.currency'),
                  value: currencyLabel(values.currency),
                },
              ]}
            />
            <p className="rounded-xl border border-amber-500/30 bg-amber-500/5 p-3.5 text-xs text-amber-700 dark:text-amber-400">
              {t('setup.steps.review.notice')}
            </p>
          </div>
        )}

        <div className="flex items-center justify-between gap-3 pt-2">
          <Button
            type="button"
            variant="ghost"
            size="md"
            onClick={goPrev}
            disabled={step === 1 || initialize.isPending}
            className={cn(step === 1 && 'invisible')}
          >
            <ArrowLeft className="h-4 w-4" /> {t('setup.actions.previous')}
          </Button>

          {step < 3 ? (
            <Button type="button" size="lg" onClick={goNext}>
              {t('setup.actions.next')} <ArrowRight className="h-4 w-4" />
            </Button>
          ) : (
            <Button type="submit" size="lg" loading={initialize.isPending}>
              <Rocket className="h-4 w-4" /> {t('setup.actions.finish')}
            </Button>
          )}
        </div>
      </form>
    </AuthLayout>
  );
}

// ---------------------------------------------------------------------------
// Sub-components
// ---------------------------------------------------------------------------

function ProgressBar({ step }: { step: Step }) {
  const { t } = useTranslation();
  const steps: Array<{ n: Step; label: string }> = [
    { n: 1, label: t('setup.steps.user.short') },
    { n: 2, label: t('setup.steps.org.short') },
    { n: 3, label: t('setup.steps.review.short') },
  ];
  return (
    <div className="mb-6">
      <div className="mb-3 flex items-center justify-between text-xs font-medium text-gray-500 dark:text-gray-400">
        <span>{t('setup.stepCounter', { current: step, total: 3 })}</span>
        <span className="text-primary">{steps[step - 1].label}</span>
      </div>
      <div className="flex items-center gap-2">
        {steps.map((s) => {
          const done = s.n < step;
          const active = s.n === step;
          return (
            <div key={s.n} className="flex flex-1 items-center gap-2">
              <div
                className={cn(
                  'flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-xs font-semibold transition-colors',
                  done && 'bg-primary text-white',
                  active && 'bg-primary/15 text-primary ring-2 ring-primary/30',
                  !done && !active &&
                    'bg-gray-100 text-gray-400 dark:bg-ink-elevated dark:text-gray-500',
                )}
              >
                {done ? <CircleCheck className="h-4 w-4" /> : s.n}
              </div>
              <div
                className={cn(
                  'h-1 flex-1 rounded-full transition-colors',
                  done
                    ? 'bg-primary'
                    : 'bg-gray-100 dark:bg-ink-elevated',
                  // Hide the trailing connector on the last item.
                  s.n === 3 && 'hidden',
                )}
              />
            </div>
          );
        })}
      </div>
    </div>
  );
}

function ReviewBlock({
  title,
  rows,
}: {
  title: string;
  rows: Array<{ label: string; value: string }>;
}) {
  return (
    <div className="rounded-xl border border-gray-200 bg-white p-4 dark:border-ink-border dark:bg-ink-elevated">
      <p className="mb-3 text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
        {title}
      </p>
      <dl className="space-y-2 text-sm">
        {rows.map((r) => (
          <div key={r.label} className="flex items-center justify-between gap-3">
            <dt className="text-gray-500 dark:text-gray-400">{r.label}</dt>
            <dd className="max-w-[60%] truncate text-right font-medium text-gray-900 dark:text-gray-100">
              {r.value || '—'}
            </dd>
          </div>
        ))}
      </dl>
    </div>
  );
}
