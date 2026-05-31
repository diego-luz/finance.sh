import { formatDateShort } from '@/lib/date';
import type { RecurrenceFrequency } from '@/types';

/** Plural noun for each frequency unit (pt-BR). */
const unitPlural: Record<RecurrenceFrequency, string> = {
  daily: 'dias',
  weekly: 'semanas',
  monthly: 'meses',
  yearly: 'anos',
};

/** Simple label for the cadence when interval is 1, e.g. "Mensal". */
const simpleLabel: Record<RecurrenceFrequency, string> = {
  daily: 'Diária',
  weekly: 'Semanal',
  monthly: 'Mensal',
  yearly: 'Anual',
};

/**
 * Humanize a frequency + interval, e.g.:
 * - interval 1 → "Mensal"
 * - interval 2 → "A cada 2 semanas"
 */
export function humanizeFrequency(
  frequency: RecurrenceFrequency,
  interval: number,
): string {
  const n = Math.max(1, Math.floor(interval || 1));
  if (n === 1) return simpleLabel[frequency];
  return `A cada ${n} ${unitPlural[frequency]}`;
}

/**
 * A short human summary used as helper text in the form, e.g.:
 * "Mensalmente a partir de 10/06/2026" or
 * "A cada 2 semanas a partir de 10/06/2026".
 */
export function summarizeSchedule(
  frequency: RecurrenceFrequency,
  interval: number,
  startDate: string,
): string {
  const n = Math.max(1, Math.floor(interval || 1));
  const when = startDate ? ` a partir de ${formatDateShort(startDate)}` : '';
  if (n === 1) {
    const adverb: Record<RecurrenceFrequency, string> = {
      daily: 'Diariamente',
      weekly: 'Semanalmente',
      monthly: 'Mensalmente',
      yearly: 'Anualmente',
    };
    return `${adverb[frequency]}${when}.`;
  }
  return `A cada ${n} ${unitPlural[frequency]}${when}.`;
}
