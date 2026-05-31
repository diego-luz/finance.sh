import { useEffect, useMemo, useState, type FormEvent } from 'react';
import {
  Plus,
  Search,
  ArrowLeftRight,
  Pencil,
  Trash2,
  ChevronLeft,
  ChevronRight,
  ArrowUpRight,
  ArrowDownRight,
  Filter,
  CheckCircle2,
  Tag,
  X,
  Paperclip,
  FileUp,
} from 'lucide-react';
import { PageHeader } from '@/components/PageHeader';
import {
  Button,
  Card,
  Input,
  Select,
  Badge,
  EmptyState,
  PeriodFilter,
  Modal,
  Table,
  THead,
  Th,
  TBody,
  Tr,
  Td,
  SkeletonRows,
} from '@/components/ui';
import { ExportMenu } from '@/components/ExportMenu';
import { Money } from '@/components/Money';
import { TransactionFormModal } from '@/components/forms/TransactionFormModal';
import { ImportStatementModal } from '@/components/forms/ImportStatementModal';
import {
  useAccounts,
  useCategories,
  useTags,
  useTransactions,
  useDeleteTransaction,
  useBulkSettle,
  useBulkCategorize,
  useBulkDelete,
  useIsViewer,
} from '@/hooks';
import { formatDateShort, toISODate, todayInput } from '@/lib/date';
import { cn } from '@/lib/cn';
import type { Tag as TagType, Transaction, TransactionFilters, TransactionType } from '@/types';

const PER_PAGE = 15;

const typeOptions = [
  { value: '', label: 'Todos os tipos' },
  { value: 'income', label: 'Receitas' },
  { value: 'expense', label: 'Despesas' },
  { value: 'transfer', label: 'Transferências' },
];

function TypeBadge({ type }: { type: TransactionType }) {
  if (type === 'income') return <Badge variant="success" dot>Receita</Badge>;
  if (type === 'expense') return <Badge variant="danger" dot>Despesa</Badge>;
  return <Badge variant="info" dot>Transferência</Badge>;
}

/** Compact tag chips for a transaction: shows up to 3, then a "+N" pill. */
function TransactionTags({ tags }: { tags?: TagType[] }) {
  if (!tags || tags.length === 0) return null;
  const shown = tags.slice(0, 3);
  const extra = tags.length - shown.length;
  return (
    <>
      {shown.map((tag) => (
        <span
          key={tag.id}
          className="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium"
          style={{ backgroundColor: `${tag.color}1a`, color: tag.color }}
        >
          <span className="h-1.5 w-1.5 rounded-full" style={{ backgroundColor: tag.color }} />
          {tag.name}
        </span>
      ))}
      {extra > 0 && (
        <span className="inline-flex items-center rounded-full bg-gray-500/10 px-2 py-0.5 text-xs font-medium text-gray-500 dark:text-gray-400">
          +{extra}
        </span>
      )}
    </>
  );
}

export function TransactionsPage() {
  const [filters, setFilters] = useState<TransactionFilters>({ page: 1, per_page: PER_PAGE });
  const [searchInput, setSearchInput] = useState('');
  const [showFilters, setShowFilters] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Transaction | null>(null);
  const [importOpen, setImportOpen] = useState(false);

  // Bulk selection
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [settleOpen, setSettleOpen] = useState(false);
  const [categorizeOpen, setCategorizeOpen] = useState(false);

  const { data, isLoading, isFetching } = useTransactions(filters);
  const { data: accounts } = useAccounts();
  const { data: categories } = useCategories();
  const { data: tags } = useTags();
  const deleteTransaction = useDeleteTransaction();
  const bulkSettle = useBulkSettle();
  const bulkCategorize = useBulkCategorize();
  const bulkDelete = useBulkDelete();
  const readOnly = useIsViewer();

  const accountMap = useMemo(
    () => new Map((accounts ?? []).map((a) => [a.id, a])),
    [accounts],
  );

  const accountFilterOptions = useMemo(
    () => [
      { value: '', label: 'Todas as contas' },
      ...(accounts ?? []).map((a) => ({ value: a.id, label: a.name })),
    ],
    [accounts],
  );
  const categoryFilterOptions = useMemo(
    () => [
      { value: '', label: 'Todas as categorias' },
      ...(categories ?? []).map((c) => ({ value: c.id, label: c.name })),
    ],
    [categories],
  );
  const tagFilterOptions = useMemo(
    () => [
      { value: '', label: 'Todas as tags' },
      ...(tags ?? []).map((t) => ({ value: t.id, label: t.name })),
    ],
    [tags],
  );

  const updateFilter = (patch: Partial<TransactionFilters>) =>
    setFilters((prev) => ({ ...prev, ...patch, page: 1 }));

  const onSearchSubmit = (e: FormEvent) => {
    e.preventDefault();
    updateFilter({ search: searchInput || undefined });
  };

  const openCreate = () => {
    setEditing(null);
    setModalOpen(true);
  };
  const openEdit = (t: Transaction) => {
    setEditing(t);
    setModalOpen(true);
  };
  const handleDelete = (t: Transaction) => {
    const total = t.installment_total ?? 0;
    if (total > 1) {
      // Parcela of a group: offer "this one" vs "all installments".
      const all = window.confirm(
        `Esta transação é a parcela ${t.installment_number}/${total} de "${t.description}".\n\n` +
          `Clique em OK para excluir TODAS as ${total} parcelas, ou em Cancelar para escolher excluir só esta.`,
      );
      if (all) {
        deleteTransaction.mutate({ id: t.id, scope: 'all' });
        return;
      }
      if (window.confirm('Excluir somente esta parcela?')) {
        deleteTransaction.mutate({ id: t.id, scope: 'one' });
      }
      return;
    }
    if (window.confirm(`Excluir a transação "${t.description}"?`)) {
      deleteTransaction.mutate({ id: t.id, scope: 'one' });
    }
  };

  const meta = data?.meta;
  const page = meta?.page ?? 1;
  const pages = meta?.pages ?? 1;
  const total = meta?.total ?? 0;
  const transactions = data?.data ?? [];

  // Clear selection when the visible rows change (page / filter changes).
  useEffect(() => {
    setSelected(new Set());
  }, [filters]);

  const pageIds = useMemo(() => transactions.map((t) => t.id), [transactions]);
  const allSelected = pageIds.length > 0 && pageIds.every((id) => selected.has(id));
  const someSelected = selected.size > 0;
  const selectedIds = useMemo(() => Array.from(selected), [selected]);

  const toggleAll = () => {
    setSelected((prev) => {
      if (pageIds.every((id) => prev.has(id))) return new Set();
      return new Set(pageIds);
    });
  };
  const toggleOne = (id: string) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };
  const clearSelection = () => setSelected(new Set());

  const handleBulkDelete = () => {
    if (window.confirm(`Excluir ${selectedIds.length} transação(ões) selecionada(s)?`)) {
      bulkDelete.mutate(selectedIds, { onSuccess: clearSelection });
    }
  };

  // The first non-read-only column is the checkbox; total columns vary by role.
  const showSelect = !readOnly;

  return (
    <div>
      <PageHeader
        eyebrow="Movimentações"
        title="Transações"
        description="Acompanhe e registre todas as entradas e saídas."
        action={
          <div className="flex items-center gap-2">
            <ExportMenu range={{ from: filters.from, to: filters.to }} compact />
            {!readOnly && (
              <Button variant="outline" onClick={() => setImportOpen(true)}>
                <FileUp className="h-4 w-4" /> Importar extrato
              </Button>
            )}
            {!readOnly && (
              <Button onClick={openCreate}>
                <Plus className="h-4 w-4" /> Nova transação
              </Button>
            )}
          </div>
        }
      />

      <Card className="mb-4 p-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
          <form onSubmit={onSearchSubmit} className="flex-1">
            <Input
              placeholder="Buscar por descrição..."
              leftIcon={<Search className="h-4 w-4" />}
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
            />
          </form>
          <div className="flex items-center gap-2">
            <Select
              className="w-auto"
              options={typeOptions}
              value={filters.type ?? ''}
              onChange={(e) => updateFilter({ type: e.target.value as TransactionType | '' })}
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
        </div>

        {showFilters && (
          <div className="mt-4 space-y-4 border-t border-gray-100 pt-4 dark:border-ink-border">
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <Select
                label="Conta"
                options={accountFilterOptions}
                value={filters.account_id ?? ''}
                onChange={(e) => updateFilter({ account_id: e.target.value || undefined })}
              />
              <Select
                label="Categoria"
                options={categoryFilterOptions}
                value={filters.category_id ?? ''}
                onChange={(e) => updateFilter({ category_id: e.target.value || undefined })}
              />
              <Select
                label="Tag"
                options={tagFilterOptions}
                value={filters.tag_id ?? ''}
                onChange={(e) => updateFilter({ tag_id: e.target.value || undefined })}
              />
            </div>
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
              {showSelect && <Th className="w-10" />}
              <Th>Descrição</Th>
              <Th>Categoria</Th>
              <Th>Conta</Th>
              <Th>Data</Th>
              <Th className="text-right">Valor</Th>
              {!readOnly && <Th className="w-20" />}
            </THead>
            <TBody>
              <SkeletonRows rows={8} cols={(showSelect ? 1 : 0) + (readOnly ? 5 : 6)} />
            </TBody>
          </Table>
        ) : transactions.length === 0 ? (
          <div className="p-6">
            <EmptyState
              icon={ArrowLeftRight}
              title="Nenhuma transação encontrada"
              description="Ajuste os filtros ou registre uma nova transação."
              action={
                !readOnly && (
                  <Button onClick={openCreate}>
                    <Plus className="h-4 w-4" /> Adicionar primeira transação
                  </Button>
                )
              }
            />
          </div>
        ) : (
          <Table>
            <THead>
              {showSelect && (
                <Th className="w-10">
                  <input
                    type="checkbox"
                    aria-label="Selecionar todas"
                    checked={allSelected}
                    onChange={toggleAll}
                    className="h-4 w-4 cursor-pointer rounded border-gray-300 text-primary focus:ring-primary/30 dark:border-ink-border dark:bg-ink-elevated"
                  />
                </Th>
              )}
              <Th>Descrição</Th>
              <Th>Categoria</Th>
              <Th>Conta</Th>
              <Th>Data</Th>
              <Th className="text-right">Valor</Th>
              {!readOnly && <Th className="w-20 text-right">Ações</Th>}
            </THead>
            <TBody>
              {transactions.map((t) => {
                const account = accountMap.get(t.account_id);
                const isIncome = t.type === 'income';
                const isChecked = selected.has(t.id);
                return (
                  <Tr key={t.id} className={cn(isChecked && 'bg-primary/5 dark:bg-primary/10')}>
                    {showSelect && (
                      <Td className="w-10">
                        <input
                          type="checkbox"
                          aria-label={`Selecionar ${t.description}`}
                          checked={isChecked}
                          onChange={() => toggleOne(t.id)}
                          className="h-4 w-4 cursor-pointer rounded border-gray-300 text-primary focus:ring-primary/30 dark:border-ink-border dark:bg-ink-elevated"
                        />
                      </Td>
                    )}
                    <Td>
                      <div className="flex items-center gap-3">
                        <span
                          className={cn(
                            'flex h-8 w-8 shrink-0 items-center justify-center rounded-lg',
                            isIncome
                              ? 'bg-primary/10 text-primary'
                              : t.type === 'transfer'
                                ? 'bg-sky-500/10 text-sky-500'
                                : 'bg-red-500/10 text-red-500',
                          )}
                        >
                          {isIncome ? (
                            <ArrowUpRight className="h-4 w-4" />
                          ) : (
                            <ArrowDownRight className="h-4 w-4" />
                          )}
                        </span>
                        <div className="min-w-0">
                          <div className="flex items-center gap-1.5">
                            <p className="truncate font-medium text-gray-900 dark:text-gray-100">
                              {t.description}
                            </p>
                            {(t.attachment_count ?? 0) > 0 && (
                              <span
                                className="inline-flex shrink-0 items-center gap-0.5 text-gray-400"
                                title={`${t.attachment_count} comprovante(s) anexado(s)`}
                                aria-label={`${t.attachment_count} comprovante(s) anexado(s)`}
                              >
                                <Paperclip className="h-3.5 w-3.5" />
                                {(t.attachment_count ?? 0) > 1 && (
                                  <span className="text-xs tabular-nums">{t.attachment_count}</span>
                                )}
                              </span>
                            )}
                          </div>
                          <div className="mt-0.5 flex flex-wrap items-center gap-1.5">
                            <TypeBadge type={t.type} />
                            {(t.installment_total ?? 0) > 1 && (
                              <Badge variant="info">
                                {t.installment_number}/{t.installment_total}
                              </Badge>
                            )}
                            {!t.paid && <Badge variant="warning">Pendente</Badge>}
                            {t.recurring && <Badge variant="neutral">Recorrente</Badge>}
                            <TransactionTags tags={t.tags} />
                          </div>
                        </div>
                      </div>
                    </Td>
                    <Td>
                      {t.category ? (
                        <span className="inline-flex items-center gap-1.5">
                          <span
                            className="h-2.5 w-2.5 rounded-full"
                            style={{ backgroundColor: t.category.color }}
                          />
                          {t.category.name}
                        </span>
                      ) : (
                        <span className="text-gray-400">—</span>
                      )}
                    </Td>
                    <Td className="text-gray-500 dark:text-gray-400">{account?.name ?? '—'}</Td>
                    <Td className="whitespace-nowrap text-gray-500 dark:text-gray-400">
                      {formatDateShort(t.date)}
                    </Td>
                    <Td className="text-right">
                      <span
                        className={cn(
                          'font-semibold tabular-nums',
                          isIncome ? 'text-primary' : 'text-red-500',
                        )}
                      >
                        {isIncome ? '+' : '−'} <Money value={t.amount} />
                      </span>
                    </Td>
                    {!readOnly && (
                      <Td className="text-right">
                        <div className="flex items-center justify-end gap-0.5">
                          <button
                            onClick={() => openEdit(t)}
                            className="rounded-md p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-ink-elevated"
                            aria-label="Editar"
                          >
                            <Pencil className="h-4 w-4" />
                          </button>
                          <button
                            onClick={() => handleDelete(t)}
                            className="rounded-md p-1.5 text-gray-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-500/10"
                            aria-label="Excluir"
                          >
                            <Trash2 className="h-4 w-4" />
                          </button>
                        </div>
                      </Td>
                    )}
                  </Tr>
                );
              })}
            </TBody>
          </Table>
        )}

        {/* Pagination */}
        {transactions.length > 0 && (
          <div className="flex items-center justify-between border-t border-gray-100 px-4 py-3 dark:border-ink-border">
            <p className="text-xs text-gray-500 dark:text-gray-400">
              {total} transaç{total === 1 ? 'ão' : 'ões'} · página {page} de {pages}
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

      {/* Sticky bulk action bar */}
      {!readOnly && someSelected && (
        <div className="sticky bottom-4 z-30 mt-4">
          <div className="mx-auto flex max-w-3xl flex-col gap-3 rounded-xl border border-gray-200 bg-white/95 px-4 py-3 shadow-card backdrop-blur dark:border-ink-border dark:bg-ink-surface/95 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex items-center gap-2">
              <span className="flex h-7 min-w-7 items-center justify-center rounded-full bg-primary/10 px-2 text-xs font-semibold text-primary">
                {selectedIds.length}
              </span>
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                selecionada{selectedIds.length === 1 ? '' : 's'}
              </span>
              <button
                onClick={clearSelection}
                className="ml-1 inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-xs text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-ink-elevated"
              >
                <X className="h-3.5 w-3.5" /> Limpar
              </button>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setSettleOpen(true)}
                loading={bulkSettle.isPending}
              >
                <CheckCircle2 className="h-4 w-4" /> Marcar como pago
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setCategorizeOpen(true)}
                loading={bulkCategorize.isPending}
              >
                <Tag className="h-4 w-4" /> Categorizar
              </Button>
              <Button
                variant="danger"
                size="sm"
                onClick={handleBulkDelete}
                loading={bulkDelete.isPending}
              >
                <Trash2 className="h-4 w-4" /> Excluir
              </Button>
            </div>
          </div>
        </div>
      )}

      <TransactionFormModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        transaction={editing}
      />

      {!readOnly && (
        <ImportStatementModal
          open={importOpen}
          onClose={() => setImportOpen(false)}
          defaultAccountId={filters.account_id}
        />
      )}

      <BulkSettleModal
        open={settleOpen}
        onClose={() => setSettleOpen(false)}
        count={selectedIds.length}
        loading={bulkSettle.isPending}
        accountOptions={accounts ?? []}
        onConfirm={(accountId, paidAt) =>
          bulkSettle.mutate(
            {
              ids: selectedIds,
              account_id: accountId || undefined,
              paid_at: paidAt ? toISODate(paidAt) : undefined,
            },
            {
              onSuccess: () => {
                setSettleOpen(false);
                clearSelection();
              },
            },
          )
        }
      />

      <BulkCategorizeModal
        open={categorizeOpen}
        onClose={() => setCategorizeOpen(false)}
        count={selectedIds.length}
        loading={bulkCategorize.isPending}
        categoryOptions={(categories ?? []).map((c) => ({ value: c.id, label: c.name }))}
        onConfirm={(categoryId) =>
          bulkCategorize.mutate(
            { ids: selectedIds, category_id: categoryId },
            {
              onSuccess: () => {
                setCategorizeOpen(false);
                clearSelection();
              },
            },
          )
        }
      />
    </div>
  );
}

function BulkSettleModal({
  open,
  onClose,
  count,
  loading,
  accountOptions,
  onConfirm,
}: {
  open: boolean;
  onClose: () => void;
  count: number;
  loading: boolean;
  accountOptions: { id: string; name: string }[];
  onConfirm: (accountId: string, paidAt: string) => void;
}) {
  const [accountId, setAccountId] = useState('');
  const [paidAt, setPaidAt] = useState(todayInput());

  useEffect(() => {
    if (open) {
      setAccountId('');
      setPaidAt(todayInput());
    }
  }, [open]);

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Marcar como pago"
      description={`${count} transação(ões) serão baixadas.`}
      size="sm"
      footer={
        <>
          <Button variant="ghost" type="button" onClick={onClose} disabled={loading}>
            Cancelar
          </Button>
          <Button type="button" loading={loading} onClick={() => onConfirm(accountId, paidAt)}>
            <CheckCircle2 className="h-4 w-4" /> Confirmar baixa
          </Button>
        </>
      }
    >
      <div className="space-y-4">
        <Select
          label="Conta (opcional)"
          placeholder="Manter a conta de cada transação"
          options={accountOptions.map((a) => ({ value: a.id, label: a.name }))}
          value={accountId}
          onChange={(e) => setAccountId(e.target.value)}
        />
        <Input
          label="Data do pagamento"
          type="date"
          value={paidAt}
          onChange={(e) => setPaidAt(e.target.value)}
        />
      </div>
    </Modal>
  );
}

function BulkCategorizeModal({
  open,
  onClose,
  count,
  loading,
  categoryOptions,
  onConfirm,
}: {
  open: boolean;
  onClose: () => void;
  count: number;
  loading: boolean;
  categoryOptions: { value: string; label: string }[];
  onConfirm: (categoryId: string) => void;
}) {
  const [categoryId, setCategoryId] = useState('');

  useEffect(() => {
    if (open) setCategoryId('');
  }, [open]);

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Categorizar transações"
      description={`Aplicar uma categoria a ${count} transação(ões).`}
      size="sm"
      footer={
        <>
          <Button variant="ghost" type="button" onClick={onClose} disabled={loading}>
            Cancelar
          </Button>
          <Button
            type="button"
            loading={loading}
            disabled={!categoryId}
            onClick={() => categoryId && onConfirm(categoryId)}
          >
            <Tag className="h-4 w-4" /> Aplicar categoria
          </Button>
        </>
      }
    >
      <Select
        label="Categoria"
        placeholder="Selecione uma categoria"
        options={categoryOptions}
        value={categoryId}
        onChange={(e) => setCategoryId(e.target.value)}
      />
    </Modal>
  );
}
