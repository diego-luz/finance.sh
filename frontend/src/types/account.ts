export type AccountType = 'bank' | 'wallet' | 'investment' | 'credit_card';

export interface Account {
  id: string;
  name: string;
  type: AccountType;
  /** Initial balance in cents (int64). */
  initial_balance: number;
  /** Current computed balance in cents (int64). */
  balance: number;
  color: string;
  icon: string;
  archived: boolean;
}

export interface AccountPayload {
  name: string;
  type: AccountType;
  initial_balance: number;
  color: string;
  icon: string;
}
