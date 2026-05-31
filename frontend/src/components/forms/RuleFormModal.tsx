import { useEffect } from 'react';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Modal, Input, Select, Button } from '@/components/ui';
import { cn } from '@/lib/cn';
import { useCategories, useCreateRule, useUpdateRule } from '@/hooks';
import type { CategoryRule, MatchType } from '@/types';

const schema = z.object({
  pattern: z.string().min(1, 'Informe a palavra-chave'),
  category_id: z.string().min(1, 'Selecione a categoria'),
  match_type: z.enum(['contains', 'prefix', 'regex']),
  priority: z.coerce.number().int().min(0).max(9999),
  active: z.boolean(),
});

type FormValues = z.infer<typeof schema>;

const matchTypeOptions: { value: MatchType; label: string }[] = [
  { value: 'contains', label: 'Contém' },
  { value: 'prefix', label: 'Começa com' },
  { value: 'regex', label: 'Expressão regular' },
];

interface Props {
  open: boolean;
  onClose: () => void;
  rule?: CategoryRule | null;
}

export function RuleFormModal({ open, onClose, rule }: Props) {
  const { data: categories } = useCategories();
  const create = useCreateRule();
  const update = useUpdateRule();
  const isEdit = Boolean(rule);

  const {
    register,
    handleSubmit,
    control,
    watch,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      pattern: '',
      category_id: '',
      match_type: 'contains',
      priority: 0,
      active: true,
    },
  });

  useEffect(() => {
    if (open) {
      reset({
        pattern: rule?.pattern ?? '',
        category_id: rule?.category_id ?? '',
        match_type: rule?.match_type ?? 'contains',
        priority: rule?.priority ?? 0,
        active: rule?.active ?? true,
      });
    }
  }, [open, rule, reset]);

  const matchType = watch('match_type');

  const categoryOptions = (categories ?? []).map((c) => ({
    value: c.id,
    label: `${c.name} (${c.kind === 'income' ? 'Receita' : 'Despesa'})`,
  }));

  const onSubmit = (values: FormValues) => {
    const payload = {
      pattern: values.pattern,
      category_id: values.category_id,
      match_type: values.match_type,
      priority: values.priority,
      active: values.active,
    };
    const opts = { onSuccess: onClose };
    if (isEdit && rule) update.mutate({ id: rule.id, payload }, opts);
    else create.mutate(payload, opts);
  };

  const pending = create.isPending || update.isPending;

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={isEdit ? 'Editar regra' : 'Nova regra'}
      description="Categorize automaticamente transações cuja descrição combine com a palavra-chave."
      footer={
        <>
          <Button variant="ghost" onClick={onClose} type="button">
            Cancelar
          </Button>
          <Button type="submit" form="rule-form" loading={pending}>
            {isEdit ? 'Salvar alterações' : 'Criar regra'}
          </Button>
        </>
      }
    >
      <form id="rule-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <Input
          label="Palavra-chave"
          placeholder="Ex.: uber"
          error={errors.pattern?.message}
          hint={
            matchType === 'regex'
              ? 'A expressão regular precisa ser válida (sintaxe JavaScript/RE2).'
              : undefined
          }
          {...register('pattern')}
        />

        <Select
          label="Categoria"
          placeholder="Selecione a categoria"
          options={categoryOptions}
          error={errors.category_id?.message}
          {...register('category_id')}
        />

        <Select
          label="Tipo de correspondência"
          options={matchTypeOptions}
          error={errors.match_type?.message}
          {...register('match_type')}
        />

        <Input
          label="Prioridade"
          type="number"
          min={0}
          inputMode="numeric"
          error={errors.priority?.message}
          hint="Regras de prioridade mais alta vencem quando várias combinam."
          {...register('priority')}
        />

        <Controller
          control={control}
          name="active"
          render={({ field }) => (
            <button
              type="button"
              role="switch"
              aria-checked={field.value}
              onClick={() => field.onChange(!field.value)}
              className="flex w-full items-center justify-between rounded-lg border border-gray-200 px-4 py-3 text-left transition hover:bg-gray-50 dark:border-ink-border dark:hover:bg-ink-elevated"
            >
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                Regra ativa
              </span>
              <span
                className={cn(
                  'relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors',
                  field.value ? 'bg-primary' : 'bg-gray-300 dark:bg-ink-border',
                )}
              >
                <span
                  className={cn(
                    'inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform',
                    field.value ? 'translate-x-4' : 'translate-x-0.5',
                  )}
                />
              </span>
            </button>
          )}
        />
      </form>
    </Modal>
  );
}
