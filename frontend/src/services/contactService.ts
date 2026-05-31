import { api, unwrap } from '@/lib/axios';
import type { ApiEnvelope, Contact, ContactPayload } from '@/types';

export const contactService = {
  list: () => unwrap<Contact[]>(api.get<ApiEnvelope<Contact[]>>('/contacts')),

  get: (id: string) => unwrap<Contact>(api.get<ApiEnvelope<Contact>>(`/contacts/${id}`)),

  create: (payload: ContactPayload) =>
    unwrap<Contact>(api.post<ApiEnvelope<Contact>>('/contacts', payload)),

  update: (id: string, payload: ContactPayload) =>
    unwrap<Contact>(api.put<ApiEnvelope<Contact>>(`/contacts/${id}`, payload)),

  remove: (id: string) => api.delete(`/contacts/${id}`),
};
