import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { recurrenceService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';
import { queryKeys } from './queryKeys';
import { ApiRequestError } from '@/lib/axios';
import type { RecurrenceRulePayload } from '@/types';

export function useRecurrences() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.recurrences(orgId),
    queryFn: () => recurrenceService.list(),
    enabled: Boolean(orgId),
  });
}

export function useCreateRecurrence() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: (payload: RecurrenceRulePayload) => recurrenceService.create(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.recurrences(orgId) });
      toast.success('Recorrência criada.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useUpdateRecurrence() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: RecurrenceRulePayload }) =>
      recurrenceService.update(id, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.recurrences(orgId) });
      toast.success('Recorrência atualizada.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useDeleteRecurrence() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: (id: string) => recurrenceService.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.recurrences(orgId) });
      toast.success('Recorrência removida.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useRunRecurrence() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: (id: string) => recurrenceService.run(id),
    onSuccess: ({ created }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.recurrences(orgId) });
      queryClient.invalidateQueries({ queryKey: ['transactions', orgId] });
      queryClient.invalidateQueries({ queryKey: queryKeys.dashboard(orgId) });
      queryClient.invalidateQueries({ queryKey: ['payables', orgId] });
      queryClient.invalidateQueries({ queryKey: ['receivables', orgId] });
      // Generated entries change balances and the projection.
      queryClient.invalidateQueries({ queryKey: ['accounts', orgId] });
      queryClient.invalidateQueries({ queryKey: ['forecast', orgId] });
      toast.success(
        `${created} lançamento${created === 1 ? '' : 's'} gerado${created === 1 ? '' : 's'}.`,
      );
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}
