/**
 * Format a byte count into a human-readable string (B / KB / MB).
 * Uses 1024 as the base, pt-BR decimal separator, e.g. 1536 -> "1,5 KB".
 */
export function formatBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return '0 B';
  const units = ['B', 'KB', 'MB'];
  const k = 1024;
  // Clamp the unit index so we never exceed MB.
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(k)), units.length - 1);
  const value = bytes / Math.pow(k, i);
  const decimals = i === 0 ? 0 : 1;
  const formatted = new Intl.NumberFormat('pt-BR', {
    minimumFractionDigits: 0,
    maximumFractionDigits: decimals,
  }).format(value);
  return `${formatted} ${units[i]}`;
}
