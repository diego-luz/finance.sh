import { Eye, EyeOff } from 'lucide-react';
import { usePrivacyStore } from '@/stores/privacyStore';
import { cn } from '@/lib/cn';

export function PrivacyToggle({ className }: { className?: string }) {
  const hidden = usePrivacyStore((s) => s.hidden);
  const toggle = usePrivacyStore((s) => s.toggle);

  return (
    <button
      onClick={toggle}
      className={cn(
        'flex h-9 w-9 items-center justify-center rounded-lg text-gray-500 transition hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-ink-elevated dark:hover:text-gray-100',
        className,
      )}
      aria-label={hidden ? 'Mostrar valores' : 'Ocultar valores'}
      aria-pressed={hidden}
      title={hidden ? 'Mostrar valores' : 'Ocultar valores'}
    >
      {hidden ? <EyeOff className="h-[18px] w-[18px]" /> : <Eye className="h-[18px] w-[18px]" />}
    </button>
  );
}
