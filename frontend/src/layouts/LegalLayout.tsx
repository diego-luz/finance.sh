import type { ReactNode } from 'react';
import { Link } from 'react-router-dom';
import { ArrowLeft } from 'lucide-react';
import { Logo } from '@/components/Logo';
import { ThemeToggle } from '@/components/ui';

interface LegalLayoutProps {
  title: string;
  /** A short "última atualização" line. */
  updatedAt: string;
  children: ReactNode;
}

/** Centered reading layout for public legal/policy pages (LGPD-oriented). */
export function LegalLayout({ title, updatedAt, children }: LegalLayoutProps) {
  return (
    <div className="min-h-screen bg-gray-100 dark:bg-ink-base">
      <header className="sticky top-0 z-20 border-b border-gray-200 bg-white/80 backdrop-blur dark:border-ink-border dark:bg-ink-base/80">
        <div className="mx-auto flex h-16 max-w-3xl items-center justify-between px-4 sm:px-6">
          <Link to="/" aria-label="Início">
            <Logo />
          </Link>
          <ThemeToggle />
        </div>
      </header>

      <main className="mx-auto max-w-3xl px-4 py-10 sm:px-6">
        <Link
          to="/register"
          className="mb-6 inline-flex items-center gap-1.5 text-sm font-medium text-primary hover:text-primary-600"
        >
          <ArrowLeft className="h-4 w-4" /> Voltar
        </Link>

        <h1 className="font-heading text-3xl font-extrabold text-gray-900 dark:text-white">
          {title}
        </h1>
        <p className="mt-2 text-sm text-gray-400">Última atualização: {updatedAt}</p>

        <article className="legal-prose mt-8 space-y-6 text-sm leading-relaxed text-gray-600 dark:text-gray-300">
          {children}
        </article>

        <footer className="mt-12 border-t border-gray-200 pt-6 text-xs text-gray-400 dark:border-ink-border">
          <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
            <Link to="/termos" className="hover:text-primary">
              Termos de Uso
            </Link>
            <Link to="/privacidade" className="hover:text-primary">
              Política de Privacidade
            </Link>
            <Link to="/login" className="hover:text-primary">
              Entrar
            </Link>
          </div>
          <p className="mt-4">
            © {new Date().getFullYear()} finance.sh. Todos os direitos reservados.
          </p>
        </footer>
      </main>
    </div>
  );
}

/** Section heading helper used inside legal pages. */
export function LegalSection({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="space-y-2">
      <h2 className="font-heading text-lg font-semibold text-gray-900 dark:text-gray-100">
        {title}
      </h2>
      {children}
    </section>
  );
}
