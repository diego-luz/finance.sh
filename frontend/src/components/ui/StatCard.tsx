import type { ReactNode } from 'react';
import type { LucideIcon } from 'lucide-react';
import { ArrowDownRight, ArrowUpRight } from 'lucide-react';
import { cn } from '@/lib/cn';

interface StatCardProps {
  label: string;
  /** Accepts a string or a node (e.g. <Money>) for privacy-aware values. */
  value: ReactNode;
  icon: LucideIcon;
  /** Accent tone for the icon chip. */
  tone?: 'primary' | 'red' | 'sky' | 'amber';
  /** Optional secondary line under the value. */
  caption?: string;
  /** Optional trend indicator. */
  trend?: 'up' | 'down';
  trendLabel?: string;
}

const tones = {
  primary: 'bg-primary/10 text-primary',
  red: 'bg-red-500/10 text-red-500',
  sky: 'bg-sky-500/10 text-sky-500',
  amber: 'bg-amber-500/10 text-amber-500',
};

export function StatCard({
  label,
  value,
  icon: Icon,
  tone = 'primary',
  caption,
  trend,
  trendLabel,
}: StatCardProps) {
  return (
    <div className="card group p-5 transition-shadow hover:shadow-card">
      <div className="flex items-start justify-between">
        <p className="text-sm font-medium text-gray-500 dark:text-gray-400">{label}</p>
        <span className={cn('flex h-9 w-9 items-center justify-center rounded-lg', tones[tone])}>
          <Icon className="h-[18px] w-[18px]" />
        </span>
      </div>
      <p className="mt-3 font-heading text-2xl font-bold tabular-nums text-gray-900 dark:text-gray-50">
        {value}
      </p>
      {(caption || trend) && (
        <div className="mt-1.5 flex items-center gap-1.5 text-xs">
          {trend && (
            <span
              className={cn(
                'inline-flex items-center gap-0.5 font-medium',
                trend === 'up' ? 'text-primary' : 'text-red-500',
              )}
            >
              {trend === 'up' ? (
                <ArrowUpRight className="h-3.5 w-3.5" />
              ) : (
                <ArrowDownRight className="h-3.5 w-3.5" />
              )}
              {trendLabel}
            </span>
          )}
          {caption && <span className="text-gray-400">{caption}</span>}
        </div>
      )}
    </div>
  );
}
