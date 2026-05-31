import type { HTMLAttributes, ReactNode } from 'react';
import { cn } from '@/lib/cn';

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
}

export function Card({ className, children, ...props }: CardProps) {
  return (
    <div className={cn('card p-5', className)} {...props}>
      {children}
    </div>
  );
}

interface CardHeaderProps {
  title: ReactNode;
  subtitle?: ReactNode;
  action?: ReactNode;
  eyebrow?: ReactNode;
  className?: string;
}

export function CardHeader({ title, subtitle, action, eyebrow, className }: CardHeaderProps) {
  return (
    <div className={cn('mb-4 flex items-start justify-between gap-3', className)}>
      <div className="min-w-0">
        {eyebrow && (
          <p className="mb-1 text-[11px] font-semibold uppercase tracking-wider text-primary">
            {eyebrow}
          </p>
        )}
        <h3 className="font-heading text-base font-semibold text-gray-900 dark:text-gray-100">
          {title}
        </h3>
        {subtitle && (
          <p className="mt-0.5 text-sm text-gray-500 dark:text-gray-400">{subtitle}</p>
        )}
      </div>
      {action && <div className="shrink-0">{action}</div>}
    </div>
  );
}
