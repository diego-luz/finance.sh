import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { reportService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { queryKeys } from './queryKeys';
import type { DateRange } from '@/types';

/** Aggregated report summary for the Relatórios page, scoped by org + range. */
export function useReportSummary(range: DateRange) {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.reportSummary(orgId, { from: range.from, to: range.to }),
    queryFn: () => reportService.summary({ from: range.from, to: range.to }),
    enabled: Boolean(orgId),
    placeholderData: keepPreviousData,
  });
}
