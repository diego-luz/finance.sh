import { useTranslation } from 'react-i18next';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Link } from 'react-router-dom';
import { Mail, Lock, User, Building2, ShieldAlert } from 'lucide-react';
import { AuthLayout } from '@/layouts/AuthLayout';
import { Button, Input } from '@/components/ui';
import { CurrencySelect } from '@/components/CurrencySelect';
import { useRegister, useRegistrationOpen } from '@/hooks';
import { SUPPORTED_CURRENCIES } from '@/lib/currency';
import { cn } from '@/lib/cn';

const currencyCodes = SUPPORTED_CURRENCIES.map((c) => c.code) as [string, ...string[]];

const schema = z.object({
  name: z.string().min(2, 'Informe seu nome completo'),
  organization_name: z.string().min(2, 'Informe o nome da empresa'),
  email: z.string().min(1, 'Informe o e-mail').email('E-mail inválido'),
  password: z.string().min(8, 'A senha deve ter ao menos 8 caracteres'),
  currency: z.enum(currencyCodes),
  accepted_terms: z.literal(true, {
    errorMap: () => ({ message: 'Você precisa aceitar os termos para continuar.' }),
  }),
});

type FormValues = z.infer<typeof schema>;

export function RegisterPage() {
  const { t } = useTranslation();
  const registerMut = useRegister();
  const { data: registration } = useRegistrationOpen();
  // Fail-open: only block the form when the backend explicitly reports closed.
  const registrationClosed = registration?.open === false;
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { currency: 'BRL' },
  });

  const onSubmit = (values: FormValues) =>
    registerMut.mutate({
      name: values.name,
      organization_name: values.organization_name,
      email: values.email,
      password: values.password,
      currency: values.currency,
      accepted_terms: true,
    });

  if (registrationClosed) {
    return (
      <AuthLayout
        title={t('auth.register.title')}
        subtitle={t('auth.register.subtitle')}
      >
        <div className="flex flex-col items-center gap-4 rounded-xl border border-amber-500/30 bg-amber-500/5 p-6 text-center">
          <span className="flex h-12 w-12 items-center justify-center rounded-2xl bg-amber-500/10 text-amber-500">
            <ShieldAlert className="h-6 w-6" />
          </span>
          <div>
            <p className="font-heading text-base font-semibold text-gray-900 dark:text-gray-100">
              Cadastro fechado
            </p>
            <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
              O cadastro de novas contas está temporariamente indisponível. Contate o
              administrador para solicitar acesso.
            </p>
          </div>
        </div>
        <p className="mt-6 text-center text-sm text-gray-500 dark:text-gray-400">
          {t('auth.register.hasAccount')}{' '}
          <Link to="/login" className="font-medium text-primary hover:text-primary-600">
            {t('auth.register.signIn')}
          </Link>
        </p>
      </AuthLayout>
    );
  }

  return (
    <AuthLayout
      title={t('auth.register.title')}
      subtitle={t('auth.register.subtitle')}
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <Input
          label={t('auth.fields.name')}
          autoComplete="name"
          placeholder={t('auth.placeholders.name')}
          leftIcon={<User className="h-4 w-4" />}
          error={errors.name?.message}
          {...register('name')}
        />
        <Input
          label={t('auth.fields.organizationName')}
          placeholder={t('auth.placeholders.organizationName')}
          leftIcon={<Building2 className="h-4 w-4" />}
          error={errors.organization_name?.message}
          {...register('organization_name')}
        />
        <CurrencySelect
          label={t('auth.fields.organizationCurrency')}
          error={errors.currency?.message}
          {...register('currency')}
        />
        <Input
          label={t('auth.fields.email')}
          type="email"
          autoComplete="email"
          placeholder={t('auth.placeholders.email')}
          leftIcon={<Mail className="h-4 w-4" />}
          error={errors.email?.message}
          {...register('email')}
        />
        <Input
          label={t('auth.fields.password')}
          type="password"
          autoComplete="new-password"
          placeholder={t('auth.placeholders.passwordMin')}
          leftIcon={<Lock className="h-4 w-4" />}
          error={errors.password?.message}
          {...register('password')}
        />

        <div>
          <label className="flex cursor-pointer items-start gap-2.5 text-sm text-gray-600 dark:text-gray-300">
            <input
              type="checkbox"
              className={cn(
                'mt-0.5 h-4 w-4 shrink-0 rounded border-gray-300 text-primary',
                'focus:ring-2 focus:ring-primary/40 focus:ring-offset-0 dark:border-ink-border dark:bg-ink-elevated',
                errors.accepted_terms && 'border-red-500',
              )}
              aria-invalid={errors.accepted_terms ? true : undefined}
              {...register('accepted_terms')}
            />
            <span>
              {t('auth.register.acceptTermsPrefix')}{' '}
              <Link
                to="/termos"
                target="_blank"
                rel="noopener noreferrer"
                className="font-medium text-primary hover:text-primary-600"
              >
                {t('auth.register.termsOfUse')}
              </Link>{' '}
              {t('auth.register.and')}{' '}
              <Link
                to="/privacidade"
                target="_blank"
                rel="noopener noreferrer"
                className="font-medium text-primary hover:text-primary-600"
              >
                {t('auth.register.privacyPolicy')}
              </Link>
              .
            </span>
          </label>
          {errors.accepted_terms && (
            <p className="mt-1 text-xs text-red-500">{errors.accepted_terms.message}</p>
          )}
        </div>

        <Button type="submit" size="lg" className="w-full" loading={registerMut.isPending}>
          {t('auth.register.submit')}
        </Button>
      </form>

      <p className="mt-6 text-center text-sm text-gray-500 dark:text-gray-400">
        {t('auth.register.hasAccount')}{' '}
        <Link to="/login" className="font-medium text-primary hover:text-primary-600">
          {t('auth.register.signIn')}
        </Link>
      </p>
    </AuthLayout>
  );
}
