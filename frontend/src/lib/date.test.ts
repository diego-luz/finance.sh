import { beforeAll, expect, it } from 'vitest';
import { formatDateShort, toISODate } from './date';

beforeAll(() => {
  process.env.TZ = 'America/Sao_Paulo'; // UTC-3, where the bug showed
});

it('a calendar date at UTC midnight keeps its day in UTC-3', () => {
  expect(formatDateShort('2026-05-12T00:00:00Z')).toBe('12/05/2026');
  expect(formatDateShort('2026-05-12T00:00:00.000Z')).toBe('12/05/2026');
  expect(formatDateShort('2026-05-12')).toBe('12/05/2026');
  expect(formatDateShort(toISODate('2026-01-01'))).toBe('01/01/2026');
});

it('a real timestamp is still shown in local time', () => {
  // 02:30 UTC on the 12th is 23:30 on the 11th in São Paulo
  expect(formatDateShort('2026-05-12T02:30:00Z')).toBe('11/05/2026');
});
