import type { Organization } from './auth';

/** A currency supported by the backend, returned by GET /currencies. */
export interface Currency {
  code: string;
  name: string;
  symbol: string;
}

/** Payload for PUT /organization (owner/admin only). */
export interface UpdateOrganizationPayload {
  name?: string;
  currency?: string;
}

/** Re-export the current-org shape for convenience. */
export type { Organization };
