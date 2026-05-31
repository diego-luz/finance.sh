import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Globe, Check } from 'lucide-react';
import { SUPPORTED_LANGUAGES } from '@/i18n';
import { cn } from '@/lib/cn';

/**
 * Compact globe button + popover to switch the UI language from the topbar.
 * Selecting a language calls `i18n.changeLanguage` (persisted by the detector
 * cache) and re-renders translated surfaces instantly.
 */
export function LanguageMenu() {
  const { i18n, t } = useTranslation();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  const current = (i18n.resolvedLanguage ?? i18n.language) as string;

  useEffect(() => {
    const onClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', onClick);
    return () => document.removeEventListener('mousedown', onClick);
  }, []);

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((v) => !v)}
        className="rounded-lg p-1.5 text-gray-500 transition hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-ink-elevated"
        aria-label={t('language.label')}
        aria-haspopup="menu"
        aria-expanded={open}
      >
        <Globe className="h-5 w-5" />
      </button>

      {open && (
        <div
          role="menu"
          className="absolute right-0 z-30 mt-2 w-44 animate-fade-in overflow-hidden rounded-xl border border-gray-200 bg-white p-1.5 shadow-card dark:border-ink-border dark:bg-ink-surface"
        >
          <p className="px-2.5 py-1.5 text-[11px] font-semibold uppercase tracking-wider text-gray-400">
            {t('language.label')}
          </p>
          {SUPPORTED_LANGUAGES.map((code) => {
            const active = current === code;
            return (
              <button
                key={code}
                role="menuitemradio"
                aria-checked={active}
                onClick={() => {
                  void i18n.changeLanguage(code);
                  setOpen(false);
                }}
                className={cn(
                  'flex w-full items-center justify-between gap-2.5 rounded-lg px-2.5 py-2 text-sm transition',
                  active
                    ? 'bg-primary/10 font-medium text-primary-700 dark:text-primary-300'
                    : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-ink-elevated',
                )}
              >
                {t(`language.${code}`)}
                {active && <Check className="h-4 w-4 text-primary" />}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
