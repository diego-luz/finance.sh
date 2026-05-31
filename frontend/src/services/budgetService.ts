import { api, unwrap } from '@/lib/axios';
import type { ApiEnvelope, Budget, BudgetFilters, BudgetPayload } from '@/types';

export const budgetService = {
  list: (filters: BudgetFilters) =>
    unwrap<Budget[]>(
      api.get<ApiEnvelope<Budget[]>>('/budgets', {
        params: { month: filters.month, year: filters.year },
      }),
    ),

  create: (payload: BudgetPayload) =>
    unwrap<Budget>(api.post<ApiEnvelope<Budget>>('/budgets', payload)),

  update: (id: string, payload: BudgetPayload) =>
    unwrap<Budget>(api.put<ApiEnvelope<Budget>>(`/budgets/${id}`, payload)),

  remove: (id: string) => api.delete(`/budgets/${id}`),
};
