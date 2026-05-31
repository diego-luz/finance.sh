import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from 'react';
import { CheckCircle2, AlertCircle, Info, X } from 'lucide-react';
import { cn } from '@/lib/cn';

export type ToastVariant = 'success' | 'error' | 'info';

interface Toast {
  id: number;
  title?: string;
  message: string;
  variant: ToastVariant;
}

interface ToastContextValue {
  toast: (message: string, variant?: ToastVariant, title?: string) => void;
  success: (message: string, title?: string) => void;
  error: (message: string, title?: string) => void;
  info: (message: string, title?: string) => void;
}

const ToastContext = createContext<ToastContextValue | null>(null);

let counter = 0;

const variantConfig: Record<
  ToastVariant,
  { icon: typeof Info; accent: string; iconColor: string }
> = {
  success: {
    icon: CheckCircle2,
    accent: 'border-l-primary',
    iconColor: 'text-primary',
  },
  error: {
    icon: AlertCircle,
    accent: 'border-l-red-500',
    iconColor: 'text-red-500',
  },
  info: {
    icon: Info,
    accent: 'border-l-sky-500',
    iconColor: 'text-sky-500',
  },
};

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const dismiss = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const toast = useCallback(
    (message: string, variant: ToastVariant = 'info', title?: string) => {
      const id = ++counter;
      setToasts((prev) => [...prev, { id, message, variant, title }]);
      window.setTimeout(() => dismiss(id), 4500);
    },
    [dismiss],
  );

  const value = useMemo<ToastContextValue>(
    () => ({
      toast,
      success: (m, t) => toast(m, 'success', t),
      error: (m, t) => toast(m, 'error', t),
      info: (m, t) => toast(m, 'info', t),
    }),
    [toast],
  );

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div className="pointer-events-none fixed bottom-4 right-4 z-[100] flex w-full max-w-sm flex-col gap-2">
        {toasts.map((t) => {
          const { icon: Icon, accent, iconColor } = variantConfig[t.variant];
          return (
            <div
              key={t.id}
              role="status"
              className={cn(
                'pointer-events-auto flex animate-fade-in items-start gap-3 rounded-xl border border-l-4 bg-white p-4 shadow-card dark:bg-ink-elevated',
                accent,
              )}
            >
              <Icon className={cn('mt-0.5 h-5 w-5 shrink-0', iconColor)} />
              <div className="min-w-0 flex-1">
                {t.title && (
                  <p className="text-sm font-semibold text-gray-900 dark:text-gray-100">
                    {t.title}
                  </p>
                )}
                <p className="text-sm text-gray-600 dark:text-gray-300">{t.message}</p>
              </div>
              <button
                onClick={() => dismiss(t.id)}
                className="shrink-0 rounded-md p-0.5 text-gray-400 transition hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-ink-surface"
                aria-label="Fechar notificação"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
          );
        })}
      </div>
    </ToastContext.Provider>
  );
}

// eslint-disable-next-line react-refresh/only-export-components
export function useToast(): ToastContextValue {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error('useToast must be used within ToastProvider');
  return ctx;
}
