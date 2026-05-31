import { useQuery } from '@tanstack/react-query';
import { dashboardService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { queryKeys } from './queryKeys';

export function useDashboard() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.dashboard(orgId),
    queryFn: () => dashboardService.get(),
    enabled: Boolean(orgId),
  });
}
