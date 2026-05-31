import { api } from '@/lib/axios';
import type { ApiEnvelope } from '@/types';

/** Summary returned by POST /me/import. */
export interface ImportSummary {
  organization_id: string;
  organization_name: string;
  accounts: number;
  categories: number;
  contacts: number;
  credit_cards: number;
  budgets: number;
  goals: number;
  transactions: number;
  skipped: number;
}

/** LGPD-oriented self-service: data portability and account deletion. */
export const accountPrivacyService = {
  /**
   * Downloads a full export of the authenticated user's data as a JSON Blob.
   * The endpoint returns a raw file (not the JSON envelope), so we request
   * `responseType: 'blob'`.
   */
  exportData: async (): Promise<Blob> => {
    const res = await api.get('/me/export', { responseType: 'blob' });
    return res.data as Blob;
  },

  /**
   * Imports a previously-exported JSON into a BRAND-NEW organization owned by
   * the user. Sends the raw file text; returns a summary of restored rows.
   */
  importData: async (file: File): Promise<ImportSummary> => {
    const text = await file.text();
    const res = await api.post<ApiEnvelope<ImportSummary>>('/me/import', text, {
      headers: { 'Content-Type': 'application/json' },
    });
    return res.data.data as ImportSummary;
  },

  /** Deletes / anonymizes the authenticated account after a password check. */
  deleteAccount: (password: string) =>
    api.delete<ApiEnvelope<unknown>>('/me/account', { data: { password } }),
};
