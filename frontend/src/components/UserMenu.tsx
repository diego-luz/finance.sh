import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { ChevronDown, LogOut, Settings, User as UserIcon } from 'lucide-react';
import { Avatar } from '@/components/ui';
import { useAuthStore } from '@/stores/authStore';
import { useLogout } from '@/hooks';

export function UserMenu() {
  const { t } = useTranslation();
  const user = useAuthStore((s) => s.user);
  const logout = useLogout();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const onClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', onClick);
    return () => document.removeEventListener('mousedown', onClick);
  }, []);

  if (!user) return null;

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((v) => !v)}
        className="flex items-center gap-2 rounded-lg p-1 pr-2 transition hover:bg-gray-100 dark:hover:bg-ink-elevated"
        aria-label={t('nav.userMenu')}
      >
        <Avatar name={user.name} src={user.avatar_url} size="sm" />
        <ChevronDown className="hidden h-4 w-4 text-gray-400 sm:block" />
      </button>

      {open && (
        <div className="absolute right-0 z-30 mt-2 w-56 animate-fade-in overflow-hidden rounded-xl border border-gray-200 bg-white p-1.5 shadow-card dark:border-ink-border dark:bg-ink-surface">
          <div className="flex items-center gap-3 px-2.5 py-2">
            <Avatar name={user.name} src={user.avatar_url} size="md" />
            <div className="min-w-0">
              <p className="truncate text-sm font-medium text-gray-900 dark:text-gray-100">
                {user.name}
              </p>
              <p className="truncate text-xs text-gray-400">{user.email}</p>
            </div>
          </div>
          <div className="my-1 h-px bg-gray-100 dark:bg-ink-border" />
          <button
            onClick={() => {
              setOpen(false);
              navigate('/settings');
            }}
            className="flex w-full items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm text-gray-700 transition hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-ink-elevated"
          >
            <UserIcon className="h-4 w-4" /> {t('nav.myProfile')}
          </button>
          <button
            onClick={() => {
              setOpen(false);
              navigate('/settings');
            }}
            className="flex w-full items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm text-gray-700 transition hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-ink-elevated"
          >
            <Settings className="h-4 w-4" /> {t('nav.settings')}
          </button>
          <div className="my-1 h-px bg-gray-100 dark:bg-ink-border" />
          <button
            onClick={() => logout.mutate()}
            className="flex w-full items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm text-red-600 transition hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-500/10"
          >
            <LogOut className="h-4 w-4" /> {t('nav.logout')}
          </button>
        </div>
      )}
    </div>
  );
}
