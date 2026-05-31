import type { Transaction } from './transaction';

/** Lifecycle status of a credit-card invoice. */
export type InvoiceStatus = 'open' | 'closed' | 'paid' | 'overdue';

/** A credit-card invoice (fatura) summary. */
export interface Invoice {
  /** Billing period reference, e.g. "2026-05". */
  reference: string;
  /** ISO date — first day covered by the invoice. */
  period_start: string;
  /** ISO date — last day covered by the invoice. */
  period_end: string;
  /** ISO date — when the invoice is due. */
  due_date: string;
  /** Total amount in cents (int64). */
  total: number;
  /** Already paid amount in cents (int64). */
  paid_total: number;
  /** Remaining open amount in cents (int64). */
  open_total: number;
  status: InvoiceStatus;
  /** Number of transactions composing the invoice. */
  transaction_count: number;
}

/** Detailed invoice including its composing transactions. */
export interface InvoiceDetail {
  invoice: Invoice;
  transactions: Transaction[];
}

/** Body for POST /credit-cards/:id/invoices/:reference/pay. */
export interface InvoicePayPayload {
  account_id: string;
  /** ISO date string — when the invoice was paid. */
  paid_at?: string;
}
