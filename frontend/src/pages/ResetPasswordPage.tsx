import { useTranslation } from 'react-i18next';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Link, useSearchParams } from 'react-router-dom';
import { Lock, ArrowLeft, AlertCircle } from 'lucide-react';
import { AuthLayout } from '@/layouts/AuthLayout';
import { Button, Input } from '@/components/ui';
import { useResetPassword } from '@/hooks';

const schema = z
  .object({
    password: z.string().min(8, 'A senha deve ter ao menos 8 caracteres'),
    confirm: z.string().min(1, 'Confirme a senha'),
  })
  .refine((v) => v.password === v.confirm, {
    message: 'As senhas não coincidem',
    path: ['confirm'],
  });

type FormValues = z.infer<typeof schema>;

export function ResetPasswordPage() {
  const { t } = useTranslation();
  const [params] = useSearchParams();
  const token = params.get('token') ?? '';
  const reset = useResetPassword();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormValues>({ resolver: zodResolver(schema) });

  const onSubmit = (values: FormValues) => {
    reset.mutate({ token, password: values.password });
  };

  if (!token) {
    return (
      <AuthLayout title={t('auth.reset.invalidTitle')} subtitle={t('auth.reset.invalidSubtitle')}>
        <div className="rounded-xl border border-red-200 bg-red-50 p-5 dark:border-red-500/30 dark:bg-red-500/10">
          <div className="mb-3 flex h-11 w-11 items-center justify-center rounded-xl bg-red-500/10 text-red-500">
            <AlertCircle className="h-5 w-5" />
          </div>
          <p className="text-sm text-gray-600 dark:text-gray-300">
            {t('auth.reset.invalidBody')}
          </p>
        </div>
        <Link
          to="/forgot-password"
          className="mt-6 inline-flex items-center gap-1.5 text-sm font-medium text-primary hover:text-primary-600"
        >
          <ArrowLeft className="h-4 w-4" /> {t('auth.reset.requestNew')}
        </Link>
      </AuthLayout>
    );
  }

  return (
    <AuthLayout title={t('auth.reset.title')} subtitle={t('auth.reset.subtitle')}>
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <Input
          label={t('auth.fields.newPassword')}
          type="password"
          autoComplete="new-password"
          placeholder={t('auth.placeholders.password')}
          leftIcon={<Lock className="h-4 w-4" />}
          error={errors.password?.message}
          {...register('password')}
        />
        <Input
          label={t('auth.fields.confirmPassword')}
          type="password"
          autoComplete="new-password"
          placeholder={t('auth.placeholders.password')}
          leftIcon={<Lock className="h-4 w-4" />}
          error={errors.confirm?.message}
          {...register('confirm')}
        />
        <Button type="submit" size="lg" className="w-full" loading={reset.isPending}>
          {t('auth.reset.submit')}
        </Button>
      </form>

      <Link
        to="/login"
        className="mt-6 inline-flex items-center gap-1.5 text-sm font-medium text-primary hover:text-primary-600"
      >
        <ArrowLeft className="h-4 w-4" /> {t('auth.reset.backToLogin')}
      </Link>
    </AuthLayout>
  );
}
