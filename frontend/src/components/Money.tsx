import { formatCurrency } from '@/lib/currency';
import { cn } from '@/lib/cn';

interface MoneyProps {
  /** Amount in integer cents (backend convention). */
  value: number;
  /** Optional currency code; defaults to the active org currency. */
  currency?: string;
  className?: string;
}

/**
 * Renders a monetary value with the `money` class so it can be blurred globally
 * when privacy mode is active (see {@link usePrivacyStore} and the
 * `.values-hidden .money` rule in index.css). Number/currency logic is
 * untouched — only visibility changes when hidden.
 */
export function Money({ value, currency, className }: MoneyProps) {
  return (
    <span className={cn('money tabular-nums', className)}>
      {formatCurrency(value, currency)}
    </span>
  );
}
