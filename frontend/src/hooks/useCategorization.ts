import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { categorizationService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';
import { queryKeys } from './queryKeys';
import { ApiRequestError } from '@/lib/axios';
import type { CategoryRulePayload } from '@/types';

export function useCategoryRules() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.categorizationRules(orgId),
    queryFn: () => categorizationService.listRules(),
    enabled: Boolean(orgId),
  });
}

export function useCreateRule() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: (payload: CategoryRulePayload) => categorizationService.createRule(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.categorizationRules(orgId) });
      toast.success('Regra criada.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useUpdateRule() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: CategoryRulePayload }) =>
      categorizationService.updateRule(id, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.categorizationRules(orgId) });
      toast.success('Regra atualizada.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useDeleteRule() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: (id: string) => categorizationService.deleteRule(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.categorizationRules(orgId) });
      toast.success('Regra removida.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useApplyCategorization() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: () => categorizationService.apply(),
    onSuccess: ({ updated }) => {
      queryClient.invalidateQueries({ queryKey: ['transactions', orgId] });
      queryClient.invalidateQueries({ queryKey: ['payables', orgId] });
      queryClient.invalidateQueries({ queryKey: ['receivables', orgId] });
      queryClient.invalidateQueries({ queryKey: queryKeys.dashboard(orgId) });
      toast.success(
        `${updated} transaç${updated === 1 ? 'ão categorizada' : 'ões categorizadas'}.`,
      );
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}
