import { cn } from '@/lib/cn';

interface SkeletonProps {
  className?: string;
}

export function Skeleton({ className }: SkeletonProps) {
  return <div className={cn('skeleton h-4 w-full', className)} />;
}

/** A card-shaped skeleton placeholder. */
export function SkeletonCard({ className }: SkeletonProps) {
  return (
    <div className={cn('card space-y-3 p-5', className)}>
      <Skeleton className="h-3 w-1/3" />
      <Skeleton className="h-8 w-2/3" />
      <Skeleton className="h-3 w-1/2" />
    </div>
  );
}

/** Repeated row skeletons for tables. */
export function SkeletonRows({ rows = 5, cols = 4 }: { rows?: number; cols?: number }) {
  return (
    <>
      {Array.from({ length: rows }).map((_, r) => (
        <tr key={r} className="border-b border-gray-100 dark:border-ink-border">
          {Array.from({ length: cols }).map((__, c) => (
            <td key={c} className="px-4 py-3.5">
              <Skeleton className="h-4" />
            </td>
          ))}
        </tr>
      ))}
    </>
  );
}
