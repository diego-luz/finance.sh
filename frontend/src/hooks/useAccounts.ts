import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { accountService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';
import { queryKeys } from './queryKeys';
import { ApiRequestError } from '@/lib/axios';
import type { AccountPayload } from '@/types';

export function useAccounts() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.accounts(orgId),
    queryFn: () => accountService.list(),
    enabled: Boolean(orgId),
  });
}

export function useCreateAccount() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: (payload: AccountPayload) => accountService.create(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.accounts(orgId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.dashboard(orgId) });
      toast.success('Conta criada com sucesso.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useUpdateAccount() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: AccountPayload }) =>
      accountService.update(id, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.accounts(orgId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.dashboard(orgId) });
      toast.success('Conta atualizada.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useDeleteAccount() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: (id: string) => accountService.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.accounts(orgId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.dashboard(orgId) });
      toast.success('Conta removida.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}
