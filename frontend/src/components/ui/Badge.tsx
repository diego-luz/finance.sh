import type { ReactNode } from 'react';
import { cn } from '@/lib/cn';

export type BadgeVariant =
  | 'success'
  | 'danger'
  | 'warning'
  | 'info'
  | 'neutral';

interface BadgeProps {
  variant?: BadgeVariant;
  children: ReactNode;
  dot?: boolean;
  className?: string;
}

const variants: Record<BadgeVariant, string> = {
  success:
    'bg-primary/10 text-primary-700 dark:text-primary-300 dark:bg-primary/15',
  danger: 'bg-red-500/10 text-red-600 dark:text-red-400',
  warning: 'bg-amber-500/10 text-amber-600 dark:text-amber-400',
  info: 'bg-sky-500/10 text-sky-600 dark:text-sky-400',
  neutral: 'bg-gray-500/10 text-gray-600 dark:text-gray-300',
};

const dotColors: Record<BadgeVariant, string> = {
  success: 'bg-primary',
  danger: 'bg-red-500',
  warning: 'bg-amber-500',
  info: 'bg-sky-500',
  neutral: 'bg-gray-400',
};

export function Badge({ variant = 'neutral', children, dot, className }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium',
        variants[variant],
        className,
      )}
    >
      {dot && <span className={cn('h-1.5 w-1.5 rounded-full', dotColors[variant])} />}
      {children}
    </span>
  );
}
