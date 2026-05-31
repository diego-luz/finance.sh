import { cn } from '@/lib/cn';

interface LogoProps {
  /** Show the "finance.sh" wordmark next to the emerald mark. */
  showWord?: boolean;
  className?: string;
  markClassName?: string;
}

/** finance.sh brand: emerald rounded mark with a vault/coin glyph + wordmark. */
export function Logo({ showWord = true, className, markClassName }: LogoProps) {
  return (
    <div className={cn('flex items-center gap-2.5', className)}>
      <div
        className={cn(
          'flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-white shadow-glow',
          markClassName,
        )}
      >
        <svg viewBox="0 0 24 24" className="h-7 w-7" fill="none" aria-hidden>
          {/* Alvo (anel) */}
          <circle cx="12" cy="12" r="11" stroke="currentColor" strokeWidth="1.6" />
          {/* Prompt > */}
          <path
            d="M7.5 8.8L11 12L7.5 15.2"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
          {/* Cifrão */}
          <text
            x="14.8"
            y="15.9"
            textAnchor="middle"
            fontSize="11"
            fontFamily="monospace"
            fontWeight="700"
            fill="currentColor"
          >
            $
          </text>
        </svg>
      </div>
      {showWord && (
        <span className="font-heading text-lg font-extrabold tracking-tight text-gray-900 dark:text-white">
          finance.sh
        </span>
      )}
    </div>
  );
}
