import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query';
import { budgetService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';
import { queryKeys } from './queryKeys';
import { ApiRequestError } from '@/lib/axios';
import type { BudgetFilters, BudgetPayload } from '@/types';

export function useBudgets(filters: BudgetFilters) {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.budgets(orgId, filters),
    queryFn: () => budgetService.list(filters),
    enabled: Boolean(orgId),
    placeholderData: keepPreviousData,
  });
}

function useInvalidateBudgets() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: ['budgets', orgId] });
}

export function useCreateBudget() {
  const invalidate = useInvalidateBudgets();
  const toast = useToast();
  return useMutation({
    mutationFn: (payload: BudgetPayload) => budgetService.create(payload),
    onSuccess: () => {
      invalidate();
      toast.success('Orçamento criado.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useUpdateBudget() {
  const invalidate = useInvalidateBudgets();
  const toast = useToast();
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: BudgetPayload }) =>
      budgetService.update(id, payload),
    onSuccess: () => {
      invalidate();
      toast.success('Orçamento atualizado.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useDeleteBudget() {
  const invalidate = useInvalidateBudgets();
  const toast = useToast();
  return useMutation({
    mutationFn: (id: string) => budgetService.remove(id),
    onSuccess: () => {
      invalidate();
      toast.success('Orçamento removido.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}
