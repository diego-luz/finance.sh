import type { CategoryKind } from './category';

/** A single category's aggregated total within the reports summary. */
export interface ReportCategoryTotal {
  category_id: string;
  name: string;
  color: string;
  kind: CategoryKind;
  /** Total in cents (int64). */
  total: number;
}

/** A single month's income/expense within the reports summary. */
export interface ReportMonthTotal {
  /** Label, e.g. "2026-05" or "mai/26". */
  month: string;
  /** Values in cents (int64). */
  income: number;
  expense: number;
}

/** Response of GET /reports/summary. All money values in cents (int64). */
export interface ReportSummary {
  income: number;
  expense: number;
  net: number;
  by_category: ReportCategoryTotal[];
  by_month: ReportMonthTotal[];
}

/** Date range produced by the period presets. ISO date strings (or empty). */
export interface DateRange {
  from?: string;
  to?: string;
}
