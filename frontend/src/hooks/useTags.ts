import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { tagService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';
import { queryKeys } from './queryKeys';
import { ApiRequestError } from '@/lib/axios';
import type { TagPayload } from '@/types';

export function useTags() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.tags(orgId),
    queryFn: () => tagService.list(),
    enabled: Boolean(orgId),
  });
}

/** Invalidate tags and transactions (tags render alongside transactions). */
function useInvalidateTags() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  return () => {
    queryClient.invalidateQueries({ queryKey: queryKeys.tags(orgId) });
    queryClient.invalidateQueries({ queryKey: ['transactions', orgId] });
  };
}

export function useCreateTag() {
  const invalidate = useInvalidateTags();
  const toast = useToast();
  return useMutation({
    mutationFn: (payload: TagPayload) => tagService.create(payload),
    onSuccess: () => {
      invalidate();
      toast.success('Tag criada.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useUpdateTag() {
  const invalidate = useInvalidateTags();
  const toast = useToast();
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: TagPayload }) =>
      tagService.update(id, payload),
    onSuccess: () => {
      invalidate();
      toast.success('Tag atualizada.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useDeleteTag() {
  const invalidate = useInvalidateTags();
  const toast = useToast();
  return useMutation({
    mutationFn: (id: string) => tagService.remove(id),
    onSuccess: () => {
      invalidate();
      toast.success('Tag removida.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}
