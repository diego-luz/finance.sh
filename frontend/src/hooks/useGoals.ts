import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { goalService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';
import { queryKeys } from './queryKeys';
import { ApiRequestError } from '@/lib/axios';
import type { GoalPayload } from '@/types';

export function useGoals() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.goals(orgId),
    queryFn: () => goalService.list(),
    enabled: Boolean(orgId),
  });
}

export function useCreateGoal() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: (payload: GoalPayload) => goalService.create(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.goals(orgId) });
      toast.success('Meta criada com sucesso.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useUpdateGoal() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: GoalPayload }) =>
      goalService.update(id, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.goals(orgId) });
      toast.success('Meta atualizada.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useDeleteGoal() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: (id: string) => goalService.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.goals(orgId) });
      toast.success('Meta removida.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}
