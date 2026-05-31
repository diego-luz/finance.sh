import { api, unwrap, unwrapPaginated } from '@/lib/axios';
import type {
  ApiEnvelope,
  BulkCategorizePayload,
  BulkResult,
  BulkSettlePayload,
  DeleteScope,
  Paginated,
  SettlePayload,
  Transaction,
  TransactionFilters,
  TransactionPayload,
} from '@/types';

function buildParams(filters: TransactionFilters): Record<string, string | number> {
  const params: Record<string, string | number> = {};
  if (filters.type) params.type = filters.type;
  if (filters.account_id) params.account_id = filters.account_id;
  if (filters.category_id) params.category_id = filters.category_id;
  if (filters.tag_id) params.tag_id = filters.tag_id;
  if (filters.from) params.from = filters.from;
  if (filters.to) params.to = filters.to;
  if (filters.search) params.search = filters.search;
  if (filters.page) params.page = filters.page;
  if (filters.per_page) params.per_page = filters.per_page;
  return params;
}

export const transactionService = {
  list: (filters: TransactionFilters): Promise<Paginated<Transaction>> =>
    unwrapPaginated<Transaction>(
      api.get<ApiEnvelope<Transaction[]>>('/transactions', { params: buildParams(filters) }),
    ),

  create: (payload: TransactionPayload) =>
    unwrap<Transaction>(api.post<ApiEnvelope<Transaction>>('/transactions', payload)),

  update: (id: string, payload: TransactionPayload) =>
    unwrap<Transaction>(api.put<ApiEnvelope<Transaction>>(`/transactions/${id}`, payload)),

  remove: (id: string, scope: DeleteScope = 'one') =>
    api.delete(`/transactions/${id}`, { params: { scope } }),

  settle: (id: string, payload: SettlePayload) =>
    unwrap<Transaction>(
      api.post<ApiEnvelope<Transaction>>(`/transactions/${id}/settle`, payload),
    ),

  unsettle: (id: string) =>
    unwrap<Transaction>(api.post<ApiEnvelope<Transaction>>(`/transactions/${id}/unsettle`, {})),

  bulkSettle: (payload: BulkSettlePayload) =>
    unwrap<BulkResult>(
      api.post<ApiEnvelope<BulkResult>>('/transactions/bulk-settle', payload),
    ),

  bulkUnsettle: (ids: string[]) =>
    unwrap<BulkResult>(
      api.post<ApiEnvelope<BulkResult>>('/transactions/bulk-unsettle', { ids }),
    ),

  bulkCategorize: (payload: BulkCategorizePayload) =>
    unwrap<BulkResult>(
      api.post<ApiEnvelope<BulkResult>>('/transactions/bulk-categorize', payload),
    ),

  bulkDelete: (ids: string[]) =>
    unwrap<BulkResult>(
      api.post<ApiEnvelope<BulkResult>>('/transactions/bulk-delete', { ids }),
    ),
};
