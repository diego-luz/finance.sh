import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useNavigate } from 'react-router-dom';
import { Lock, KeyRound, LogOut, ShieldAlert } from 'lucide-react';
import { AuthLayout } from '@/layouts/AuthLayout';
import { Button, Input } from '@/components/ui';
import { useChangePassword, useLogout } from '@/hooks';
import { useAuthStore } from '@/stores/authStore';

const schema = z
  .object({
    current_password: z.string().min(1, 'Informe a senha atual'),
    new_password: z.string().min(8, 'A nova senha deve ter ao menos 8 caracteres'),
    confirm: z.string().min(1, 'Confirme a nova senha'),
  })
  .refine((v) => v.new_password === v.confirm, {
    message: 'As senhas não coincidem',
    path: ['confirm'],
  })
  .refine((v) => v.new_password !== v.current_password, {
    message: 'A nova senha deve ser diferente da atual',
    path: ['new_password'],
  });

type FormValues = z.infer<typeof schema>;

/**
 * Forced (and self-service) password change page. Used right after first
 * login when the backend sets `must_change_password = true`. The page lives
 * outside any sidebar/dashboard chrome so it can't be visually escaped — only
 * the explicit "Sair" button releases the user.
 *
 * Hard-blocks the browser back button: while `must_change_password` is true,
 * `popstate` re-pushes the user onto /change-password.
 */
export function ChangePasswordPage() {
  const user = useAuthStore((s) => s.user);
  const navigate = useNavigate();
  const change = useChangePassword();
  const logout = useLogout();

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({ resolver: zodResolver(schema) });

  // Block back-navigation while the change is pending/required. The router's
  // history is a Stack, so re-pushing the same path is enough to defeat the
  // browser back button without trapping the user inside the SPA.
  useEffect(() => {
    if (!user?.must_change_password) return;
    const onPop = () => {
      // Re-pin the URL; React Router will treat this as a no-op route change.
      window.history.pushState(null, '', '/change-password');
    };
    window.history.pushState(null, '', '/change-password');
    window.addEventListener('popstate', onPop);
    return () => window.removeEventListener('popstate', onPop);
  }, [user?.must_change_password]);

  const onSubmit = (values: FormValues) => {
    if (change.isPending) return;
    change.mutate(
      {
        current_password: values.current_password,
        new_password: values.new_password,
      },
      {
        onSuccess: () => {
          // Super-admins land in the back-office; regular users in the app.
          navigate(user?.super_admin ? '/admin' : '/', { replace: true });
        },
      },
    );
  };

  const forced = user?.must_change_password === true;
  const submitting = change.isPending || isSubmitting;

  return (
    <AuthLayout
      title={forced ? 'Defina sua nova senha' : 'Alterar senha'}
      subtitle={
        forced
          ? 'Por segurança, é necessário trocar a senha inicial antes de continuar.'
          : 'Atualize sua senha de acesso.'
      }
    >
      {forced && (
        <div className="mb-5 flex items-start gap-3 rounded-xl border border-amber-300/60 bg-amber-50 p-3.5 dark:border-amber-500/30 dark:bg-amber-500/10">
          <ShieldAlert className="mt-0.5 h-5 w-5 shrink-0 text-amber-600 dark:text-amber-400" />
          <p className="text-sm text-gray-700 dark:text-gray-200">
            Sua conta foi provisionada por um administrador. Escolha uma senha pessoal
            antes de acessar a plataforma. Outras sessões serão encerradas.
          </p>
        </div>
      )}

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <Input
          label="Senha atual"
          type="password"
          autoComplete="current-password"
          autoFocus
          placeholder="Senha recebida do administrador"
          leftIcon={<KeyRound className="h-4 w-4" />}
          error={errors.current_password?.message}
          {...register('current_password')}
        />
        <Input
          label="Nova senha"
          type="password"
          autoComplete="new-password"
          placeholder="Mínimo 8 caracteres"
          leftIcon={<Lock className="h-4 w-4" />}
          error={errors.new_password?.message}
          {...register('new_password')}
        />
        <Input
          label="Confirmar nova senha"
          type="password"
          autoComplete="new-password"
          placeholder="Repita a nova senha"
          leftIcon={<Lock className="h-4 w-4" />}
          error={errors.confirm?.message}
          {...register('confirm')}
        />

        <Button
          type="submit"
          size="lg"
          className="w-full"
          loading={submitting}
          disabled={submitting}
        >
          Alterar senha
        </Button>
      </form>

      <button
        type="button"
        onClick={() => logout.mutate(undefined)}
        className="mt-6 inline-flex w-full items-center justify-center gap-1.5 text-sm font-medium text-gray-500 transition hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
      >
        <LogOut className="h-4 w-4" /> Sair
      </button>
    </AuthLayout>
  );
}
