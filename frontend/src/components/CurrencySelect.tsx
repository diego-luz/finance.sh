import { forwardRef, useMemo, type SelectHTMLAttributes } from 'react';
import { Select } from '@/components/ui';
import { SUPPORTED_CURRENCIES } from '@/lib/currency';
import type { Currency } from '@/types';

interface CurrencySelectProps
  extends Omit<SelectHTMLAttributes<HTMLSelectElement>, 'children'> {
  label?: string;
  error?: string;
  /**
   * Currency list to offer. Defaults to the static SUPPORTED_CURRENCIES so the
   * component works while unauthenticated (registration). Pass the result of
   * `useCurrencies()` once authenticated to mirror the backend exactly.
   */
  currencies?: Currency[];
}

/** Format an option label like "BRL — Real brasileiro (R$)". */
function currencyLabel(c: Currency): string {
  return `${c.code} — ${c.name} (${c.symbol})`;
}

/**
 * A <Select> of supported currencies. Compatible with react-hook-form
 * (forwards a ref) and with controlled usage (value/onChange).
 */
export const CurrencySelect = forwardRef<HTMLSelectElement, CurrencySelectProps>(
  ({ label = 'Moeda', error, currencies, ...props }, ref) => {
    const options = useMemo(() => {
      const list = currencies && currencies.length > 0 ? currencies : SUPPORTED_CURRENCIES;
      return list.map((c) => ({ value: c.code, label: currencyLabel(c) }));
    }, [currencies]);

    return <Select ref={ref} label={label} error={error} options={options} {...props} />;
  },
);
CurrencySelect.displayName = 'CurrencySelect';
