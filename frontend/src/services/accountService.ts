import { api, unwrap } from '@/lib/axios';
import type { Account, AccountPayload, ApiEnvelope } from '@/types';

export const accountService = {
  list: () => unwrap<Account[]>(api.get<ApiEnvelope<Account[]>>('/accounts')),

  create: (payload: AccountPayload) =>
    unwrap<Account>(api.post<ApiEnvelope<Account>>('/accounts', payload)),

  update: (id: string, payload: AccountPayload) =>
    unwrap<Account>(api.put<ApiEnvelope<Account>>(`/accounts/${id}`, payload)),

  remove: (id: string) => api.delete(`/accounts/${id}`),
};
