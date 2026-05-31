import { api, unwrap } from '@/lib/axios';
import type {
  ApiEnvelope,
  BulkResult,
  CategoryRule,
  CategoryRulePayload,
  SuggestResponse,
} from '@/types';

/** Transaction kinds the suggestion endpoint understands. */
export type SuggestType = 'expense' | 'income';

export const categorizationService = {
  listRules: () =>
    unwrap<CategoryRule[]>(
      api.get<ApiEnvelope<CategoryRule[]>>('/categorization/rules'),
    ),

  createRule: (payload: CategoryRulePayload) =>
    unwrap<CategoryRule>(
      api.post<ApiEnvelope<CategoryRule>>('/categorization/rules', payload),
    ),

  updateRule: (id: string, payload: CategoryRulePayload) =>
    unwrap<CategoryRule>(
      api.put<ApiEnvelope<CategoryRule>>(`/categorization/rules/${id}`, payload),
    ),

  deleteRule: (id: string) => api.delete(`/categorization/rules/${id}`),

  apply: () =>
    unwrap<BulkResult>(
      api.post<ApiEnvelope<BulkResult>>('/categorization/apply', {}),
    ),

  suggest: (description: string, type: SuggestType) =>
    unwrap<SuggestResponse>(
      api.get<ApiEnvelope<SuggestResponse>>('/categorization/suggest', {
        params: { description, type },
      }),
    ),
};
