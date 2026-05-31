/**
 * Money is stored as int64 CENTS everywhere on the backend.
 * These helpers convert to/from the human-facing decimal representation.
 *
 * Formatting is CURRENCY-AWARE: the organization picks a single currency and
 * every amount renders with that currency's symbol + the matching locale.
 * The "active" currency is a module-level value kept in sync with the current
 * organization (see {@link setActiveCurrency}). Existing `formatCurrency(cents)`
 * call sites therefore automatically reflect the org currency without changes.
 */

/** A currency supported by the backend, as returned by GET /currencies. */
export interface SupportedCurrency {
  code: string;
  name: string;
  symbol: string;
}

/**
 * Static list of supported currencies, mirroring the backend set. Used at
 * registration (unauthenticated, so GET /currencies is unavailable) and as a
 * fallback elsewhere. Keep in sync with the backend's supported list.
 */
export const SUPPORTED_CURRENCIES: SupportedCurrency[] = [
  { code: 'BRL', name: 'Real brasileiro', symbol: 'R$' },
  { code: 'USD', name: 'Dólar americano', symbol: 'US$' },
  { code: 'EUR', name: 'Euro', symbol: '€' },
  { code: 'GBP', name: 'Libra esterlina', symbol: '£' },
  { code: 'ARS', name: 'Peso argentino', symbol: '$' },
  { code: 'CLP', name: 'Peso chileno', symbol: '$' },
  { code: 'MXN', name: 'Peso mexicano', symbol: '$' },
  { code: 'JPY', name: 'Iene japonês', symbol: '¥' },
];

/**
 * Map a currency code to the locale used to format it. Falls back to pt-BR
 * (the project default) for any currency not listed here.
 */
const currencyLocale: Record<string, string> = {
  BRL: 'pt-BR',
  USD: 'en-US',
  EUR: 'de-DE',
  GBP: 'en-GB',
  JPY: 'ja-JP',
  ARS: 'es-AR',
  CLP: 'es-CL',
  MXN: 'es-MX',
};

const DEFAULT_CURRENCY = 'BRL';
const DEFAULT_LOCALE = 'pt-BR';

/** Module-level active currency, kept in sync with the current organization. */
let activeCurrency = DEFAULT_CURRENCY;

/** Set the active currency (call when the org loads or switches). */
export function setActiveCurrency(code: string | undefined | null): void {
  activeCurrency = (code || DEFAULT_CURRENCY).toUpperCase();
}

/** Get the active currency code (e.g. for display or as a Select default). */
export function getActiveCurrency(): string {
  return activeCurrency;
}

/** Resolve the locale to use for a given currency code. */
export function localeForCurrency(currency: string): string {
  return currencyLocale[currency?.toUpperCase()] ?? DEFAULT_LOCALE;
}

/**
 * Number of fraction digits a currency uses (0 for zero-decimal currencies
 * such as JPY/CLP). Derived from Intl so the set stays in sync with the
 * runtime's CLDR data, with a safe fallback to 2.
 */
function fractionDigitsFor(currency: string): number {
  try {
    const opts = new Intl.NumberFormat(localeForCurrency(currency), {
      style: 'currency',
      currency,
    }).resolvedOptions();
    return opts.maximumFractionDigits ?? 2;
  } catch {
    return 2;
  }
}

/**
 * Format an integer cents amount as currency. When `currency` is omitted the
 * active organization currency is used, e.g. 123456 -> "R$ 1.234,56" (BRL) or
 * "$1,234.56" (USD). Zero-decimal currencies (JPY) are handled correctly.
 */
export function formatCurrency(cents: number, currency: string = activeCurrency): string {
  const code = (currency || DEFAULT_CURRENCY).toUpperCase();
  // The backend ALWAYS stores amounts as int64 hundredths (×100), regardless of
  // currency. So scaling is a fixed /100. The DISPLAY decimals are left to Intl
  // per currency (e.g. JPY shows 0 decimals, BRL/USD show 2).
  const value = (cents ?? 0) / 100;
  try {
    return new Intl.NumberFormat(localeForCurrency(code), {
      style: 'currency',
      currency: code,
    }).format(value);
  } catch {
    // Invalid/unknown currency code: fall back to a plain number + code.
    return `${value.toFixed(2)} ${code}`;
  }
}

/**
 * Format a compact value WITHOUT the currency symbol, e.g. "1.234,56".
 * Uses the active currency's locale and fraction digits.
 */
export function formatAmount(cents: number, currency: string = activeCurrency): string {
  const code = (currency || DEFAULT_CURRENCY).toUpperCase();
  const digits = fractionDigitsFor(code);
  // Scale is fixed /100 (backend convention); display decimals vary per currency.
  return new Intl.NumberFormat(localeForCurrency(code), {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  }).format((cents ?? 0) / 100);
}

/**
 * Parse a user-typed currency string into integer cents (minor units).
 * Tolerant of both pt-BR ("1.234,56") and en-US ("1,234.56") notations, as
 * well as bare numbers and a leading symbol. Honors the active currency's
 * fraction digits, so JPY "1234" -> 1234 (no implicit centavos).
 */
export function parseCurrencyToCents(
  input: string,
  _currency: string = activeCurrency,
): number {
  if (!input) return 0;

  // Strip everything except digits, comma, dot and minus.
  let cleaned = input.replace(/[^0-9,.-]/g, '').trim();
  if (!cleaned) return 0;

  const hasComma = cleaned.includes(',');
  const hasDot = cleaned.includes('.');

  if (hasComma && hasDot) {
    // Decimal separator is whichever appears LAST; the other is a thousands
    // grouping. Covers both "1.234,56" (pt-BR) and "1,234.56" (en-US).
    if (cleaned.lastIndexOf(',') > cleaned.lastIndexOf('.')) {
      cleaned = cleaned.replace(/\./g, '').replace(',', '.');
    } else {
      cleaned = cleaned.replace(/,/g, '');
    }
  } else if (hasComma) {
    cleaned = cleaned.replace(',', '.');
  } else if (hasDot) {
    // Only dot(s): a single group of exactly 3 trailing digits (e.g. "1.234",
    // "1.234.567") is thousands grouping -> drop. Otherwise it's a decimal.
    if (/^-?\d{1,3}(\.\d{3})+$/.test(cleaned)) {
      cleaned = cleaned.replace(/\./g, '');
    }
  }

  const value = parseFloat(cleaned);
  if (Number.isNaN(value)) return 0;
  // Backend stores hundredths (×100) for every currency.
  return Math.round(value * 100);
}

/** Convert integer cents to a plain decimal string suitable for an input value. */
export function centsToInput(cents: number, currency: string = activeCurrency): string {
  if (cents === 0 || cents == null) return '';
  const digits = fractionDigitsFor((currency || DEFAULT_CURRENCY).toUpperCase());
  // Scale is fixed /100; the input string shows the currency's display decimals.
  const str = (cents / 100).toFixed(digits);
  return digits > 0 ? str.replace('.', ',') : str;
}
