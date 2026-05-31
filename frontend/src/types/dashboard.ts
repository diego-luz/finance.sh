import type { Account } from './account';
import type { Transaction } from './transaction';

export interface CashFlowPoint {
  month: string;
  income: number;
  expense: number;
}

export interface TopCategory {
  category_id: string;
  name: string;
  color: string;
  total: number;
}

export interface Dashboard {
  /** All values in cents (int64). */
  balance: number;
  month_income: number;
  month_expense: number;
  month_net: number;
  upcoming_bills: Transaction[];
  cash_flow: CashFlowPoint[];
  top_categories: TopCategory[];
  accounts_summary: Account[];
}
