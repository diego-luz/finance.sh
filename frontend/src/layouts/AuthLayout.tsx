import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { ShieldCheck, TrendingUp, Sparkles } from 'lucide-react';
import { Logo } from '@/components/Logo';
import { ThemeToggle } from '@/components/ui';
import { LanguageMenu } from '@/components/LanguageMenu';

interface AuthLayoutProps {
  children: ReactNode;
  title: string;
  subtitle: string;
}

/** Split layout: branded emerald panel on the left, form card on the right. */
export function AuthLayout({ children, title, subtitle }: AuthLayoutProps) {
  const { t } = useTranslation();
  return (
    <div className="flex min-h-screen bg-gray-100 dark:bg-ink-base">
      {/* Brand panel */}
      <div className="relative hidden w-1/2 overflow-hidden bg-gradient-to-br from-primary-600 via-primary to-emerald-700 lg:flex">
        <div className="absolute inset-0 opacity-20 [background-image:radial-gradient(circle_at_20%_20%,white_1px,transparent_1px)] [background-size:28px_28px]" />
        <div className="relative z-10 flex w-full flex-col justify-between p-12 text-white">
          <Logo markClassName="bg-white/15 backdrop-blur" className="[&_span]:text-white" />

          <div className="max-w-md">
            <p className="mb-3 inline-flex items-center gap-2 rounded-full bg-white/15 px-3 py-1 text-xs font-medium backdrop-blur">
              <Sparkles className="h-3.5 w-3.5" /> {t('auth.tagline')}
            </p>
            <h1 className="font-heading text-4xl font-extrabold leading-tight">
              {t('auth.brandHeadline')}
            </h1>
            <p className="mt-4 text-base text-white/80">
              {t('auth.brandSubtitle')}
            </p>

            <ul className="mt-8 space-y-3 text-sm text-white/90">
              <li className="flex items-center gap-3">
                <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-white/15">
                  <TrendingUp className="h-4 w-4" />
                </span>
                {t('auth.feature1')}
              </li>
              <li className="flex items-center gap-3">
                <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-white/15">
                  <ShieldCheck className="h-4 w-4" />
                </span>
                {t('auth.feature2')}
              </li>
            </ul>
          </div>

          <p className="text-xs text-white/60">
            {t('footer.rights', { year: new Date().getFullYear() })}
          </p>
        </div>
      </div>

      {/* Form side */}
      <div className="flex w-full flex-col lg:w-1/2">
        <div className="flex items-center justify-between p-6">
          <div className="lg:hidden">
            <Logo />
          </div>
          <div className="ml-auto flex items-center gap-1.5">
            <LanguageMenu />
            <ThemeToggle />
          </div>
        </div>

        <div className="flex flex-1 items-center justify-center px-6 pb-12">
          <div className="w-full max-w-md">
            <div className="mb-8">
              <h2 className="font-heading text-2xl font-bold text-gray-900 dark:text-white">
                {title}
              </h2>
              <p className="mt-1.5 text-sm text-gray-500 dark:text-gray-400">{subtitle}</p>
            </div>
            {children}
          </div>
        </div>
      </div>
    </div>
  );
}
