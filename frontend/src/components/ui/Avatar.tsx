import { cn } from '@/lib/cn';

interface AvatarProps {
  name?: string;
  src?: string;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

const sizes = {
  sm: 'h-8 w-8 text-xs',
  md: 'h-9 w-9 text-sm',
  lg: 'h-12 w-12 text-base',
};

function initials(name?: string): string {
  if (!name) return '?';
  const parts = name.trim().split(/\s+/);
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

export function Avatar({ name, src, size = 'md', className }: AvatarProps) {
  if (src) {
    return (
      <img
        src={src}
        alt={name ?? 'Avatar'}
        className={cn('rounded-full object-cover', sizes[size], className)}
      />
    );
  }
  return (
    <div
      className={cn(
        'flex items-center justify-center rounded-full bg-primary/15 font-semibold text-primary-700 dark:text-primary-300',
        sizes[size],
        className,
      )}
      aria-hidden
    >
      {initials(name)}
    </div>
  );
}
