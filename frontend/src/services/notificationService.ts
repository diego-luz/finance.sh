import { api, unwrap } from '@/lib/axios';
import type { ApiEnvelope, Notification } from '@/types';

export const notificationService = {
  list: () =>
    unwrap<Notification[]>(api.get<ApiEnvelope<Notification[]>>('/notifications')),

  markRead: (id: string) => api.post(`/notifications/${id}/read`),

  markAllRead: () => api.post('/notifications/read-all'),
};
