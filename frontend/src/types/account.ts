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
  /** Archived accounts stay in the ledger but drop out of day-to-day pickers.
   *  This is how an account that still has entries is retired, since deleting
   *  one is refused by the API. */
  archived: boolean;
}
