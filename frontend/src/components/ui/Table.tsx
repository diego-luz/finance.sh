import type { HTMLAttributes, ReactNode, ThHTMLAttributes } from 'react';
import { cn } from '@/lib/cn';

export function Table({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <div className="overflow-x-auto">
      <table className={cn('w-full border-collapse text-sm', className)}>{children}</table>
    </div>
  );
}

export function THead({ children }: { children: ReactNode }) {
  return (
    <thead className="border-b border-gray-200 dark:border-ink-border">
      <tr>{children}</tr>
    </thead>
  );
}

export function Th({
  children,
  className,
  ...props
}: ThHTMLAttributes<HTMLTableCellElement> & { children?: ReactNode }) {
  return (
    <th
      className={cn(
        'px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400',
        className,
      )}
      {...props}
    >
      {children}
    </th>
  );
}

export function TBody({ children }: { children: ReactNode }) {
  return <tbody>{children}</tbody>;
}

export function Tr({
  children,
  className,
  ...props
}: HTMLAttributes<HTMLTableRowElement> & { children: ReactNode }) {
  return (
    <tr
      className={cn(
        'border-b border-gray-100 transition-colors last:border-0 hover:bg-gray-50 dark:border-ink-border dark:hover:bg-ink-elevated/60',
        className,
      )}
      {...props}
    >
      {children}
    </tr>
  );
}

export function Td({
  children,
  className,
  ...props
}: HTMLAttributes<HTMLTableCellElement> & { children?: ReactNode }) {
  return (
    <td
      className={cn('px-4 py-3.5 text-gray-700 dark:text-gray-300', className)}
      {...props}
    >
      {children}
    </td>
  );
}
