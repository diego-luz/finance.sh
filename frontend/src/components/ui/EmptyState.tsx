import type { LucideIcon } from 'lucide-react';
import type { ReactNode } from 'react';

interface EmptyStateProps {
  icon: LucideIcon;
  title: string;
  description?: string;
  action?: ReactNode;
}

export function EmptyState({ icon: Icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-gray-300 px-6 py-14 text-center dark:border-ink-border">
      <div className="mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/10 text-primary">
        <Icon className="h-7 w-7" />
      </div>
      <h3 className="font-heading text-base font-semibold text-gray-900 dark:text-gray-100">
        {title}
      </h3>
      {description && (
        <p className="mt-1 max-w-sm text-sm text-gray-500 dark:text-gray-400">{description}</p>
      )}
      {action && <div className="mt-5">{action}</div>}
    </div>
  );
}
