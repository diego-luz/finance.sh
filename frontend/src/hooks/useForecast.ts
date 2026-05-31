import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { forecastService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { queryKeys } from './queryKeys';

export function useForecast(months: number) {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.forecast(orgId, months),
    queryFn: () => forecastService.get(months),
    enabled: Boolean(orgId),
    placeholderData: keepPreviousData,
  });
}
