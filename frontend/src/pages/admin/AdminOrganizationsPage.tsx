import { useEffect, useState } from 'react';
import { Navigate } from 'react-router-dom';
import { Building2, Search, ChevronLeft, ChevronRight } from 'lucide-react';
import { PageHeader } from '@/components/PageHeader';
import {
  Button,
  Card,
  EmptyState,
  Input,
  Table,
  THead,
  Th,
  TBody,
  Tr,
  Td,
  SkeletonRows,
} from '@/components/ui';
import { useAdminOrganizations, useDebouncedValue } from '@/hooks';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';
import { formatDateShort } from '@/lib/date';

const PER_PAGE = 20;

export function AdminOrganizationsPage() {
  const user = useAuthStore((s) => s.user);
  const toast = useToast();

  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);

  const debouncedSearch = useDebouncedValue(search, 400);
  const { data, isLoading, isFetching, isError } = useAdminOrganizations({
    search: debouncedSearch || undefined,
    page,
    per_page: PER_PAGE,
  });

  useEffect(() => {
    if (isError) toast.error('Não foi possível carregar as organizações.');
  }, [isError, toast]);

  // Reset to page 1 whenever the (debounced) search changes.
  useEffect(() => {
    setPage(1);
  }, [debouncedSearch]);

  // Hooks run before the guard to preserve hook order (mirrors MembersPage).
  if (!user?.super_admin) {
    return <Navigate to="/" replace />;
  }

  const meta = data?.meta;
  const pages = meta?.pages ?? 1;
  const total = meta?.total ?? 0;
  const orgs = data?.data ?? [];

  return (
    <div>
      <PageHeader
        eyebrow="Plataforma"
        title="Organizações"
        description="Visão geral de todas as organizações da instância."
      />

      <Card className="mb-4 p-4">
        <Input
          placeholder="Buscar por nome, slug ou e-mail do dono…"
          leftIcon={<Search className="h-4 w-4" />}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          aria-label="Buscar organizações"
        />
      </Card>

      <Card className="p-0">
        {isLoading ? (
          <Table>
            <THead>
              <Th>Organização</Th>
              <Th>Dono</Th>
              <Th>Moeda</Th>
              <Th className="text-right">Membros</Th>
              <Th className="text-right">Transações</Th>
              <Th>Criada em</Th>
            </THead>
            <TBody>
              <SkeletonRows rows={8} cols={6} />
            </TBody>
          </Table>
        ) : orgs.length === 0 ? (
          <div className="p-6">
            <EmptyState
              icon={Building2}
              title="Nenhuma organização"
              description="Ajuste a busca."
            />
          </div>
        ) : (
          <Table>
            <THead>
              <Th>Organização</Th>
              <Th>Dono</Th>
              <Th>Moeda</Th>
              <Th className="text-right">Membros</Th>
              <Th className="text-right">Transações</Th>
              <Th>Criada em</Th>
            </THead>
            <TBody>
              {orgs.map((org) => (
                <Tr key={org.id}>
                  <Td>
                    <div className="min-w-0">
                      <p className="truncate font-medium text-gray-900 dark:text-gray-100">
                        {org.name}
                      </p>
                      <p className="truncate text-xs text-gray-400">{org.slug}</p>
                    </div>
                  </Td>
                  <Td>
                    <span className="text-gray-600 dark:text-gray-300" title={org.owner?.email ?? '—'}>
                      {org.owner?.email ?? '—'}
                    </span>
                  </Td>
                  <Td className="text-gray-600 dark:text-gray-300">{org.currency}</Td>
                  <Td className="text-right tabular-nums">{org.members}</Td>
                  <Td className="text-right tabular-nums">{org.transactions}</Td>
                  <Td className="whitespace-nowrap text-gray-500 dark:text-gray-400">
                    {formatDateShort(org.created_at)}
                  </Td>
                </Tr>
              ))}
            </TBody>
          </Table>
        )}

        {orgs.length > 0 && (
          <div className="flex items-center justify-between border-t border-gray-100 px-4 py-3 dark:border-ink-border">
            <p className="text-xs text-gray-500 dark:text-gray-400">
              {total} organização{total === 1 ? '' : 's'} · página {page} de {pages}
              {isFetching && <span className="ml-2 text-primary">atualizando…</span>}
            </p>
            <div className="flex items-center gap-1.5">
              <Button
                variant="outline"
                size="sm"
                disabled={page <= 1 || isFetching}
                onClick={() => setPage((p) => p - 1)}
              >
                <ChevronLeft className="h-4 w-4" /> Anterior
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={page >= pages || isFetching}
                onClick={() => setPage((p) => p + 1)}
              >
                Próxima <ChevronRight className="h-4 w-4" />
              </Button>
            </div>
          </div>
        )}
      </Card>
    </div>
  );
}
