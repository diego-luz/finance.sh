import { useState } from 'react';
import { Plus, Wand2, Sparkles, Pencil, Trash2 } from 'lucide-react';
import { PageHeader } from '@/components/PageHeader';
import {
  Badge,
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
import { RuleFormModal } from '@/components/forms/RuleFormModal';
import {
  useApplyCategorization,
  useCategoryRules,
  useDeleteRule,
  useIsViewer,
  useUpdateRule,
} from '@/hooks';
import { cn } from '@/lib/cn';
import type { BadgeVariant } from '@/components/ui';
import type { CategoryRule, MatchType } from '@/types';

const matchTypeMeta: Record<MatchType, { label: string; variant: BadgeVariant }> = {
  contains: { label: 'Contém', variant: 'info' },
  prefix: { label: 'Começa com', variant: 'warning' },
  regex: { label: 'Regex', variant: 'neutral' },
};

function RuleRow({
  rule,
  readOnly,
  onEdit,
  onDelete,
  onToggle,
  toggling,
}: {
  rule: CategoryRule;
  readOnly: boolean;
  onEdit: () => void;
  onDelete: () => void;
  onToggle: () => void;
  toggling: boolean;
}) {
  const match = matchTypeMeta[rule.match_type];
  return (
    <Tr>
      <Td className="font-medium text-gray-900 dark:text-gray-100">{rule.pattern}</Td>
      <Td>
        <Badge variant={match.variant}>{match.label}</Badge>
      </Td>
      <Td>
        <span className="flex items-center gap-2">
          <span
            className="h-2.5 w-2.5 shrink-0 rounded-full"
            style={{ backgroundColor: rule.category?.color || '#10b981' }}
          />
          <span className="truncate">{rule.category?.name ?? '—'}</span>
        </span>
      </Td>
      <Td className="tabular-nums">{rule.priority}</Td>
      <Td>
        <button
          type="button"
          role="switch"
          aria-checked={rule.active}
          aria-label={rule.active ? 'Desativar regra' : 'Ativar regra'}
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
              onClick={onEdit}
              className="rounded-md p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-ink-elevated"
              aria-label="Editar regra"
            >
              <Pencil className="h-4 w-4" />
            </button>
            <button
              onClick={onDelete}
              className="rounded-md p-1.5 text-gray-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-500/10"
              aria-label="Excluir regra"
            >
              <Trash2 className="h-4 w-4" />
            </button>
          </div>
        </Td>
      )}
    </Tr>
  );
}

export function CategorizationPage() {
  const { data: rules, isLoading } = useCategoryRules();
  const apply = useApplyCategorization();
  const updateRule = useUpdateRule();
  const deleteRule = useDeleteRule();
  const readOnly = useIsViewer();

  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<CategoryRule | null>(null);

  const list = rules ?? [];
  const cols = readOnly ? 5 : 6;

  const openCreate = () => {
    setEditing(null);
    setModalOpen(true);
  };
  const openEdit = (rule: CategoryRule) => {
    setEditing(rule);
    setModalOpen(true);
  };
  const handleDelete = (rule: CategoryRule) => {
    if (window.confirm(`Excluir a regra "${rule.pattern}"?`)) {
      deleteRule.mutate(rule.id);
    }
  };
  const handleToggle = (rule: CategoryRule) => {
    updateRule.mutate({
      id: rule.id,
      payload: {
        pattern: rule.pattern,
        category_id: rule.category_id,
        match_type: rule.match_type,
        priority: rule.priority,
        active: !rule.active,
      },
    });
  };

  return (
    <div>
      <PageHeader
        eyebrow="Automação"
        title="Categorização automática"
        description="Crie regras por palavra-chave para classificar suas transações sem esforço."
        action={
          !readOnly && (
            <div className="flex flex-wrap items-center gap-2">
              <Button
                variant="outline"
                onClick={() => apply.mutate()}
                loading={apply.isPending}
                disabled={list.length === 0}
              >
                <Wand2 className="h-4 w-4" /> Aplicar regras às transações sem categoria
              </Button>
              <Button onClick={openCreate}>
                <Plus className="h-4 w-4" /> Nova regra
              </Button>
            </div>
          )
        }
      />

      <Card className="p-0">
        {isLoading ? (
          <Table>
            <THead>
              <Th>Palavra-chave</Th>
              <Th>Correspondência</Th>
              <Th>Categoria</Th>
              <Th>Prioridade</Th>
              <Th>Ativo</Th>
              {!readOnly && <Th className="text-right">Ações</Th>}
            </THead>
            <TBody>
              <SkeletonRows rows={5} cols={cols} />
            </TBody>
          </Table>
        ) : list.length === 0 ? (
          <div className="p-5">
            <EmptyState
              icon={Sparkles}
              title="Nenhuma regra cadastrada"
              description="Defina palavras-chave para categorizar automaticamente as transações importadas e lançadas."
              action={
                !readOnly && (
                  <Button onClick={openCreate}>
                    <Plus className="h-4 w-4" /> Criar primeira regra
                  </Button>
                )
              }
            />
          </div>
        ) : (
          <Table>
            <THead>
              <Th>Palavra-chave</Th>
              <Th>Correspondência</Th>
              <Th>Categoria</Th>
              <Th>Prioridade</Th>
              <Th>Ativo</Th>
              {!readOnly && <Th className="text-right">Ações</Th>}
            </THead>
            <TBody>
              {list.map((rule) => (
                <RuleRow
                  key={rule.id}
                  rule={rule}
                  readOnly={readOnly}
                  onEdit={() => openEdit(rule)}
                  onDelete={() => handleDelete(rule)}
                  onToggle={() => handleToggle(rule)}
                  toggling={updateRule.isPending}
                />
              ))}
            </TBody>
          </Table>
        )}
      </Card>

      <RuleFormModal open={modalOpen} onClose={() => setModalOpen(false)} rule={editing} />
    </div>
  );
}
