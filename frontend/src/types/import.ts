import type { TransactionType } from './transaction';

/** Statement parsing format requested for the preview. */
export type ImportFormat = 'auto' | 'ofx' | 'csv';

/** CSV delimiter override; empty string means auto-detect. */
export type CsvDelimiter = '' | ',' | ';';

/** Decimal separator used by the CSV's amount column. */
export type DecimalSep = ',' | '.';

/**
 * Optional CSV parsing knobs sent alongside the preview request. Every value
 * has an "auto" default (-1 columns, '' delimiter) so the simple case needs no
 * configuration at all.
 */
export interface CsvOptions {
  /** Field delimiter. '' = auto-detect. */
  delimiter?: CsvDelimiter;
  /** Whether the first row is a header. */
  has_header?: boolean;
  /** 0-based date column index, or -1 to auto-detect. */
  date_col?: number;
  /** 0-based description column index, or -1 to auto-detect. */
  desc_col?: number;
  /** 0-based amount column index, or -1 to auto-detect. */
  amount_col?: number;
  /** Go-style date layout, e.g. '02/01/2006'. */
  date_format?: string;
  /** Decimal separator of the amount column. */
  decimal_sep?: DecimalSep;
}

/** A single parsed statement line returned by the preview. */
export interface ImportRow {
  /** Position within the parsed file (stable key). */
  index: number;
  /** ISO date string. */
  date: string;
  description: string;
  /** Amount in cents (int64), always positive. */
  amount: number;
  type: TransactionType;
  /** Bank-provided unique id (FITID for OFX), used for dedupe. */
  external_id: string;
  /** True when a matching transaction already exists. */
  duplicate: boolean;
  /** Human-readable explanation for a duplicate / skipped row. */
  reason?: string;
}

/** Aggregate counts for the parsed preview. */
export interface ImportSummary {
  total: number;
  new: number;
  duplicates: number;
}

/** Response of POST /imports/preview. */
export interface ImportPreview {
  format: Exclude<ImportFormat, 'auto'>;
  rows: ImportRow[];
  summary: ImportSummary;
}

/** A row selected by the user to be committed. */
export interface ImportCommitRow {
  /** ISO date string. */
  date: string;
  description: string;
  /** Amount in cents (int64), always positive. */
  amount: number;
  type: TransactionType;
  external_id: string;
}

/** Body for POST /imports/commit. */
export interface ImportCommitPayload {
  account_id: string;
  category_id?: string;
  rows: ImportCommitRow[];
}

/** Response of POST /imports/commit. */
export interface ImportCommitResult {
  created: number;
  skipped: number;
}
