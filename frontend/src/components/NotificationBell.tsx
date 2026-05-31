import { useEffect, useRef, useState } from 'react';
import { Bell, CheckCheck, Info, AlertTriangle, CheckCircle2, AlertCircle } from 'lucide-react';
import { cn } from '@/lib/cn';
import { formatDateShort } from '@/lib/date';
import {
  useNotifications,
  useMarkNotificationRead,
  useMarkAllNotificationsRead,
} from '@/hooks';
import type { Notification, NotificationType } from '@/types';

function typeIcon(type: NotificationType) {
  switch (type) {
    case 'success':
    case 'goal':
      return { Icon: CheckCircle2, color: 'text-primary' };
    case 'warning':
    case 'budget':
    case 'bill':
      return { Icon: AlertTriangle, color: 'text-amber-500' };
    case 'error':
      return { Icon: AlertCircle, color: 'text-red-500' };
    default:
      return { Icon: Info, color: 'text-sky-500' };
  }
}

function NotificationItem({
  item,
  onRead,
}: {
  item: Notification;
  onRead: (id: string) => void;
}) {
  const { Icon, color } = typeIcon(item.type);
  return (
    <button
      onClick={() => !item.read && onRead(item.id)}
      className={cn(
        'flex w-full items-start gap-3 px-3 py-2.5 text-left transition hover:bg-gray-50 dark:hover:bg-ink-elevated',
        !item.read && 'bg-primary/5 dark:bg-primary/10',
      )}
    >
      <Icon className={cn('mt-0.5 h-4 w-4 shrink-0', color)} />
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium text-gray-900 dark:text-gray-100">
          {item.title}
        </p>
        <p className="line-clamp-2 text-xs text-gray-500 dark:text-gray-400">{item.message}</p>
        <p className="mt-0.5 text-[11px] text-gray-400">{formatDateShort(item.created_at)}</p>
      </div>
      {!item.read && <span className="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-primary" />}
    </button>
  );
}

export function NotificationBell() {
  const { data: notifications, isLoading } = useNotifications();
  const markRead = useMarkNotificationRead();
  const markAll = useMarkAllNotificationsRead();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const onClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', onClick);
    return () => document.removeEventListener('mousedown', onClick);
  }, []);

  const items = notifications ?? [];
  const unread = items.filter((n) => !n.read).length;

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((v) => !v)}
        className="relative flex h-9 w-9 items-center justify-center rounded-lg text-gray-500 transition hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-ink-elevated"
        aria-label="Notificações"
      >
        <Bell className="h-[18px] w-[18px]" />
        {unread > 0 && (
          <span className="absolute -right-0.5 -top-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-primary px-1 text-[10px] font-bold text-white">
            {unread > 9 ? '9+' : unread}
          </span>
        )}
      </button>

      {open && (
        <div className="absolute right-0 z-30 mt-2 w-80 animate-fade-in overflow-hidden rounded-xl border border-gray-200 bg-white shadow-card dark:border-ink-border dark:bg-ink-surface">
          <div className="flex items-center justify-between border-b border-gray-100 px-4 py-3 dark:border-ink-border">
            <p className="text-sm font-semibold text-gray-900 dark:text-gray-100">Notificações</p>
            {unread > 0 && (
              <button
                onClick={() => markAll.mutate()}
                disabled={markAll.isPending}
                className="inline-flex items-center gap-1 text-xs font-medium text-primary hover:text-primary-600 disabled:opacity-60"
              >
                <CheckCheck className="h-3.5 w-3.5" /> Marcar todas como lidas
              </button>
            )}
          </div>

          <div className="max-h-96 overflow-y-auto">
            {isLoading ? (
              <div className="space-y-2 p-3">
                {Array.from({ length: 3 }).map((_, i) => (
                  <div key={i} className="skeleton h-12 w-full rounded-lg" />
                ))}
              </div>
            ) : items.length === 0 ? (
              <div className="flex flex-col items-center gap-2 px-4 py-10 text-center">
                <span className="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/10 text-primary">
                  <Bell className="h-5 w-5" />
                </span>
                <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                  Tudo em dia
                </p>
                <p className="text-xs text-gray-400">Você não tem notificações no momento.</p>
              </div>
            ) : (
              <div className="divide-y divide-gray-100 dark:divide-ink-border">
                {items.map((item) => (
                  <NotificationItem key={item.id} item={item} onRead={(id) => markRead.mutate(id)} />
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
