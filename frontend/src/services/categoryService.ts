import { api, unwrap } from '@/lib/axios';
import type { ApiEnvelope, Category, CategoryPayload } from '@/types';

export const categoryService = {
  list: () => unwrap<Category[]>(api.get<ApiEnvelope<Category[]>>('/categories')),

  create: (payload: CategoryPayload) =>
    unwrap<Category>(api.post<ApiEnvelope<Category>>('/categories', payload)),

  update: (id: string, payload: CategoryPayload) =>
    unwrap<Category>(api.put<ApiEnvelope<Category>>(`/categories/${id}`, payload)),

  remove: (id: string) => api.delete(`/categories/${id}`),
};
