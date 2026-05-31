import type { Transaction } from './transaction';

/** Status filter for the payables/receivables list endpoints. */
export type BillStatusFilter = 'open' | 'overdue' | 'paid' | 'all';

export interface BillFilters {
  status?: BillStatusFilter;
  from?: string;
  to?: string;
  contact_id?: string;
  page?: number;
  per_page?: number;
}

/** Aggregated totals (cents) and counts for a payables/receivables list. */
export interface BillSummary {
  total_open: number;
  total_overdue: number;
  due_next_7d: number;
  count_open: number;
}

/** Response of GET /payables and GET /receivables. */
export interface BillListResponse {
  data: Transaction[];
  meta: {
    page: number;
    per_page: number;
    total: number;
    pages: number;
  };
  summary: BillSummary;
}

/** Body for POST /transactions/:id/settle. */
export interface SettlePayload {
  account_id?: string;
  paid_at?: string;
}
