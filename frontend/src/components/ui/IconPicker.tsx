import { iconMap, iconOptions } from '@/lib/icons';
import { cn } from '@/lib/cn';

interface IconPickerProps {
  label?: string;
  value: string;
  onChange: (icon: string) => void;
  color?: string;
}

export function IconPicker({ label, value, onChange, color = '#10b981' }: IconPickerProps) {
  return (
    <div>
      {label && (
        <span className="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {label}
        </span>
      )}
      <div className="grid grid-cols-8 gap-2">
        {iconOptions.map((name) => {
          const Icon = iconMap[name];
          const active = value === name;
          return (
            <button
              key={name}
              type="button"
              onClick={() => onChange(name)}
              className={cn(
                'flex h-9 w-9 items-center justify-center rounded-lg border transition',
                active
                  ? 'border-primary bg-primary/10'
                  : 'border-gray-200 hover:border-gray-300 dark:border-ink-border dark:hover:border-gray-600',
              )}
              style={active ? { color } : undefined}
              aria-label={name}
            >
              <Icon className="h-4 w-4" />
            </button>
          );
        })}
      </div>
    </div>
  );
}
