import { api, unwrap } from '@/lib/axios';
import type { ApiEnvelope, SearchResponse } from '@/types';

export const searchService = {
  /** Global search across the main entities. `limit` caps hits per group. */
  search: (q: string, limit = 5) =>
    unwrap<SearchResponse>(
      api.get<ApiEnvelope<SearchResponse>>('/search', { params: { q, limit } }),
    ),
};
