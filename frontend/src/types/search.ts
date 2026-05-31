import type { CategoryKind } from './category';

/** Lightweight transaction hit returned by the global search endpoint. */
export interface SearchTransaction {
  id: string;
  description: string;
  /** Amount in cents (int64). */
  amount: number;
  type: 'income' | 'expense' | 'transfer';
  /** ISO date string. */
  date: string;
  category?: { id: string; name: string; color: string };
}

export interface SearchContact {
  id: string;
  name: string;
  type: string;
}

export interface SearchCategory {
  id: string;
  name: string;
  color: string;
  kind: CategoryKind;
  icon: string;
}

export interface SearchAccount {
  id: string;
  name: string;
  type: string;
}

export interface SearchCreditCard {
  id: string;
  name: string;
}

export interface SearchGoal {
  id: string;
  name: string;
}

export interface SearchResults {
  transactions: SearchTransaction[];
  contacts: SearchContact[];
  categories: SearchCategory[];
  accounts: SearchAccount[];
  credit_cards: SearchCreditCard[];
  goals: SearchGoal[];
}

export interface SearchResponse {
  query: string;
  results: SearchResults;
  total: number;
}
