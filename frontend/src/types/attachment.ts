/**
 * A receipt / supporting document attached to a transaction.
 * Files are uploaded via multipart and downloaded as a binary stream.
 */
export interface Attachment {
  id: string;
  transaction_id: string;
  file_name: string;
  /** MIME type, e.g. "image/png" or "application/pdf". */
  content_type: string;
  /** File size in bytes. */
  size: number;
  /** ISO date string. */
  created_at: string;
}

/** Allowed MIME types for receipt attachments. */
export const ATTACHMENT_ACCEPT = 'image/jpeg,image/png,image/webp,image/gif,application/pdf';

/** Allowed MIME types as a set for validation. */
export const ATTACHMENT_ALLOWED_TYPES = [
  'image/jpeg',
  'image/png',
  'image/webp',
  'image/gif',
  'application/pdf',
] as const;

/** Maximum allowed attachment size in bytes (10 MB). */
export const ATTACHMENT_MAX_SIZE = 10 * 1024 * 1024;
