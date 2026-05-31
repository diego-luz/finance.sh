import { api, unwrap } from '@/lib/axios';
import type { ApiEnvelope, RecurrenceRule, RecurrenceRulePayload } from '@/types';

export const recurrenceService = {
  list: () =>
    unwrap<RecurrenceRule[]>(api.get<ApiEnvelope<RecurrenceRule[]>>('/recurrences')),

  create: (payload: RecurrenceRulePayload) =>
    unwrap<RecurrenceRule>(api.post<ApiEnvelope<RecurrenceRule>>('/recurrences', payload)),

  update: (id: string, payload: RecurrenceRulePayload) =>
    unwrap<RecurrenceRule>(
      api.put<ApiEnvelope<RecurrenceRule>>(`/recurrences/${id}`, payload),
    ),

  remove: (id: string) => api.delete(`/recurrences/${id}`),

  /** Generate any occurrences due now; returns how many were created. */
  run: (id: string) =>
    unwrap<{ created: number }>(
      api.post<ApiEnvelope<{ created: number }>>(`/recurrences/${id}/run`, {}),
    ),
};
