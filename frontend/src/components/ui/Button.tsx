import { forwardRef, type ButtonHTMLAttributes } from 'react';
import { Loader2 } from 'lucide-react';
import { cn } from '@/lib/cn';

export type ButtonVariant =
  | 'primary'
  | 'secondary'
  | 'ghost'
  | 'danger'
  | 'outline';
export type ButtonSize = 'sm' | 'md' | 'lg' | 'icon';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  loading?: boolean;
}

const variants: Record<ButtonVariant, string> = {
  primary:
    'bg-primary text-white shadow-sm hover:bg-primary-600 active:bg-primary-700 disabled:bg-primary/50',
  secondary:
    'bg-gray-100 text-gray-900 hover:bg-gray-200 dark:bg-ink-elevated dark:text-gray-100 dark:hover:bg-ink-border',
  ghost:
    'bg-transparent text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-ink-elevated',
  outline:
    'border border-gray-300 bg-transparent text-gray-800 hover:bg-gray-50 dark:border-ink-border dark:text-gray-200 dark:hover:bg-ink-elevated',
  danger:
    'bg-red-500 text-white hover:bg-red-600 active:bg-red-700 disabled:bg-red-500/50',
};

const sizes: Record<ButtonSize, string> = {
  sm: 'h-8 px-3 text-xs gap-1.5',
  md: 'h-10 px-4 text-sm gap-2',
  lg: 'h-12 px-6 text-base gap-2',
  icon: 'h-10 w-10 p-0',
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (
    { variant = 'primary', size = 'md', loading, className, children, disabled, ...props },
    ref,
  ) => (
    <button
      ref={ref}
      disabled={disabled || loading}
      className={cn(
        'inline-flex items-center justify-center whitespace-nowrap rounded-lg font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-70',
        variants[variant],
        sizes[size],
        className,
      )}
      {...props}
    >
      {loading && <Loader2 className="h-4 w-4 animate-spin" />}
      {children}
    </button>
  ),
);
Button.displayName = 'Button';
