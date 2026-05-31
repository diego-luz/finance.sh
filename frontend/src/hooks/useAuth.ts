import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { authService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';
import { ApiRequestError } from '@/lib/axios';
import { isMfaChallenge } from '@/types';
import type {
  AuthResponse,
  ChangePasswordPayload,
  LoginPayload,
  LoginResult,
  RegisterPayload,
  TwoFactorVerifyPayload,
} from '@/types';
import { queryKeys } from './queryKeys';

export function useForgotPassword() {
  return useMutation({
    mutationFn: (email: string) => authService.forgotPassword(email),
  });
}

export function useResetPassword() {
  const navigate = useNavigate();
  const toast = useToast();
  return useMutation({
    mutationFn: ({ token, password }: { token: string; password: string }) =>
      authService.resetPassword(token, password),
    onSuccess: () => {
      toast.success('Senha redefinida com sucesso. Faça login novamente.');
      navigate('/login', { replace: true });
    },
    onError: (err: ApiRequestError) =>
      toast.error(err.message || 'Não foi possível redefinir a senha.'),
  });
}

/**
 * Login mutation. Resolves to either a full session (handled here) or a 2FA
 * challenge (returned to the caller so the LoginPage can render its code step).
 * Returns the {@link LoginResult} so callers can branch on `mfa_required`.
 */
export function useLogin() {
  const setAuth = useAuthStore((s) => s.setAuth);
  const navigate = useNavigate();
  const toast = useToast();

  return useMutation<LoginResult, ApiRequestError, LoginPayload>({
    mutationFn: (payload: LoginPayload) => authService.login(payload),
    onSuccess: (res) => {
      if (isMfaChallenge(res)) {
        // Defer: the LoginPage transitions to the 2FA step. No toast/redirect.
        return;
      }
      setAuth(res);
      toast.success(`Bem-vindo de volta, ${res.user.name.split(' ')[0]}!`);
      // Forced password change takes priority over any other landing route.
      if (res.user.must_change_password) {
        navigate('/change-password', { replace: true });
        return;
      }
      // Platform super-admins have no org and the tenant app needs one — send
      // them straight to the back-office.
      navigate(res.user.super_admin ? '/admin' : '/', { replace: true });
    },
    onError: (err: ApiRequestError) => {
      toast.error(err.message || 'Não foi possível entrar.');
    },
  });
}

/** Completes a 2FA login challenge with the user's 6-digit code. */
export function useVerifyTwoFactorLogin() {
  const setAuth = useAuthStore((s) => s.setAuth);
  const navigate = useNavigate();
  const toast = useToast();

  return useMutation<AuthResponse, ApiRequestError, TwoFactorVerifyPayload>({
    mutationFn: (payload: TwoFactorVerifyPayload) => authService.verifyTwoFactor(payload),
    onSuccess: (res) => {
      setAuth(res);
      toast.success(`Bem-vindo de volta, ${res.user.name.split(' ')[0]}!`);
      if (res.user.must_change_password) {
        navigate('/change-password', { replace: true });
        return;
      }
      navigate(res.user.super_admin ? '/admin' : '/', { replace: true });
    },
    onError: (err: ApiRequestError) => {
      toast.error(err.message || 'Código inválido. Tente novamente.');
    },
  });
}

export function useRegister() {
  const setAuth = useAuthStore((s) => s.setAuth);
  const navigate = useNavigate();
  const toast = useToast();

  return useMutation<AuthResponse, ApiRequestError, RegisterPayload>({
    mutationFn: (payload: RegisterPayload) => authService.register(payload),
    onSuccess: (res) => {
      setAuth(res);
      toast.success('Conta criada com sucesso!');
      navigate('/', { replace: true });
    },
    onError: (err: ApiRequestError) => {
      toast.error(err.message || 'Não foi possível criar a conta.');
    },
  });
}

/** Verifies an email confirmation token (from the /verify-email page). */
export function useVerifyEmail() {
  const queryClient = useQueryClient();
  const setUser = useAuthStore((s) => s.setUser);
  const user = useAuthStore((s) => s.user);

  return useMutation<unknown, ApiRequestError, string>({
    mutationFn: (token: string) => authService.verifyEmail(token),
    onSuccess: () => {
      // Reflect the verified state immediately when a session is present.
      if (user && !user.email_verified) {
        setUser({ ...user, email_verified: true });
      }
      queryClient.invalidateQueries({ queryKey: queryKeys.me });
    },
  });
}

/** Resends the email verification link to the given address. */
export function useResendVerification() {
  const toast = useToast();
  return useMutation<unknown, ApiRequestError, string>({
    mutationFn: (email: string) => authService.resendVerification(email),
    onSuccess: () => {
      toast.success('Enviamos um novo e-mail de confirmação. Verifique sua caixa de entrada.');
    },
    onError: (err: ApiRequestError) =>
      toast.error(err.message || 'Não foi possível reenviar o e-mail.'),
  });
}

/**
 * Forced/self-service password change for the authenticated user. On success:
 *  - clears local `must_change_password` flag (instant UI reflection),
 *  - invalidates the `me` query so the next /me reflects the backend truth,
 *  - shows a toast.
 *
 * The caller decides where to navigate (typically /admin for super-admin,
 * otherwise /), so this hook stays reusable from settings as well.
 */
export function useChangePassword() {
  const queryClient = useQueryClient();
  const setUser = useAuthStore((s) => s.setUser);
  const toast = useToast();

  return useMutation<unknown, ApiRequestError, ChangePasswordPayload>({
    mutationFn: (payload: ChangePasswordPayload) => authService.changePassword(payload),
    onSuccess: () => {
      const current = useAuthStore.getState().user;
      if (current) {
        setUser({ ...current, must_change_password: false });
      }
      queryClient.invalidateQueries({ queryKey: queryKeys.me });
      toast.success('Senha alterada com sucesso.');
    },
    onError: (err: ApiRequestError) => {
      // 401 from this endpoint specifically means the *current* password is wrong.
      if (err.status === 401) {
        toast.error('Senha atual incorreta.');
        return;
      }
      toast.error(err.message || 'Não foi possível alterar a senha.');
    },
  });
}

export function useLogout() {
  const doLogout = useAuthStore((s) => s.logout);
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const rt = useAuthStore.getState().refreshToken;
      if (rt) {
        try {
          await authService.logout(rt);
        } catch {
          // Ignore network errors on logout; we clear local state regardless.
        }
      }
    },
    onSettled: () => {
      doLogout();
      queryClient.clear();
      navigate('/login', { replace: true });
    },
  });
}
