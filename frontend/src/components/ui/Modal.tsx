import { useEffect, type ReactNode } from 'react';
import { createPortal } from 'react-dom';
import { X } from 'lucide-react';
import { cn } from '@/lib/cn';

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title?: string;
  description?: string;
  children: ReactNode;
  footer?: ReactNode;
  /** 'modal' = centered dialog, 'drawer' = right-side slide-over. */
  variant?: 'modal' | 'drawer';
  size?: 'sm' | 'md' | 'lg';
}

const sizes = {
  sm: 'max-w-sm',
  md: 'max-w-lg',
  lg: 'max-w-2xl',
};

export function Modal({
  open,
  onClose,
  title,
  description,
  children,
  footer,
  variant = 'modal',
  size = 'md',
}: ModalProps) {
  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', onKey);
    document.body.style.overflow = 'hidden';
    return () => {
      document.removeEventListener('keydown', onKey);
      document.body.style.overflow = '';
    };
  }, [open, onClose]);

  if (!open) return null;

  const isDrawer = variant === 'drawer';

  return createPortal(
    <div
      className="fixed inset-0 z-50 flex"
      role="dialog"
      aria-modal="true"
      aria-label={title}
    >
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={onClose}
        aria-hidden
      />
      <div
        className={cn(
          'relative z-10 flex w-full flex-col bg-white shadow-card dark:bg-ink-surface',
          isDrawer
            ? 'ml-auto h-full max-w-md animate-slide-in-right'
            : cn(
                'm-auto max-h-[90vh] animate-fade-in rounded-2xl',
                sizes[size],
              ),
        )}
      >
        {(title || description) && (
          <div className="flex items-start justify-between gap-4 border-b border-gray-100 px-6 py-4 dark:border-ink-border">
            <div>
              {title && (
                <h2 className="font-heading text-lg font-semibold text-gray-900 dark:text-gray-100">
                  {title}
                </h2>
              )}
              {description && (
                <p className="mt-0.5 text-sm text-gray-500 dark:text-gray-400">{description}</p>
              )}
            </div>
            <button
              onClick={onClose}
              className="rounded-lg p-1.5 text-gray-400 transition hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-ink-elevated"
              aria-label="Fechar"
            >
              <X className="h-5 w-5" />
            </button>
          </div>
        )}
        <div className="flex-1 overflow-y-auto px-6 py-5">{children}</div>
        {footer && (
          <div className="flex items-center justify-end gap-3 border-t border-gray-100 px-6 py-4 dark:border-ink-border">
            {footer}
          </div>
        )}
      </div>
    </div>,
    document.body,
  );
}
