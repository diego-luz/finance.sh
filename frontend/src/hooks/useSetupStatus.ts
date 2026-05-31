import { useQuery } from '@tanstack/react-query';
import { getSetupStatus, type SetupStatus } from '@/lib/api/setup';
import { ApiRequestError } from '@/lib/axios';

/** Centralized React Query key for the setup probe. */
export const setupStatusQueryKey = ['setup', 'status'] as const;

/**
 * Public, single-shot probe of `GET /api/v1/setup/status`. The result is
 * sticky (5-minute stale time, no refocus refetch, no mount refetch) because
 * once the platform is initialized it cannot un-initialize, and a fresh
 * install only flips to `false` after a successful POST /setup/initialize
 * (which the mutation invalidates explicitly).
 */
export function useSetupStatus() {
  return useQuery<SetupStatus, ApiRequestError>({
    queryKey: setupStatusQueryKey,
    queryFn: getSetupStatus,
    staleTime: 5 * 60_000,
    gcTime: 30 * 60_000,
    refetchOnWindowFocus: false,
    refetchOnMount: false,
    refetchOnReconnect: false,
    // Fail-open: don't retry hard so the boot gate can decide quickly.
    retry: 1,
  });
}
