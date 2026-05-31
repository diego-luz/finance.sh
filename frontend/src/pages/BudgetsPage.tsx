import { useState } from 'react';
import { Plus, PiggyBank, Pencil, Trash2, ChevronLeft, ChevronRight } from 'lucide-react';
import { PageHeader } from '@/components/PageHeader';
import { Button, Card, EmptyState, Skeleton, Select } from '@/components/ui';
import { BudgetFormModal } from '@/components/forms/BudgetFormModal';
import { Money } from '@/components/Money';
import { useBudgets, useDeleteBudget, useIsViewer } from '@/hooks';
import { resolveIcon } from '@/lib/icons';
import { cn } from '@/lib/cn';
import type { Budget } from '@/types';

const MONTHS = [
  'Janeiro', 'Fevereiro', 'Março', 'Abril', 'Maio', 'Junho',
  'Julho', 'Agosto', 'Setembro', 'Outubro', 'Novembro', 'Dezembro',
];

function BudgetRow({
  budget,
  readOnly,
  onEdit,
  onDelete,
}: {
  budget: Budget;
  readOnly: boolean;
  onEdit: () => void;
  onDelete: () => void;
}) {
  const Icon = resolveIcon(budget.category?.icon);
  const over = budget.percent > 100;
  const width = Math.min(100, Math.max(0, budget.percent));
  const color = budget.category?.color || '#10b981';

  return (
    <div className="group flex items-center gap-4 rounded-xl border border-gray-100 bg-white p-4 transition hover:border-gray-200 dark:border-ink-border dark:bg-ink-surface dark:hover:border-gray-600">
      <span
        className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg"
        style={{ backgroundColor: `${color}1a`, color }}
      >
        <Icon className="h-5 w-5" />
      </span>

      <div className="min-w-0 flex-1">
        <div className="flex items-center justify-between gap-3">
          <p className="truncate text-sm font-medium text-gray-900 dark:text-gray-100">
            {budget.category?.name ?? 'Categoria'}
          </p>
          <p className="shrink-0 text-sm tabular-nums text-gray-600 dark:text-gray-300">
            <Money
              value={budget.spent}
              className={cn('font-semibold', over ? 'text-red-500' : 'text-gray-900 dark:text-gray-100')}
            />
            <span className="text-gray-400">
              {' '}/ <Money value={budget.amount} />
            </span>
          </p>
        </div>
        <div className="mt-2 flex items-center gap-3">
          <div className="h-2 flex-1 overflow-hidden rounded-full bg-gray-100 dark:bg-ink-elevated">
            <div
              className="h-full rounded-full transition-all"
              style={{ width: `${width}%`, backgroundColor: over ? '#ef4444' : color }}
            />
          </div>
          <span
            className={cn(
              'w-12 shrink-0 text-right text-xs font-medium tabular-nums',
              over ? 'text-red-500' : 'text-gray-500 dark:text-gray-400',
            )}
          >
            {Math.round(budget.percent)}%
          </span>
        </div>
      </div>

      {!readOnly && (
        <div className="flex shrink-0 items-center gap-0.5 opacity-0 transition group-hover:opacity-100">
          <button
            onClick={onEdit}
            className="rounded-md p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-ink-elevated"
            aria-label="Editar orçamento"
          >
            <Pencil className="h-4 w-4" />
          </button>
          <button
            onClick={onDelete}
            className="rounded-md p-1.5 text-gray-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-500/10"
            aria-label="Excluir orçamento"
          >
            <Trash2 className="h-4 w-4" />
          </button>
        </div>
      )}
    </div>
  );
}

export function BudgetsPage() {
  const now = new Date();
  const [month, setMonth] = useState(now.getMonth() + 1);
  const [year, setYear] = useState(now.getFullYear());
  const readOnly = useIsViewer();

  const { data: budgets, isLoading, isFetching } = useBudgets({ month, year });
  const deleteBudget = useDeleteBudget();

  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Budget | null>(null);

  const openCreate = () => {
    setEditing(null);
    setModalOpen(true);
  };
  const openEdit = (b: Budget) => {
    setEditing(b);
    setModalOpen(true);
  };
  const handleDelete = (b: Budget) => {
    if (window.confirm(`Excluir o orçamento de "${b.category?.name}"?`)) {
      deleteBudget.mutate(b.id);
    }
  };

  const prevMonth = () => {
    if (month === 1) {
      setMonth(12);
      setYear((y) => y - 1);
    } else setMonth((m) => m - 1);
  };
  const nextMonth = () => {
    if (month === 12) {
      setMonth(1);
      setYear((y) => y + 1);
    } else setMonth((m) => m + 1);
  };

  const yearOptions = Array.from({ length: 7 }, (_, i) => {
    const y = now.getFullYear() - 3 + i;
    return { value: String(y), label: String(y) };
  });
  const monthOptions = MONTHS.map((label, i) => ({ value: String(i + 1), label }));

  const list = budgets ?? [];
  const totalBudget = list.reduce((s, b) => s + b.amount, 0);
  const totalSpent = list.reduce((s, b) => s + b.spent, 0);

  return (
    <div>
      <PageHeader
        eyebrow="Planejamento"
        title="Orçamentos"
        description="Defina limites de gastos por categoria e acompanhe o consumo no mês."
        action={
          !readOnly && (
            <Button onClick={openCreate}>
              <Plus className="h-4 w-4" /> Novo orçamento
            </Button>
          )
        }
      />

      <Card className="mb-4 p-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-2">
            <Button variant="outline" size="icon" onClick={prevMonth} aria-label="Mês anterior">
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <Select
              className="w-40"
              options={monthOptions}
              value={String(month)}
              onChange={(e) => setMonth(Number(e.target.value))}
            />
            <Select
              className="w-28"
              options={yearOptions}
              value={String(year)}
              onChange={(e) => setYear(Number(e.target.value))}
            />
            <Button variant="outline" size="icon" onClick={nextMonth} aria-label="Próximo mês">
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
          {list.length > 0 && (
            <p className="text-sm text-gray-500 dark:text-gray-400">
              Total gasto{' '}
              <Money
                value={totalSpent}
                className="font-semibold text-gray-900 dark:text-gray-100"
              />{' '}
              de{' '}
              <Money
                value={totalBudget}
                className="font-semibold text-gray-900 dark:text-gray-100"
              />
              {isFetching && <span className="ml-2 text-primary">atualizando…</span>}
            </p>
          )}
        </div>
      </Card>

      {isLoading ? (
        <div className="space-y-3">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-[72px] rounded-xl" />
          ))}
        </div>
      ) : list.length === 0 ? (
        <EmptyState
          icon={PiggyBank}
          title="Nenhum orçamento neste período"
          description="Crie orçamentos por categoria para controlar seus gastos do mês."
          action={
            !readOnly && (
              <Button onClick={openCreate}>
                <Plus className="h-4 w-4" /> Criar orçamento
              </Button>
            )
          }
        />
      ) : (
        <div className="space-y-3">
          {list.map((b) => (
            <BudgetRow
              key={b.id}
              budget={b}
              readOnly={readOnly}
              onEdit={() => openEdit(b)}
              onDelete={() => handleDelete(b)}
            />
          ))}
        </div>
      )}

      <BudgetFormModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        budget={editing}
        month={month}
        year={year}
      />
    </div>
  );
}
