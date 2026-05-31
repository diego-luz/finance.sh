import { useState } from 'react';
import { Plus, Repeat, Play, Pencil, Trash2 } from 'lucide-react';
import { PageHeader } from '@/components/PageHeader';
import {
  Button,
  Card,
  EmptyState,
  Table,
  THead,
  Th,
  TBody,
  Tr,
  Td,
  SkeletonRows,
} from '@/components/ui';
import { RecurrenceFormModal } from '@/components/forms/RecurrenceFormModal';
import {
  useRecurrences,
  useDeleteRecurrence,
  useUpdateRecurrence,
  useRunRecurrence,
  useIsViewer,
} from '@/hooks';
import { cn } from '@/lib/cn';
import { Money } from '@/components/Money';
import { formatDateShort } from '@/lib/date';
import { humanizeFrequency } from '@/lib/recurrence';
import type { RecurrenceRule, RecurrenceRulePayload } from '@/types';

/** Build the full payload from a rule (used by the inline active toggle). */
function toPayload(rule: RecurrenceRule, active: boolean): RecurrenceRulePayload {
  return {
    type: rule.type,
    description: rule.description,
    amount: rule.amount,
    account_id: rule.account_id,
    category_id: rule.category_id,
    contact_id: rule.contact_id,
    frequency: rule.frequency,
    interval: rule.interval,
    start_date: rule.start_date,
    end_date: rule.end_date,
    max_occurrences: rule.max_occurrences,
    paid: rule.paid,
    active,
  };
}

function RecurrenceRow({
  rule,
  readOnly,
  onEdit,
  onDelete,
  onToggle,
  onRun,
  toggling,
  running,
}: {
  rule: RecurrenceRule;
  readOnly: boolean;
  onEdit: () => void;
  onDelete: () => void;
  onToggle: () => void;
  onRun: () => void;
  toggling: boolean;
  running: boolean;
}) {
  const isIncome = rule.type === 'income';
  const counter =
    rule.max_occurrences > 0
      ? `${rule.occurrences_count}/${rule.max_occurrences}`
      : `${rule.occurrences_count}/∞`;

  return (
    <Tr>
      <Td className="font-medium text-gray-900 dark:text-gray-100">{rule.description}</Td>
      <Td
        className={cn(
          'tabular-nums font-medium',
          isIncome ? 'text-primary' : 'text-red-600 dark:text-red-400',
        )}
      >
        {isIncome ? '+' : '−'}
        <Money value={rule.amount} />
      </Td>
      <Td>{humanizeFrequency(rule.frequency, rule.interval)}</Td>
      <Td className="tabular-nums">
        {rule.next_run_date ? formatDateShort(rule.next_run_date) : '—'}
      </Td>
      <Td className="truncate">{rule.account?.name ?? '—'}</Td>
      <Td>
        {rule.category ? (
          <span className="flex items-center gap-2">
            <span
              className="h-2.5 w-2.5 shrink-0 rounded-full"
              style={{ backgroundColor: rule.category.color || '#10b981' }}
            />
            <span className="truncate">{rule.category.name}</span>
          </span>
        ) : (
          '—'
        )}
      </Td>
      <Td className="tabular-nums text-gray-500 dark:text-gray-400">{counter}</Td>
      <Td>
        <button
          type="button"
          role="switch"
          aria-checked={rule.active}
          aria-label={rule.active ? 'Pausar recorrência' : 'Ativar recorrência'}
          disabled={readOnly || toggling}
          onClick={onToggle}
          className={cn(
            'relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors disabled:cursor-not-allowed disabled:opacity-60',
            rule.active ? 'bg-primary' : 'bg-gray-300 dark:bg-ink-border',
          )}
        >
          <span
            className={cn(
              'inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform',
              rule.active ? 'translate-x-4' : 'translate-x-0.5',
            )}
          />
        </button>
      </Td>
      {!readOnly && (
        <Td className="text-right">
          <div className="flex items-center justify-end gap-0.5">
            <button
              onClick={onRun}
              disabled={running}
              className="rounded-md p-1.5 text-gray-400 transition hover:bg-primary/10 hover:text-primary disabled:cursor-not-allowed disabled:opacity-60"
              aria-label="Gerar lançamentos agora"
              title="Gerar agora"
            >
              <Play className="h-4 w-4" />
            </button>
            <button
              onClick={onEdit}
              className="rounded-md p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-ink-elevated"
              aria-label="Editar recorrência"
            >
              <Pencil className="h-4 w-4" />
            </button>
            <button
              onClick={onDelete}
              className="rounded-md p-1.5 text-gray-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-500/10"
              aria-label="Excluir recorrência"
            >
              <Trash2 className="h-4 w-4" />
            </button>
          </div>
        </Td>
      )}
    </Tr>
  );
}

export function RecurrencesPage() {
  const { data: recurrences, isLoading } = useRecurrences();
  const updateRecurrence = useUpdateRecurrence();
  const deleteRecurrence = useDeleteRecurrence();
  const runRecurrence = useRunRecurrence();
  const readOnly = useIsViewer();

  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<RecurrenceRule | null>(null);

  const list = recurrences ?? [];
  const cols = readOnly ? 8 : 9;

  const openCreate = () => {
    setEditing(null);
    setModalOpen(true);
  };
  const openEdit = (rule: RecurrenceRule) => {
    setEditing(rule);
    setModalOpen(true);
  };
  const handleDelete = (rule: RecurrenceRule) => {
    if (window.confirm(`Excluir a recorrência "${rule.description}"?`)) {
      deleteRecurrence.mutate(rule.id);
    }
  };
  const handleToggle = (rule: RecurrenceRule) => {
    updateRecurrence.mutate({ id: rule.id, payload: toPayload(rule, !rule.active) });
  };

  const header = (
    <>
      <Th>Descrição</Th>
      <Th>Valor</Th>
      <Th>Frequência</Th>
      <Th>Próxima</Th>
      <Th>Conta</Th>
      <Th>Categoria</Th>
      <Th>Ocorrências</Th>
      <Th>Ativa</Th>
      {!readOnly && <Th className="text-right">Ações</Th>}
    </>
  );

  return (
    <div>
      <PageHeader
        eyebrow="Automação"
        title="Recorrências"
        description="Agende lançamentos repetidos e gere as transações automaticamente."
        action={
          !readOnly && (
            <Button onClick={openCreate}>
              <Plus className="h-4 w-4" /> Nova recorrência
            </Button>
          )
        }
      />

      <Card className="p-0">
        {isLoading ? (
          <Table>
            <THead>{header}</THead>
            <TBody>
              <SkeletonRows rows={5} cols={cols} />
            </TBody>
          </Table>
        ) : list.length === 0 ? (
          <div className="p-5">
            <EmptyState
              icon={Repeat}
              title="Nenhuma recorrência cadastrada"
              description="Crie recorrências para lançar despesas e receitas fixas automaticamente, como aluguel ou assinaturas."
              action={
                !readOnly && (
                  <Button onClick={openCreate}>
                    <Plus className="h-4 w-4" /> Criar primeira recorrência
                  </Button>
                )
              }
            />
          </div>
        ) : (
          <Table>
            <THead>{header}</THead>
            <TBody>
              {list.map((rule) => (
                <RecurrenceRow
                  key={rule.id}
                  rule={rule}
                  readOnly={readOnly}
                  onEdit={() => openEdit(rule)}
                  onDelete={() => handleDelete(rule)}
                  onToggle={() => handleToggle(rule)}
                  onRun={() => runRecurrence.mutate(rule.id)}
                  toggling={updateRecurrence.isPending}
                  running={runRecurrence.isPending}
                />
              ))}
            </TBody>
          </Table>
        )}
      </Card>

      <RecurrenceFormModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        recurrence={editing}
      />
    </div>
  );
}
