import { api, unwrap } from '@/lib/axios';
import type { ApiEnvelope, AuthResponse } from '@/types';

/**
 * Payload for first-run setup. The backend creates the bootstrap admin user
 * and the inaugural organization atomically, then returns a full session
 * envelope identical to the login response.
 */
export interface SetupInitializePayload {
  user: {
    name: string;
    email: string;
    password: string;
  };
  organization: {
    name: string;
    currency: string;
  };
}

export interface SetupStatus {
  needs_setup: boolean;
}

/**
 * Public probe used by the routing gate to decide whether to force the wizard.
 * Returns `{ needs_setup: false }` for any non-OK envelope so the UI stays
 * fail-open on transient backend errors.
 */
export function getSetupStatus(): Promise<SetupStatus> {
  return unwrap<SetupStatus>(
    api.get<ApiEnvelope<SetupStatus>>('/setup/status'),
  );
}

/**
 * Bootstrap the platform. On 201 returns a full session (same shape as login).
 * On 409 the backend reports "already initialized" — surfaced as an
 * ApiRequestError with status 409 so callers can redirect to /login.
 */
export function initializeSetup(
  payload: SetupInitializePayload,
): Promise<AuthResponse> {
  return unwrap<AuthResponse>(
    api.post<ApiEnvelope<AuthResponse>>('/setup/initialize', payload),
  );
}
