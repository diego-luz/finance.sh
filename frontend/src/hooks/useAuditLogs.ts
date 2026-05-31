import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { auditService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { queryKeys } from './queryKeys';
import type { AuditFilters } from '@/types';

/** Org-scoped, paginated audit-log listing (read-only; owner/admin only). */
export function useAuditLogs(filters: AuditFilters) {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.auditLogs(orgId, filters),
    queryFn: () => auditService.list(filters),
    enabled: Boolean(orgId),
    placeholderData: keepPreviousData,
  });
}
