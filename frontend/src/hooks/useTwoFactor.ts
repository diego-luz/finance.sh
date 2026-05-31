import { useMutation, useQueryClient } from '@tanstack/react-query';
import { twoFactorService } from '@/services';
import { useToast } from '@/contexts/ToastContext';
import { useAuthStore } from '@/stores/authStore';
import { ApiRequestError } from '@/lib/axios';
import type { TwoFactorEnableResponse, TwoFactorSetupResponse } from '@/types';
import { queryKeys } from './queryKeys';

/** Reflect the 2FA flag on the locally-stored user (auth store is the source of truth). */
function setTwoFactorFlag(enabled: boolean) {
  const { user, setUser } = useAuthStore.getState();
  if (user) setUser({ ...user, two_factor_enabled: enabled });
}

/** Starts 2FA enrollment: fetches the TOTP secret + otpauth URL. */
export function useTwoFactorSetup() {
  const toast = useToast();
  return useMutation<TwoFactorSetupResponse, ApiRequestError, void>({
    mutationFn: () => twoFactorService.setup(),
    onError: (err: ApiRequestError) =>
      toast.error(err.message || 'Não foi possível iniciar a configuração.'),
  });
}

/** Confirms enrollment and returns recovery codes. Invalidates /me. */
export function useEnableTwoFactor() {
  const queryClient = useQueryClient();
  const toast = useToast();
  return useMutation<TwoFactorEnableResponse, ApiRequestError, string>({
    mutationFn: (code: string) => twoFactorService.enable(code),
    onSuccess: () => {
      toast.success('Autenticação em dois fatores ativada.');
      setTwoFactorFlag(true);
      queryClient.invalidateQueries({ queryKey: queryKeys.me });
    },
    onError: (err: ApiRequestError) =>
      toast.error(err.message || 'Código inválido. Tente novamente.'),
  });
}

/** Disables 2FA after verifying a current code. Invalidates /me. */
export function useDisableTwoFactor() {
  const queryClient = useQueryClient();
  const toast = useToast();
  return useMutation<unknown, ApiRequestError, string>({
    mutationFn: (code: string) => twoFactorService.disable(code),
    onSuccess: () => {
      toast.success('Autenticação em dois fatores desativada.');
      setTwoFactorFlag(false);
      queryClient.invalidateQueries({ queryKey: queryKeys.me });
    },
    onError: (err: ApiRequestError) =>
      toast.error(err.message || 'Não foi possível desativar. Verifique o código.'),
  });
}
