import type { Category } from './category';

export interface Budget {
  id: string;
  category_id: string;
  category: Pick<Category, 'id' | 'name' | 'color' | 'icon' | 'kind'>;
  /** Budgeted amount in cents (int64). */
  amount: number;
  /** Month 1..12. */
  month: number;
  year: number;
  /** Amount already spent in cents (int64). */
  spent: number;
  /** Percentage spent (0..100+, can exceed 100). */
  percent: number;
}

export interface BudgetPayload {
  category_id: string;
  amount: number;
  month: number;
  year: number;
}

export interface BudgetFilters {
  month: number;
  year: number;
}
