import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { categoryService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';
import { queryKeys } from './queryKeys';
import { ApiRequestError } from '@/lib/axios';
import type { CategoryPayload } from '@/types';

export function useCategories() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.categories(orgId),
    queryFn: () => categoryService.list(),
    enabled: Boolean(orgId),
  });
}

export function useCreateCategory() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: (payload: CategoryPayload) => categoryService.create(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.categories(orgId) });
      toast.success('Categoria criada.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useUpdateCategory() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: CategoryPayload }) =>
      categoryService.update(id, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.categories(orgId) });
      toast.success('Categoria atualizada.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useDeleteCategory() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: (id: string) => categoryService.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.categories(orgId) });
      toast.success('Categoria removida.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}
