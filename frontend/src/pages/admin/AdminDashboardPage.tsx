import { useEffect } from 'react';
import { Navigate } from 'react-router-dom';
import { Building2, Users, ArrowLeftRight } from 'lucide-react';
import { PageHeader } from '@/components/PageHeader';
import { SkeletonCard, StatCard } from '@/components/ui';
import { useAdminStats } from '@/hooks';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';

/** Formats an integer counter with pt-BR thousands separators. */
function fmt(n: number | undefined): string {
  return new Intl.NumberFormat('pt-BR').format(n ?? 0);
}

export function AdminDashboardPage() {
  const user = useAuthStore((s) => s.user);
  const toast = useToast();
  const { data, isLoading, isError } = useAdminStats();

  useEffect(() => {
    if (isError) toast.error('Não foi possível carregar as estatísticas da plataforma.');
  }, [isError, toast]);

  // Hooks run before the guard to preserve hook order (mirrors MembersPage).
  if (!user?.super_admin) {
    return <Navigate to="/" replace />;
  }

  return (
    <div>
      <PageHeader
        eyebrow="Plataforma"
        title="Painel administrativo"
        description="Visão geral de toda a plataforma finance.sh."
      />

      {isLoading ? (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <SkeletonCard key={i} />
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
          <StatCard label="Organizações" value={fmt(data?.organizations)} icon={Building2} tone="primary" />
          <StatCard label="Usuários" value={fmt(data?.users)} icon={Users} tone="sky" />
          <StatCard
            label="Transações"
            value={fmt(data?.transactions)}
            icon={ArrowLeftRight}
            tone="primary"
          />
        </div>
      )}
    </div>
  );
}
