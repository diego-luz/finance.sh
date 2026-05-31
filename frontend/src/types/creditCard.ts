export interface CreditCard {
  id: string;
  name: string;
  /** Credit limit in cents (int64). */
  limit: number;
  /** Day of month the invoice closes (1..31). */
  closing_day: number;
  /** Day of month the invoice is due (1..31). */
  due_day: number;
  color: string;
  /** Amount already used in cents (int64). */
  used: number;
  /** Remaining available limit in cents (int64). */
  available: number;
}

export interface CreditCardPayload {
  name: string;
  limit: number;
  closing_day: number;
  due_day: number;
  color: string;
}
