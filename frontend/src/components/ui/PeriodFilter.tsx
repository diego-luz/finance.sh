import { useState } from 'react';
import { CalendarRange } from 'lucide-react';
import { Input } from './Input';
import { cn } from '@/lib/cn';
import { resolvePeriodPreset, type PeriodPreset } from '@/lib/date';

export interface PeriodValue {
  from?: string;
  to?: string;
}

interface PeriodFilterProps {
  value: PeriodValue;
  onChange: (value: PeriodValue) => void;
  /** Hide the leading calendar icon + label (compact embedded usage). */
  compact?: boolean;
  className?: string;
}

const PRESETS: { id: PeriodPreset; label: string }[] = [
  { id: 'this_month', label: 'Este mês' },
  { id: 'last_month', label: 'Mês passado' },
  { id: 'last_30', label: 'Últimos 30 dias' },
  { id: 'last_90', label: 'Últimos 90 dias' },
  { id: 'this_year', label: 'Este ano' },
  { id: 'custom', label: 'Personalizado' },
];

/**
 * Quick date-range presets plus a manual ("Personalizado") path. Returns the
 * resolved {from,to} as yyyy-MM-dd strings (or undefined when cleared).
 */
export function PeriodFilter({ value, onChange, compact, className }: PeriodFilterProps) {
  const [preset, setPreset] = useState<PeriodPreset>('custom');

  const selectPreset = (id: PeriodPreset) => {
    setPreset(id);
    if (id === 'custom') return;
    const { from, to } = resolvePeriodPreset(id);
    onChange({ from: from || undefined, to: to || undefined });
  };

  const updateCustom = (patch: PeriodValue) => {
    setPreset('custom');
    onChange({ ...value, ...patch });
  };

  return (
    <div className={cn('flex flex-col gap-3', className)}>
      <div className="flex flex-wrap items-center gap-1.5">
        {!compact && (
          <span className="mr-1 inline-flex items-center gap-1.5 text-xs font-medium text-gray-500 dark:text-gray-400">
            <CalendarRange className="h-3.5 w-3.5" /> Período
          </span>
        )}
        {PRESETS.map((p) => (
          <button
            key={p.id}
            type="button"
            onClick={() => selectPreset(p.id)}
            aria-pressed={preset === p.id}
            className={cn(
              'rounded-full border px-3 py-1 text-xs font-medium transition',
              preset === p.id
                ? 'border-primary bg-primary/10 text-primary-700 dark:text-primary-300'
                : 'border-gray-200 text-gray-600 hover:border-gray-300 hover:bg-gray-50 dark:border-ink-border dark:text-gray-400 dark:hover:bg-ink-elevated',
            )}
          >
            {p.label}
          </button>
        ))}
      </div>

      {preset === 'custom' && (
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <Input
            label="De"
            type="date"
            value={value.from ?? ''}
            onChange={(e) => updateCustom({ from: e.target.value || undefined })}
          />
          <Input
            label="Até"
            type="date"
            value={value.to ?? ''}
            onChange={(e) => updateCustom({ to: e.target.value || undefined })}
          />
        </div>
      )}
    </div>
  );
}
