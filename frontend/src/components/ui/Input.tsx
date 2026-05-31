import { forwardRef, type InputHTMLAttributes, type ReactNode } from 'react';
import { cn } from '@/lib/cn';

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  hint?: string;
  leftIcon?: ReactNode;
  rightAddon?: ReactNode;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, hint, leftIcon, rightAddon, className, id, ...props }, ref) => {
    const inputId = id ?? props.name;
    return (
      <div className="w-full">
        {label && (
          <label
            htmlFor={inputId}
            className="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300"
          >
            {label}
          </label>
        )}
        <div className="relative">
          {leftIcon && (
            <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
              {leftIcon}
            </span>
          )}
          <input
            ref={ref}
            id={inputId}
            className={cn(
              'input-base',
              leftIcon && 'pl-9',
              rightAddon && 'pr-12',
              error &&
                'border-red-500 focus:border-red-500 focus:ring-red-500/30 dark:border-red-500',
              className,
            )}
            aria-invalid={error ? true : undefined}
            {...props}
          />
          {rightAddon && (
            <span className="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-gray-400">
              {rightAddon}
            </span>
          )}
        </div>
        {error ? (
          <p className="mt-1 text-xs text-red-500">{error}</p>
        ) : hint ? (
          <p className="mt-1 text-xs text-gray-400">{hint}</p>
        ) : null}
      </div>
    );
  },
);
Input.displayName = 'Input';
