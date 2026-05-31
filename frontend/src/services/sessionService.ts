import { api, unwrap } from '@/lib/axios';
import type { ApiEnvelope, Session } from '@/types';

/** Active session management for the authenticated user. */
export const sessionService = {
  list: (): Promise<Session[]> =>
    unwrap<Session[]>(api.get<ApiEnvelope<Session[]>>('/me/sessions')),

  revoke: (id: string) => api.delete<ApiEnvelope<unknown>>(`/me/sessions/${id}`),

  /**
   * Revokes every session except the current one. The current refresh token is
   * passed so the backend can keep that session alive.
   */
  revokeOthers: (keepRefreshToken?: string) =>
    api.post<ApiEnvelope<unknown>>('/me/sessions/revoke-others', {
      keep_refresh_token: keepRefreshToken,
    }),
};
