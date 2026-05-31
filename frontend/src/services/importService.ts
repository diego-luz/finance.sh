import { api, unwrap } from '@/lib/axios';
import type {
  ApiEnvelope,
  CsvOptions,
  ImportCommitPayload,
  ImportCommitResult,
  ImportFormat,
  ImportPreview,
} from '@/types';

/**
 * Bank statement import endpoints (OFX / CSV).
 *
 * `preview` sends `multipart/form-data` with the field name `file` plus a set
 * of scalar form fields. As with {@link attachmentService}, we do NOT hardcode
 * the multipart Content-Type: the axios instance defaults to
 * `application/json`, so we override the header to `undefined`, which makes
 * axios detect the FormData body and emit `multipart/form-data` with the
 * correct boundary automatically.
 *
 * `commit` is a plain JSON POST.
 */
export const importService = {
  /** Parse a statement file and return the deduplicated row preview. */
  preview: (
    file: File,
    accountId: string,
    format: ImportFormat = 'auto',
    csv?: CsvOptions,
  ): Promise<ImportPreview> => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('account_id', accountId);
    formData.append('format', format);

    if (csv) {
      // Only meaningful for CSV; harmless for OFX (the backend ignores them).
      if (csv.delimiter !== undefined) formData.append('delimiter', csv.delimiter);
      if (csv.has_header !== undefined) {
        formData.append('has_header', csv.has_header ? 'true' : 'false');
      }
      if (csv.date_col !== undefined) formData.append('date_col', String(csv.date_col));
      if (csv.desc_col !== undefined) formData.append('desc_col', String(csv.desc_col));
      if (csv.amount_col !== undefined) {
        formData.append('amount_col', String(csv.amount_col));
      }
      if (csv.date_format !== undefined) formData.append('date_format', csv.date_format);
      if (csv.decimal_sep !== undefined) formData.append('decimal_sep', csv.decimal_sep);
    }

    return unwrap<ImportPreview>(
      api.post<ApiEnvelope<ImportPreview>>('/imports/preview', formData, {
        // Override the instance JSON default so axios sets the multipart
        // Content-Type (with boundary) for the FormData body itself.
        headers: { 'Content-Type': undefined },
      }),
    );
  },

  /** Commit the selected rows as transactions. */
  commit: (payload: ImportCommitPayload): Promise<ImportCommitResult> =>
    unwrap<ImportCommitResult>(
      api.post<ApiEnvelope<ImportCommitResult>>('/imports/commit', payload),
    ),
};
