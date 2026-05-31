import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Plus, Target, MoreVertical, Pencil, Trash2, PlusCircle, CalendarClock } from 'lucide-react';
import { PageHeader } from '@/components/PageHeader';
import { Button, Card, EmptyState, SkeletonCard, Modal, Input } from '@/components/ui';
import { GoalFormModal } from '@/components/forms/GoalFormModal';
import { Money } from '@/components/Money';
import { useGoals, useDeleteGoal, useUpdateGoal, useIsViewer } from '@/hooks';
import { parseCurrencyToCents } from '@/lib/currency';
import { formatDate } from '@/lib/date';
import type { Goal } from '@/types';

function ProgressRing({ progress, color }: { progress: number; color: string }) {
  const pct = Math.max(0, Math.min(1, progress));
  const size = 88;
  const stroke = 8;
  const radius = (size - stroke) / 2;
  const circumference = 2 * Math.PI * radius;
  const offset = circumference * (1 - pct);

  return (
    <div className="relative" style={{ width: size, height: size }}>
      <svg width={size} height={size} className="-rotate-90">
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          strokeWidth={stroke}
          className="stroke-gray-100 dark:stroke-ink-elevated"
        />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          strokeWidth={stroke}
          stroke={color}
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={offset}
          className="transition-all duration-500"
        />
      </svg>
      <span className="absolute inset-0 flex items-center justify-center font-heading text-sm font-bold tabular-nums text-gray-900 dark:text-gray-100">
        {Math.round(pct * 100)}%
      </span>
    </div>
  );
}

function GoalCard({
  goal,
  readOnly,
  onEdit,
  onDelete,
  onAddValue,
}: {
  goal: Goal;
  readOnly: boolean;
  onEdit: () => void;
  onDelete: () => void;
  onAddValue: () => void;
}) {
  const [menu, setMenu] = useState(false);
  const completed = goal.progress >= 1;

  return (
    <Card className="relative overflow-hidden">
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-4">
          <ProgressRing progress={goal.progress} color={goal.color || '#0ea5e9'} />
          <div className="min-w-0">
            <h3 className="font-heading text-base font-semibold text-gray-900 dark:text-gray-100">
              {goal.name}
            </h3>
            <p className="mt-1 text-sm font-medium tabular-nums text-gray-900 dark:text-gray-100">
              <Money value={goal.current_amount} />
              <span className="text-gray-400">
                {' '}/ <Money value={goal.target_amount} />
              </span>
            </p>
            {goal.deadline && (
              <p className="mt-1 flex items-center gap-1 text-xs text-gray-400">
                <CalendarClock className="h-3.5 w-3.5" /> Prazo {formatDate(goal.deadline)}
              </p>
            )}
          </div>
        </div>

        {!readOnly && (
          <div className="relative">
            <button
              onClick={() => setMenu((v) => !v)}
              onBlur={() => setTimeout(() => setMenu(false), 150)}
              className="rounded-lg p-1.5 text-gray-400 transition hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-ink-elevated"
              aria-label="Ações da meta"
            >
              <MoreVertical className="h-4 w-4" />
            </button>
            {menu && (
              <div className="absolute right-0 z-10 mt-1 w-40 animate-fade-in overflow-hidden rounded-lg border border-gray-200 bg-white p-1 shadow-card dark:border-ink-border dark:bg-ink-surface">
                <button
                  onClick={onEdit}
                  className="flex w-full items-center gap-2 rounded-md px-2.5 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-ink-elevated"
                >
                  <Pencil className="h-4 w-4" /> Editar
                </button>
                <button
                  onClick={onDelete}
                  className="flex w-full items-center gap-2 rounded-md px-2.5 py-2 text-sm text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-500/10"
                >
                  <Trash2 className="h-4 w-4" /> Excluir
                </button>
              </div>
            )}
          </div>
        )}
      </div>

      {!readOnly && (
        <div className="mt-4">
          <Button variant="outline" size="sm" className="w-full" onClick={onAddValue} disabled={completed}>
            <PlusCircle className="h-4 w-4" /> {completed ? 'Meta concluída' : 'Adicionar valor'}
          </Button>
        </div>
      )}
    </Card>
  );
}

const addSchema = z.object({
  amount: z.string().min(1, 'Informe o valor'),
});
type AddValues = z.infer<typeof addSchema>;

function AddValueModal({
  open,
  onClose,
  goal,
}: {
  open: boolean;
  onClose: () => void;
  goal: Goal | null;
}) {
  const update = useUpdateGoal();
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<AddValues>({ resolver: zodResolver(addSchema), defaultValues: { amount: '' } });

  useEffect(() => {
    if (open) reset({ amount: '' });
  }, [open, reset]);

  const onSubmit = (values: AddValues) => {
    if (!goal) return;
    const added = parseCurrencyToCents(values.amount);
    const payload = {
      name: goal.name,
      target_amount: goal.target_amount,
      current_amount: goal.current_amount + added,
      deadline: goal.deadline,
      color: goal.color,
    };
    update.mutate({ id: goal.id, payload }, { onSuccess: onClose });
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      size="sm"
      title="Adicionar valor"
      description={goal ? `Adicione um aporte à meta "${goal.name}".` : undefined}
      footer={
        <>
          <Button variant="ghost" onClick={onClose} type="button">
            Cancelar
          </Button>
          <Button type="submit" form="add-value-form" loading={update.isPending}>
            Adicionar
          </Button>
        </>
      }
    >
      <form id="add-value-form" onSubmit={handleSubmit(onSubmit)} className="space-y-2" noValidate>
        <Input
          label="Valor do aporte"
          placeholder="0,00"
          inputMode="decimal"
          rightAddon="R$"
          autoFocus
          error={errors.amount?.message}
          {...register('amount')}
        />
        {goal && (
          <p className="text-xs text-gray-400">
            Acumulado atual: <Money value={goal.current_amount} />
          </p>
        )}
      </form>
    </Modal>
  );
}

export function GoalsPage() {
  const { data: goals, isLoading } = useGoals();
  const deleteGoal = useDeleteGoal();
  const readOnly = useIsViewer();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Goal | null>(null);
  const [addOpen, setAddOpen] = useState(false);
  const [addGoal, setAddGoal] = useState<Goal | null>(null);

  const openCreate = () => {
    setEditing(null);
    setModalOpen(true);
  };
  const openEdit = (goal: Goal) => {
    setEditing(goal);
    setModalOpen(true);
  };
  const openAdd = (goal: Goal) => {
    setAddGoal(goal);
    setAddOpen(true);
  };
  const handleDelete = (goal: Goal) => {
    if (window.confirm(`Excluir a meta "${goal.name}"?`)) {
      deleteGoal.mutate(goal.id);
    }
  };

  return (
    <div>
      <PageHeader
        eyebrow="Planejamento"
        title="Metas"
        description="Defina objetivos financeiros e acompanhe o progresso de cada um."
        action={
          !readOnly && (
            <Button onClick={openCreate}>
              <Plus className="h-4 w-4" /> Nova meta
            </Button>
          )
        }
      />

      {isLoading ? (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <SkeletonCard key={i} />
          ))}
        </div>
      ) : !goals || goals.length === 0 ? (
        <EmptyState
          icon={Target}
          title="Nenhuma meta cadastrada"
          description="Crie sua primeira meta para acompanhar quanto falta para alcançá-la."
          action={
            !readOnly && (
              <Button onClick={openCreate}>
                <Plus className="h-4 w-4" /> Criar meta
              </Button>
            )
          }
        />
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {goals.map((goal) => (
            <GoalCard
              key={goal.id}
              goal={goal}
              readOnly={readOnly}
              onEdit={() => openEdit(goal)}
              onDelete={() => handleDelete(goal)}
              onAddValue={() => openAdd(goal)}
            />
          ))}
        </div>
      )}

      <GoalFormModal open={modalOpen} onClose={() => setModalOpen(false)} goal={editing} />
      <AddValueModal open={addOpen} onClose={() => setAddOpen(false)} goal={addGoal} />
    </div>
  );
}
