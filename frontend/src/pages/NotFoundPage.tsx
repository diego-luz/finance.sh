import { Link } from 'react-router-dom';
import { Logo } from '@/components/Logo';
import { Button } from '@/components/ui';

export function NotFoundPage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-6 bg-gray-100 px-6 text-center dark:bg-ink-base">
      <Logo />
      <div>
        <p className="font-heading text-6xl font-extrabold text-primary">404</p>
        <h1 className="mt-2 font-heading text-xl font-semibold text-gray-900 dark:text-gray-100">
          Página não encontrada
        </h1>
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
          O endereço que você tentou acessar não existe ou foi movido.
        </p>
      </div>
      <Link to="/">
        <Button>Voltar ao início</Button>
      </Link>
    </div>
  );
}
