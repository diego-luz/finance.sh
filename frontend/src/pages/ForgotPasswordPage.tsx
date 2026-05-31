import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Link } from 'react-router-dom';
import { Mail, ArrowLeft, MailCheck, ServerCog } from 'lucide-react';
import { AuthLayout } from '@/layouts/AuthLayout';
import { Button, Input } from '@/components/ui';
import { useForgotPassword } from '@/hooks';

const schema = z.object({
  email: z.string().min(1, 'Informe o e-mail').email('E-mail inválido'),
});

type FormValues = z.infer<typeof schema>;

export function ForgotPasswordPage() {
  const { t } = useTranslation();
  const [sent, setSent] = useState(false);
  // Whether the instance actually e-mails (SMTP configured). Drives the message.
  const [emailSent, setEmailSent] = useState(true);
  const forgot = useForgotPassword();
  const {
    register,
    handleSubmit,
    getValues,
    formState: { errors },
  } = useForm<FormValues>({ resolver: zodResolver(schema) });

  // The endpoint always returns a generic 200 to avoid account enumeration, so
  // we show the confirmation screen regardless of the response outcome.
  const onSubmit = (values: FormValues) => {
    forgot.mutate(values.email, {
      onSuccess: (data) => setEmailSent(data.email_sent),
      onSettled: () => setSent(true),
    });
  };

  if (sent) {
    // SMTP configured → "check your e-mail". No SMTP → tell the user the link is
    // not e-mailed (operator gets it from the server log / reset CLI).
    return (
      <AuthLayout
        title={emailSent ? t('auth.forgot.sentTitle') : t('auth.forgot.noEmailTitle')}
        subtitle={emailSent ? t('auth.forgot.sentSubtitle') : t('auth.forgot.noEmailSubtitle')}
      >
        <div className="rounded-xl border border-gray-200 bg-gray-50 p-5 dark:border-ink-border dark:bg-ink-elevated">
          <div className="mb-3 flex h-11 w-11 items-center justify-center rounded-xl bg-primary/10 text-primary">
            {emailSent ? <MailCheck className="h-5 w-5" /> : <ServerCog className="h-5 w-5" />}
          </div>
          {emailSent ? (
            <p className="text-sm text-gray-600 dark:text-gray-300">
              {t('auth.forgot.sentBodyPrefix')}{' '}
              <span className="font-medium text-gray-900 dark:text-gray-100">
                {getValues('email')}
              </span>
              {t('auth.forgot.sentBodySuffix')}
            </p>
          ) : (
            <p className="text-sm text-gray-600 dark:text-gray-300">{t('auth.forgot.noEmailBody')}</p>
          )}
        </div>
        <Link
          to="/login"
          className="mt-6 inline-flex items-center gap-1.5 text-sm font-medium text-primary hover:text-primary-600"
        >
          <ArrowLeft className="h-4 w-4" /> {t('auth.forgot.backToLogin')}
        </Link>
      </AuthLayout>
    );
  }

  return (
    <AuthLayout
      title={t('auth.forgot.title')}
      subtitle={t('auth.forgot.subtitle')}
    >
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
        <Button type="submit" size="lg" className="w-full" loading={forgot.isPending}>
          {t('auth.forgot.submit')}
        </Button>
      </form>

      <Link
        to="/login"
        className="mt-6 inline-flex items-center gap-1.5 text-sm font-medium text-primary hover:text-primary-600"
      >
        <ArrowLeft className="h-4 w-4" /> {t('auth.forgot.backToLogin')}
      </Link>
    </AuthLayout>
  );
}
