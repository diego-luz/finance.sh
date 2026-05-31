import { api, unwrap } from '@/lib/axios';
import type {
  ApiEnvelope,
  AuthResponse,
  ChangePasswordPayload,
  LoginPayload,
  LoginResult,
  MeResponse,
  RegisterPayload,
  TwoFactorVerifyPayload,
} from '@/types';

export const authService = {
  register: (payload: RegisterPayload) =>
    unwrap<AuthResponse>(api.post<ApiEnvelope<AuthResponse>>('/auth/register', payload)),

  login: (payload: LoginPayload) =>
    unwrap<LoginResult>(api.post<ApiEnvelope<LoginResult>>('/auth/login', payload)),

  /** Completes a 2FA login challenge, returning a full session. */
  verifyTwoFactor: (payload: TwoFactorVerifyPayload) =>
    unwrap<AuthResponse>(api.post<ApiEnvelope<AuthResponse>>('/auth/2fa/verify', payload)),

  refresh: (refresh_token: string) =>
    unwrap<AuthResponse>(api.post<ApiEnvelope<AuthResponse>>('/auth/refresh', { refresh_token })),

  logout: (refresh_token: string) =>
    api.post('/auth/logout', { refresh_token }),

  me: () => unwrap<MeResponse>(api.get<ApiEnvelope<MeResponse>>('/me')),

  forgotPassword: (email: string) =>
    unwrap<{ email_sent: boolean }>(
      api.post<ApiEnvelope<{ email_sent: boolean }>>('/auth/forgot-password', { email }),
    ),

  resetPassword: (token: string, password: string) =>
    api.post<ApiEnvelope<unknown>>('/auth/reset-password', { token, password }),

  /**
   * Forced/self-service password change for the authenticated user. Backend
   * clears `must_change_password` and revokes all other sessions on success.
   */
  changePassword: (payload: ChangePasswordPayload) =>
    api.post<ApiEnvelope<unknown>>('/me/change-password', payload),

  verifyEmail: (token: string) =>
    api.post<ApiEnvelope<unknown>>('/auth/verify-email', { token }),

  resendVerification: (email: string) =>
    api.post<ApiEnvelope<unknown>>('/auth/verify-email/resend', { email }),

  /**
   * Public check of whether self-service registration is currently open.
   * Used to hide the "Criar conta" link and block the register form.
   */
  registrationOpen: () =>
    unwrap<{ open: boolean }>(
      api.get<ApiEnvelope<{ open: boolean }>>('/auth/registration-open'),
    ),
};
