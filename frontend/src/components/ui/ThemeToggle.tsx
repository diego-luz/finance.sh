import { Moon, Sun } from 'lucide-react';
import { useThemeStore } from '@/stores/themeStore';
import { cn } from '@/lib/cn';

export function ThemeToggle({ className }: { className?: string }) {
  const theme = useThemeStore((s) => s.theme);
  const toggle = useThemeStore((s) => s.toggleTheme);
  const isDark = theme === 'dark';

  return (
    <button
      onClick={toggle}
      className={cn(
        'flex h-9 w-9 items-center justify-center rounded-lg text-gray-500 transition hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-ink-elevated dark:hover:text-gray-100',
        className,
      )}
      aria-label={isDark ? 'Ativar modo claro' : 'Ativar modo escuro'}
      title={isDark ? 'Modo claro' : 'Modo escuro'}
    >
      {isDark ? <Sun className="h-[18px] w-[18px]" /> : <Moon className="h-[18px] w-[18px]" />}
    </button>
  );
}
