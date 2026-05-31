import { format, parseISO, isValid } from 'date-fns';
import { ptBR, enUS, es } from 'date-fns/locale';
import type { Locale } from 'date-fns';
import i18n from '@/i18n';

/**
 * Map a UI language tag to its date-fns locale. The DISPLAY locale of dates
 * follows the chosen interface language (independent of the org currency
 * locale, which stays in currency.ts). Defaults to pt-BR.
 */
const dateFnsLocales: Record<string, Locale> = {
  'pt-BR': ptBR,
  en: enUS,
  es,
};

/** Resolve the active date-fns locale from the current i18n language. */
function activeDateLocale(): Locale {
  const lang = (i18n.resolvedLanguage ?? i18n.language ?? 'pt-BR') as string;
  return dateFnsLocales[lang] ?? ptBR;
}

/** Format an ISO date string for display, e.g. "12 mai 2026". */
export function formatDate(iso: string, pattern = "dd MMM yyyy"): string {
  if (!iso) return '';
  const d = parseISO(iso);
  if (!isValid(d)) return '';
  return format(d, pattern, { locale: activeDateLocale() });
}

/** Short numeric date, e.g. "12/05/2026". */
export function formatDateShort(iso: string): string {
  return formatDate(iso, 'dd/MM/yyyy');
}

/**
 * Convert a date input (yyyy-MM-dd) into a UTC-midnight ISO string.
 * Building from the parts (not `new Date(str)`) avoids local-timezone drift,
 * so a date never walks backward a day in negative-offset zones like UTC-3.
 */
export function toISODate(value: string): string {
  const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(value ?? '');
  if (!m) return '';
  const [, y, mo, d] = m;
  return new Date(Date.UTC(Number(y), Number(mo) - 1, Number(d))).toISOString();
}

/**
 * Get yyyy-MM-dd string for an input[type=date] from an ISO string.
 * Derives the date from the ISO string's UTC components so a value stored as
 * `...T00:00:00Z` renders as the same calendar day regardless of local offset.
 */
export function toInputDate(iso?: string): string {
  if (!iso) return format(new Date(), 'yyyy-MM-dd');
  // Fast path: ISO strings start with yyyy-MM-dd; use it verbatim (UTC date part).
  const m = /^(\d{4}-\d{2}-\d{2})/.exec(iso);
  if (m) return m[1];
  const d = parseISO(iso);
  if (!isValid(d)) return format(new Date(), 'yyyy-MM-dd');
  return format(d, 'yyyy-MM-dd');
}

/** Today as yyyy-MM-dd. */
export function todayInput(): string {
  return format(new Date(), 'yyyy-MM-dd');
}

/** Identifiers for the quick date-range presets. */
export type PeriodPreset =
  | 'this_month'
  | 'last_month'
  | 'last_30'
  | 'last_90'
  | 'this_year'
  | 'custom';

/** Format a Date as yyyy-MM-dd using its local calendar parts. */
function ymd(d: Date): string {
  return format(d, 'yyyy-MM-dd');
}

/**
 * Resolve a preset into a {from,to} pair of yyyy-MM-dd strings.
 * Returns empty strings for 'custom' (the caller keeps its manual values).
 */
export function resolvePeriodPreset(preset: PeriodPreset): { from: string; to: string } {
  const now = new Date();
  const y = now.getFullYear();
  const m = now.getMonth();
  const d = now.getDate();

  switch (preset) {
    case 'this_month':
      return { from: ymd(new Date(y, m, 1)), to: ymd(new Date(y, m + 1, 0)) };
    case 'last_month':
      return { from: ymd(new Date(y, m - 1, 1)), to: ymd(new Date(y, m, 0)) };
    case 'last_30':
      return { from: ymd(new Date(y, m, d - 29)), to: ymd(now) };
    case 'last_90':
      return { from: ymd(new Date(y, m, d - 89)), to: ymd(now) };
    case 'this_year':
      return { from: ymd(new Date(y, 0, 1)), to: ymd(new Date(y, 11, 31)) };
    case 'custom':
    default:
      return { from: '', to: '' };
  }
}
