import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { searchService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { queryKeys } from './queryKeys';
import { useDebouncedValue } from './useDebouncedValue';

/**
 * Debounced global search. The query is only enabled once the (trimmed)
 * search term has at least 2 characters, and previous results are kept while
 * a new term resolves to avoid flicker in the command palette.
 */
export function useSearch(query: string, limit = 5) {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const debounced = useDebouncedValue(query.trim(), 300);
  const enabled = Boolean(orgId) && debounced.length >= 2;

  return useQuery({
    queryKey: queryKeys.search(orgId, debounced),
    queryFn: () => searchService.search(debounced, limit),
    enabled,
    placeholderData: keepPreviousData,
  });
}
