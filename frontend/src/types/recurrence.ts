/** A recurrence only models cash in/out (never transfers). */
export type RecurrenceType = 'income' | 'expense';

/** How often a recurrence rule fires. */
export type RecurrenceFrequency = 'daily' | 'weekly' | 'monthly' | 'yearly';

/** Lightweight category snapshot embedded in a recurrence rule. */
interface RecurrenceCategoryRef {
  id: string;
  name: string;
  color: string;
  icon: string;
  kind: 'income' | 'expense';
}

/** Lightweight account snapshot embedded in a recurrence rule. */
interface RecurrenceAccountRef {
  id: string;
  name: string;
}

/**
 * A scheduled transaction template that generates real transactions
 * automatically on its cadence. Money is stored in CENTS, dates as ISO.
 */
export interface RecurrenceRule {
  id: string;
  type: RecurrenceType;
  /** Amount in cents (int64). */
  amount: number;
  description: string;
  account_id: string;
  category_id?: string;
  contact_id?: string;
  /** Whether generated transactions are marked paid/received. */
  paid: boolean;
  frequency: RecurrenceFrequency;
  /** Fire every N units of {@link frequency} (e.g. 2 = every 2 weeks). */
  interval: number;
  /** ISO date the schedule starts from. */
  start_date: string;
  /** Optional ISO date the schedule stops at. */
  end_date?: string;
  /** Cap on total occurrences (0 = unlimited). */
  max_occurrences: number;
  /** How many occurrences have already been generated. */
  occurrences_count: number;
  /** ISO date of the next scheduled occurrence. */
  next_run_date: string;
  active: boolean;
  account?: RecurrenceAccountRef;
  category?: RecurrenceCategoryRef;
}

/** Payload for creating or updating a recurrence rule. */
export interface RecurrenceRulePayload {
  type: RecurrenceType;
  /** Amount in cents (int64). */
  amount: number;
  description: string;
  account_id: string;
  category_id?: string;
  contact_id?: string;
  paid?: boolean;
  frequency: RecurrenceFrequency;
  interval?: number;
  start_date: string;
  end_date?: string;
  max_occurrences?: number;
  active?: boolean;
}
