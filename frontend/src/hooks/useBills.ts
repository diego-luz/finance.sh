import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { billService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { queryKeys } from './queryKeys';
import type { BillFilters } from '@/types';

export function usePayables(filters: BillFilters, enabled = true) {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.payables(orgId, filters),
    queryFn: () => billService.payables(filters),
    enabled: Boolean(orgId) && enabled,
    placeholderData: keepPreviousData,
  });
}

export function useReceivables(filters: BillFilters, enabled = true) {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.receivables(orgId, filters),
    queryFn: () => billService.receivables(filters),
    enabled: Boolean(orgId) && enabled,
    placeholderData: keepPreviousData,
  });
}
