import { api, unwrapPaginated } from '@/lib/axios';
import type { ApiEnvelope, AuditFilters, AuditLog, Paginated } from '@/types';

function buildParams(filters: AuditFilters): Record<string, string | number> {
  const params: Record<string, string | number> = {};
  if (filters.action) params.action = filters.action;
  if (filters.entity) params.entity = filters.entity;
  if (filters.user_id) params.user_id = filters.user_id;
  if (filters.from) params.from = filters.from;
  if (filters.to) params.to = filters.to;
  if (filters.page) params.page = filters.page;
  if (filters.per_page) params.per_page = filters.per_page;
  return params;
}

export const auditService = {
  list: (filters: AuditFilters): Promise<Paginated<AuditLog>> =>
    unwrapPaginated<AuditLog>(
      api.get<ApiEnvelope<AuditLog[]>>('/audit-logs', { params: buildParams(filters) }),
    ),
};
