import { useEffect, useMemo, useState } from 'react';
import { Navigate } from 'react-router-dom';
import { ScrollText, ChevronLeft, ChevronRight } from 'lucide-react';
import { PageHeader } from '@/components/PageHeader';
import {
  Badge,
  Card,
  Button,
  Select,
  EmptyState,
  PeriodFilter,
  Table,
  THead,
  Th,
  TBody,
  Tr,
  Td,
  SkeletonRows,
} from '@/components/ui';
import type { BadgeVariant } from '@/components/ui';
import { useAuditLogs, useMembers, useIsAdmin } from '@/hooks';
import { useToast } from '@/contexts/ToastContext';
import { formatDate } from '@/lib/date';
import type { AuditFilters } from '@/types';

const PER_PAGE = 20;

const actionOptions = [
  { value: '', label: 'Todas as ações' },
  { value: 'POST', label: 'Criação (POST)' },
  { value: 'PUT', label: 'Edição (PUT)' },
  { value: 'DELETE', label: 'Exclusão (DELETE)' },
];

/** Maps an HTTP method to a humanized verb + color-coded badge variant. */
const actionLabels: Record<string, { label: string; variant: BadgeVariant }> = {
  POST: { label: 'Criou', variant: 'success' },
  PUT: { label: 'Editou', variant: 'warning' },
  PATCH: { label: 'Editou', variant: 'warning' },
  DELETE: { label: 'Excluiu', variant: 'danger' },
};

/** Maps an entity path segment to a friendly pt-BR label. */
const entityLabels: Record<string, string> = {
  transactions: 'Transação',
  accounts: 'Conta',
  categories: 'Categoria',
  tags: 'Tag',
  contacts: 'Contato',
  members: 'Membro',
  invitations: 'Convite',
  budgets: 'Orçamento',
  goals: 'Meta',
  'credit-cards': 'Cartão',
  recurrences: 'Recorrência',
  'categorization-rules': 'Regra de categorização',
  attachments: 'Anexo',
  subscription: 'Assinatura',
  organizations: 'Organização',
};

function humanizeEntity(entity: string): string {
  return entityLabels[entity] ?? entity;
}

function ActionBadge({ action }: { action: string }) {
  const cfg = actionLabels[action.toUpperCase()];
  if (!cfg) return <Badge variant="neutral">{action}</Badge>;
  return (
    <Badge variant={cfg.variant} dot>
      {cfg.label}
    </Badge>
  );
}

export function AuditPage() {
  const isAdmin = useIsAdmin();
  const toast = useToast();

  const [filters, setFilters] = useState<AuditFilters>({ page: 1, per_page: PER_PAGE });

  const { data, isLoading, isFetching, isError } = useAuditLogs(filters);
  const { data: members } = useMembers();

  // Surface load failures as a toast (e.g. 403 for non-admins reaching the API).
  useEffect(() => {
    if (isError) {
      toast.error('Não foi possível carregar os registros de auditoria.');
    }
  }, [isError, toast]);

  const userFilterOptions = useMemo(
    () => [
      { value: '', label: 'Todos os usuários' },
      ...(members ?? []).map((m) => ({ value: m.user.id, label: m.user.name })),
    ],
    [members],
  );

  // Route guard runs AFTER all hooks to preserve hook order (mirrors MembersPage).
  if (!isAdmin) {
    return <Navigate to="/" replace />;
  }

  const updateFilter = (patch: Partial<AuditFilters>) =>
    setFilters((prev) => ({ ...prev, ...patch, page: 1 }));

  const meta = data?.meta;
  const page = meta?.page ?? 1;
  const pages = meta?.pages ?? 1;
  const total = meta?.total ?? 0;
  const logs = data?.data ?? [];

  return (
    <div>
      <PageHeader
        eyebrow="Organização"
        title="Auditoria"
        description="Histórico de alterações realizadas na organização."
      />

      <Card className="mb-4 p-4">
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
          <Select
            label="Ação"
            options={actionOptions}
            value={filters.action ?? ''}
            onChange={(e) => updateFilter({ action: e.target.value || undefined })}
          />
          <Select
            label="Usuário"
            options={userFilterOptions}
            value={filters.user_id ?? ''}
            onChange={(e) => updateFilter({ user_id: e.target.value || undefined })}
          />
          <Select
            label="Entidade"
            options={[
              { value: '', label: 'Todas as entidades' },
              ...Object.entries(entityLabels).map(([value, label]) => ({ value, label })),
            ]}
            value={filters.entity ?? ''}
            onChange={(e) => updateFilter({ entity: e.target.value || undefined })}
          />
        </div>
        <div className="mt-4 border-t border-gray-100 pt-4 dark:border-ink-border">
          <PeriodFilter
            value={{ from: filters.from, to: filters.to }}
            onChange={(range) => updateFilter({ from: range.from, to: range.to })}
          />
        </div>
      </Card>

      <Card className="p-0">
        {isLoading ? (
          <Table>
            <THead>
              <Th>Quando</Th>
              <Th>Usuário</Th>
              <Th>Ação</Th>
              <Th>Entidade</Th>
              <Th>ID</Th>
              <Th>IP</Th>
            </THead>
            <TBody>
              <SkeletonRows rows={10} cols={6} />
            </TBody>
          </Table>
        ) : logs.length === 0 ? (
          <div className="p-6">
            <EmptyState
              icon={ScrollText}
              title="Nenhum registro de auditoria"
              description="Ajuste os filtros ou aguarde novas atividades na organização."
            />
          </div>
        ) : (
          <Table>
            <THead>
              <Th>Quando</Th>
              <Th>Usuário</Th>
              <Th>Ação</Th>
              <Th>Entidade</Th>
              <Th>ID</Th>
              <Th>IP</Th>
            </THead>
            <TBody>
              {logs.map((log) => (
                <Tr key={log.id}>
                  <Td className="whitespace-nowrap text-gray-500 dark:text-gray-400">
                    {formatDate(log.created_at, "dd MMM yyyy 'às' HH:mm")}
                  </Td>
                  <Td>
                    {log.user ? (
                      <div className="min-w-0">
                        <p className="truncate font-medium text-gray-900 dark:text-gray-100">
                          {log.user.name}
                        </p>
                        <p className="truncate text-xs text-gray-400">{log.user.email}</p>
                      </div>
                    ) : (
                      <span className="text-gray-400">Sistema</span>
                    )}
                  </Td>
                  <Td>
                    <ActionBadge action={log.action} />
                  </Td>
                  <Td className="text-gray-600 dark:text-gray-300">{humanizeEntity(log.entity)}</Td>
                  <Td>
                    {log.entity_id ? (
                      <code
                        className="font-mono text-xs text-gray-500 dark:text-gray-400"
                        title={log.entity_id}
                      >
                        {log.entity_id.slice(0, 8)}…
                      </code>
                    ) : (
                      <span className="text-gray-300 dark:text-gray-600">—</span>
                    )}
                  </Td>
                  <Td className="whitespace-nowrap text-gray-500 dark:text-gray-400">
                    {log.ip || <span className="text-gray-300 dark:text-gray-600">—</span>}
                  </Td>
                </Tr>
              ))}
            </TBody>
          </Table>
        )}

        {/* Pagination */}
        {logs.length > 0 && (
          <div className="flex items-center justify-between border-t border-gray-100 px-4 py-3 dark:border-ink-border">
            <p className="text-xs text-gray-500 dark:text-gray-400">
              {total} registro{total === 1 ? '' : 's'} · página {page} de {pages}
              {isFetching && <span className="ml-2 text-primary">atualizando…</span>}
            </p>
            <div className="flex items-center gap-1.5">
              <Button
                variant="outline"
                size="sm"
                disabled={page <= 1 || isFetching}
                onClick={() => setFilters((p) => ({ ...p, page: page - 1 }))}
              >
                <ChevronLeft className="h-4 w-4" /> Anterior
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={page >= pages || isFetching}
                onClick={() => setFilters((p) => ({ ...p, page: page + 1 }))}
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
