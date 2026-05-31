import { api, unwrap } from '@/lib/axios';
import type {
  ApiEnvelope,
  Currency,
  Organization,
  UpdateOrganizationPayload,
} from '@/types';

export const organizationService = {
  /** Supported currencies (requires auth). */
  getCurrencies: () =>
    unwrap<Currency[]>(api.get<ApiEnvelope<Currency[]>>('/currencies')),

  /** The current organization (resolved from the X-Organization-ID header). */
  getOrganization: () =>
    unwrap<Organization>(api.get<ApiEnvelope<Organization>>('/organization')),

  /** Update the current organization (owner/admin only). */
  updateOrganization: (payload: UpdateOrganizationPayload) =>
    unwrap<Organization>(api.put<ApiEnvelope<Organization>>('/organization', payload)),

  /** Create an ADDITIONAL organization owned by the caller (e.g. Casa + Microempresa). */
  createOrganization: (payload: { name: string; currency?: string }) =>
    unwrap<Organization>(api.post<ApiEnvelope<Organization>>('/organizations', payload)),
};
