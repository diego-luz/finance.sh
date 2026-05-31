import type { Category } from './category';
import type { Tag } from './tag';

export type TransactionType = 'income' | 'expense' | 'transfer';

/** Derived settlement status for payables/receivables. */
export type TransactionStatus = 'open' | 'paid' | 'overdue';

export interface Transaction {
  id: string;
  account_id: string;
  /** Destination account for `transfer` transactions. */
  transfer_account_id?: string;
  type: TransactionType;
  /** Amount in cents (int64), always positive. */
  amount: number;
  description: string;
  /** ISO date string. */
  date: string;
  paid: boolean;
  recurring: boolean;
  notes?: string;
  category?: Category;
  /** ISO date string — when the bill is due. */
  due_date?: string;
  /** Linked contact (customer/supplier). */
  contact_id?: string;
  /** ISO date string — when the bill was settled (output only). */
  paid_at?: string;
  /** Derived settlement status. */
  status?: TransactionStatus;
  /** Expanded contact summary, when present. */
  contact?: { id: string; name: string };
  /** Identifier shared by every transaction of the same installment group. */
  installment_group_id?: string;
  /** 1-based position of this transaction within its installment group. */
  installment_number?: number;
  /** Total number of installments (0 or absent = not parcelado). */
  installment_total?: number;
  /** Number of receipt attachments linked to this transaction. */
  attachment_count?: number;
  /** Tags applied to this transaction (output only). */
  tags?: Tag[];
}

export interface TransactionPayload {
  account_id: string;
  category_id?: string;
  transfer_account_id?: string;
  type: TransactionType;
  amount: number;
  description: string;
  /** ISO date string. */
  date: string;
  paid: boolean;
  /**
   * Legacy per-transaction recurrence flag. No longer sent by the form (the
   * proper recurrence feature lives at /recorrencias); kept optional for
   * backward compatibility with existing callers/DTO.
   */
  recurring?: boolean;
  notes?: string;
  /** ISO date string — when the bill is due. */
  due_date?: string;
  /** Linked contact (customer/supplier). */
  contact_id?: string;
  /**
   * Number of installments (>=1). When >1 on create, the backend splits the
   * amount and creates a group of N monthly transactions.
   */
  installments?: number;
  /** Tags to apply to the transaction. */
  tag_ids?: string[];
}

/** Scope for deleting a transaction that belongs to an installment group. */
export type DeleteScope = 'one' | 'all';

export interface TransactionFilters {
  type?: TransactionType | '';
  account_id?: string;
  category_id?: string;
  tag_id?: string;
  from?: string;
  to?: string;
  search?: string;
  page?: number;
  per_page?: number;
}

/** Body for POST /transactions/bulk-settle. */
export interface BulkSettlePayload {
  ids: string[];
  account_id?: string;
  /** ISO date string. */
  paid_at?: string;
}

/** Body for POST /transactions/bulk-categorize. */
export interface BulkCategorizePayload {
  ids: string[];
  category_id: string;
}

/** Result of a bulk transaction operation. */
export interface BulkResult {
  updated: number;
}
