import { useMemo, useState } from 'react';
import {
  Filter,
  Plus,
  ChevronLeft,
  ChevronRight,
  Wallet,
  AlertTriangle,
  CalendarClock,
  ListChecks,
  CheckCircle2,
  Undo2,
  Receipt,
} from 'lucide-react';
import { PageHeader } from '@/components/PageHeader';
import {
  Button,
  Card,
  Select,
  StatCard,
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
import { BillStatusBadge } from './BillStatusBadge';
import { Money } from '@/components/Money';
import { SettleModal } from '@/components/forms/SettleModal';
import { TransactionFormModal } from '@/components/forms/TransactionFormModal';
import {
  usePayables,
  useReceivables,
  useContacts,
  useUnsettleTransaction,
  useIsViewer,
} from '@/hooks';
import { formatDateShort } from '@/lib/date';
import { cn } from '@/lib/cn';
import type { BillFilters, BillStatusFilter, Transaction } from '@/types';

const PER_PAGE = 15;

const statusOptions: { value: BillStatusFilter; label: string }[] = [
  { value: 'open', label: 'Em aberto' },
  { value: 'overdue', label: 'Vencidas' },
  { value: 'paid', label: 'Pagas' },
  { value: 'all', label: 'Todas' },
];

interface Props {
  kind: 'payable' | 'receivable';
}

export function BillsView({ kind }: Props) {
  const isPayable = kind === 'payable';
  const [filters, setFilters] = useState<BillFilters>({
    status: 'open',
    page: 1,
    per_page: PER_PAGE,
  });
  const [showFilters, setShowFilters] = useState(false);
  const [settling, setSettling] = useState<Transaction | null>(null);
  const [createOpen, setCreateOpen] = useState(false);

  const payables = usePayables(filters, isPayable);
  const receivables = useReceivables(filters, !isPayable);
  const query = isPayable ? payables : receivables;
  const { data, isLoading, isFetching } = query;

  const { data: contacts } = useContacts();
  const unsettle = useUnsettleTransaction();
  const readOnly = useIsViewer();

  const contactFilterOptions = useMemo(
    () => [
      { value: '', label: 'Todos os contatos' },
      ...(contacts ?? []).map((c) => ({ value: c.id, label: c.name })),
    ],
    [contacts],
  );

  const updateFilter = (patch: Partial<BillFilters>) =>
    setFilters((prev) => ({ ...prev, ...patch, page: 1 }));

  const handleUnsettle = (t: Transaction) => {
    if (window.confirm(`Desfazer a baixa de "${t.description}"?`)) {
      unsettle.mutate(t.id);
    }
  };

  const summary = data?.summary;
  const meta = data?.meta;
  const page = meta?.page ?? 1;
  const pages = meta?.pages ?? 1;
  const total = meta?.total ?? 0;
  const bills = data?.data ?? [];

  const labels = isPayable
    ? {
        eyebrow: 'A pagar',
        title: 'Contas a pagar',
        description: 'Acompanhe e quite seus compromissos com fornecedores.',
        emptyTitle: 'Nenhuma conta a pagar',
        emptyDesc: 'Não há contas a pagar para os filtros selecionados.',
        action: 'Pagar',
        settleMode: 'pay' as const,
      }
    : {
        eyebrow: 'A receber',
        title: 'Contas a receber',
        description: 'Acompanhe e baixe os recebimentos dos seus clientes.',
        emptyTitle: 'Nenhuma conta a receber',
        emptyDesc: 'Não há contas a receber para os filtros selecionados.',
        action: 'Receber',
        settleMode: 'receive' as const,
      };

  const cols = readOnly ? 5 : 6;

  const createLabel = isPayable ? 'Nova conta a pagar' : 'Nova conta a receber';

  return (
    <div>
      <PageHeader
        eyebrow={labels.eyebrow}
        title={labels.title}
        description={labels.description}
        action={
          !readOnly && (
            <Button onClick={() => setCreateOpen(true)}>
              <Plus className="h-4 w-4" /> {createLabel}
            </Button>
          )
        }
      />

      {/* Summary cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard
          label="Em aberto"
          value={<Money value={summary?.total_open ?? 0} />}
          icon={Wallet}
          tone="primary"
        />
        <StatCard
          label="Vencidas"
          value={<Money value={summary?.total_overdue ?? 0} />}
          icon={AlertTriangle}
          tone="red"
        />
        <StatCard
          label="Vencem em 7 dias"
          value={<Money value={summary?.due_next_7d ?? 0} />}
          icon={CalendarClock}
          tone="amber"
        />
        <StatCard
          label="Contas em aberto"
          value={String(summary?.count_open ?? 0)}
          icon={ListChecks}
          tone="sky"
        />
      </div>

      {/* Filters */}
      <Card className="my-4 p-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
          <Select
            className="w-auto"
            options={statusOptions}
            value={filters.status ?? 'open'}
            onChange={(e) => updateFilter({ status: e.target.value as BillStatusFilter })}
          />
          <Button
            variant="outline"
            size="md"
            onClick={() => setShowFilters((v) => !v)}
            className={cn(showFilters && 'border-primary text-primary')}
          >
            <Filter className="h-4 w-4" /> Filtros
          </Button>
        </div>

        {showFilters && (
          <div className="mt-4 space-y-4 border-t border-gray-100 pt-4 dark:border-ink-border">
            <Select
              label="Contato"
              className="sm:max-w-xs"
              options={contactFilterOptions}
              value={filters.contact_id ?? ''}
              onChange={(e) => updateFilter({ contact_id: e.target.value || undefined })}
            />
            <PeriodFilter
              value={{ from: filters.from, to: filters.to }}
              onChange={(range) => updateFilter({ from: range.from, to: range.to })}
            />
          </div>
        )}
      </Card>

      <Card className="p-0">
        {isLoading ? (
          <Table>
            <THead>
              <Th>Descrição</Th>
              <Th>Contato</Th>
              <Th>Vencimento</Th>
              <Th className="text-right">Valor</Th>
              <Th>Status</Th>
              {!readOnly && <Th className="w-28" />}
            </THead>
            <TBody>
              <SkeletonRows rows={8} cols={cols} />
            </TBody>
          </Table>
        ) : bills.length === 0 ? (
          <div className="p-6">
            <EmptyState
              icon={Receipt}
              title={labels.emptyTitle}
              description={labels.emptyDesc}
              action={
                !readOnly && (
                  <Button onClick={() => setCreateOpen(true)}>
                    <Plus className="h-4 w-4" /> {createLabel}
                  </Button>
                )
              }
            />
          </div>
        ) : (
          <Table>
            <THead>
              <Th>Descrição</Th>
              <Th>Contato</Th>
              <Th>Vencimento</Th>
              <Th className="text-right">Valor</Th>
              <Th>Status</Th>
              {!readOnly && <Th className="w-28 text-right">Ações</Th>}
            </THead>
            <TBody>
              {bills.map((t) => {
                const isOverdue = t.status === 'overdue';
                const isPaid = t.status === 'paid';
                return (
                  <Tr key={t.id}>
                    <Td>
                      <p className="truncate font-medium text-gray-900 dark:text-gray-100">
                        {t.description}
                      </p>
                      {t.category && (
                        <span className="mt-0.5 inline-flex items-center gap-1.5 text-xs text-gray-400">
                          <span
                            className="h-2 w-2 rounded-full"
                            style={{ backgroundColor: t.category.color }}
                          />
                          {t.category.name}
                        </span>
                      )}
                    </Td>
                    <Td className="text-gray-500 dark:text-gray-400">{t.contact?.name ?? '—'}</Td>
                    <Td className="whitespace-nowrap">
                      {t.due_date ? (
                        <span
                          className={cn(
                            isOverdue ? 'font-medium text-red-500' : 'text-gray-500 dark:text-gray-400',
                          )}
                        >
                          {formatDateShort(t.due_date)}
                        </span>
                      ) : (
                        <span className="text-gray-400">—</span>
                      )}
                    </Td>
                    <Td className="text-right">
                      <Money
                        value={t.amount}
                        className="font-semibold text-gray-900 dark:text-gray-100"
                      />
                      {isPaid && t.paid_at && (
                        <p className="mt-0.5 text-xs text-gray-400">
                          Baixada {formatDateShort(t.paid_at)}
                        </p>
                      )}
                    </Td>
                    <Td>
                      <BillStatusBadge status={t.status} />
                    </Td>
                    {!readOnly && (
                      <Td className="text-right">
                        {isPaid ? (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleUnsettle(t)}
                            loading={unsettle.isPending && unsettle.variables === t.id}
                          >
                            <Undo2 className="h-4 w-4" /> Desfazer
                          </Button>
                        ) : (
                          <Button variant="outline" size="sm" onClick={() => setSettling(t)}>
                            <CheckCircle2 className="h-4 w-4" /> {labels.action}
                          </Button>
                        )}
                      </Td>
                    )}
                  </Tr>
                );
              })}
            </TBody>
          </Table>
        )}

        {bills.length > 0 && (
          <div className="flex items-center justify-between border-t border-gray-100 px-4 py-3 dark:border-ink-border">
            <p className="text-xs text-gray-500 dark:text-gray-400">
              {total} conta{total === 1 ? '' : 's'} · página {page} de {pages}
              {isFetching && <span className="ml-2 text-primary">atualizando…</span>}
            </p>
            <div className="flex items-center gap-1.5">
              <Button
                variant="outline"
                size="sm"
                disabled={page <= 1}
                onClick={() => setFilters((p) => ({ ...p, page: page - 1 }))}
              >
                <ChevronLeft className="h-4 w-4" /> Anterior
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={page >= pages}
                onClick={() => setFilters((p) => ({ ...p, page: page + 1 }))}
              >
                Próxima <ChevronRight className="h-4 w-4" />
              </Button>
            </div>
          </div>
        )}
      </Card>

      <SettleModal
        open={Boolean(settling)}
        onClose={() => setSettling(null)}
        transaction={settling}
        mode={labels.settleMode}
      />

      <TransactionFormModal open={createOpen} onClose={() => setCreateOpen(false)} />
    </div>
  );
}
