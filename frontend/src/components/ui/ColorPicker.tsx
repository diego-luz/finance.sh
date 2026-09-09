import { Check } from 'lucide-react';
import { cn } from '@/lib/cn';
import { PALETTE } from './palette';

interface ColorPickerProps {
  label?: string;
  value: string;
  onChange: (color: string) => void;
}

export function ColorPicker({ label, value, onChange }: ColorPickerProps) {
  return (
    <div>
      {label && (
        <span className="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {label}
        </span>
      )}
      <div className="flex flex-wrap gap-2">
        {PALETTE.map((color) => (
          <button
            key={color}
            type="button"
            onClick={() => onChange(color)}
            className={cn(
              'flex h-8 w-8 items-center justify-center rounded-full transition hover:scale-110',
              value === color && 'ring-2 ring-offset-2 ring-offset-white dark:ring-offset-ink-surface',
            )}
            style={{ backgroundColor: color, boxShadow: value === color ? `0 0 0 2px ${color}` : undefined }}
            aria-label={`Cor ${color}`}
          >
            {value === color && <Check className="h-4 w-4 text-white" />}
          </button>
        ))}
      </div>
    </div>
  );
}
