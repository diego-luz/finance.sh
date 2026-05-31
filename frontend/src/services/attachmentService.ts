import { api, unwrap } from '@/lib/axios';
import { downloadBlob } from './reportService';
import type { ApiEnvelope, Attachment } from '@/types';

/**
 * Receipt attachment endpoints.
 *
 * Upload sends `multipart/form-data` with the field name `file`. We do NOT
 * hardcode the Content-Type header value — the axios instance defaults to
 * `application/json`, so we override it to `undefined` for this call, which
 * makes axios detect the FormData body and emit `multipart/form-data` with the
 * correct boundary automatically.
 *
 * Download returns a raw binary stream (not the JSON envelope), so it requests
 * `responseType: 'blob'` and triggers a browser download via {@link downloadBlob}.
 */
export const attachmentService = {
  /** List all attachments for a transaction. */
  list: (transactionId: string): Promise<Attachment[]> =>
    unwrap<Attachment[]>(
      api.get<ApiEnvelope<Attachment[]>>(`/transactions/${transactionId}/attachments`),
    ),

  /** Upload a single file as an attachment of the given transaction. */
  upload: (transactionId: string, file: File): Promise<Attachment> => {
    const formData = new FormData();
    formData.append('file', file);
    return unwrap<Attachment>(
      api.post<ApiEnvelope<Attachment>>(
        `/transactions/${transactionId}/attachments`,
        formData,
        // Override the instance JSON default so axios sets the multipart
        // Content-Type (with boundary) for the FormData body itself.
        { headers: { 'Content-Type': undefined } },
      ),
    );
  },

  /** Download an attachment as a Blob and trigger a browser save. */
  download: async (attachment: Attachment): Promise<void> => {
    const res = await api.get(`/attachments/${attachment.id}/download`, {
      responseType: 'blob',
    });
    downloadBlob(res.data as Blob, attachment.file_name);
  },

  /** Permanently delete an attachment. */
  remove: (id: string) => api.delete(`/attachments/${id}`),
};
