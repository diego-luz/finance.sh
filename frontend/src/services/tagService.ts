import { api, unwrap } from '@/lib/axios';
import type { ApiEnvelope, Tag, TagPayload } from '@/types';

export const tagService = {
  list: () => unwrap<Tag[]>(api.get<ApiEnvelope<Tag[]>>('/tags')),

  create: (payload: TagPayload) =>
    unwrap<Tag>(api.post<ApiEnvelope<Tag>>('/tags', payload)),

  update: (id: string, payload: TagPayload) =>
    unwrap<Tag>(api.put<ApiEnvelope<Tag>>(`/tags/${id}`, payload)),

  remove: (id: string) => api.delete(`/tags/${id}`),
};
