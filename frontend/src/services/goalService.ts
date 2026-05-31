import { api, unwrap } from '@/lib/axios';
import type { ApiEnvelope, Goal, GoalPayload } from '@/types';

export const goalService = {
  list: () => unwrap<Goal[]>(api.get<ApiEnvelope<Goal[]>>('/goals')),

  create: (payload: GoalPayload) =>
    unwrap<Goal>(api.post<ApiEnvelope<Goal>>('/goals', payload)),

  update: (id: string, payload: GoalPayload) =>
    unwrap<Goal>(api.put<ApiEnvelope<Goal>>(`/goals/${id}`, payload)),

  remove: (id: string) => api.delete(`/goals/${id}`),
};
