import { useEffect, useMemo } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Modal, Input, Select, Button } from '@/components/ui';
import { useCategories, useCreateBudget, useUpdateBudget } from '@/hooks';
import { centsToInput, parseCurrencyToCents } from '@/lib/currency';
import type { Budget } from '@/types';

const schema = z.object({
  category_id: z.string().min(1, 'Selecione uma categoria'),
  amount: z.string().min(1, 'Informe o valor'),
});

type FormValues = z.infer<typeof schema>;

interface Props {
  open: boolean;
  onClose: () => void;
  budget?: Budget | null;
  month: number;
  year: number;
}

export function BudgetFormModal({ open, onClose, budget, month, year }: Props) {
  const { data: categories } = useCategories();
  const create = useCreateBudget();
  const update = useUpdateBudget();
  const isEdit = Boolean(budget);

  const expenseCategories = useMemo(
    () => (categories ?? []).filter((c) => c.kind === 'expense'),
    [categories],
  );
  const categoryOptions = expenseCategories.map((c) => ({ value: c.id, label: c.name }));

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { category_id: '', amount: '' },
  });

  useEffect(() => {
    if (open) {
      reset({
        category_id: budget?.category_id ?? '',
        amount: budget ? centsToInput(budget.amount) : '',
      });
    }
  }, [open, budget, reset]);

  const onSubmit = (values: FormValues) => {
    const payload = {
      category_id: values.category_id,
      amount: parseCurrencyToCents(values.amount),
      month,
      year,
    };
    const opts = { onSuccess: onClose };
    if (isEdit && budget) update.mutate({ id: budget.id, payload }, opts);
    else create.mutate(payload, opts);
  };

  const pending = create.isPending || update.isPending;

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={isEdit ? 'Editar orçamento' : 'Novo orçamento'}
      description="Defina um limite de gastos por categoria neste período."
      footer={
        <>
          <Button variant="ghost" onClick={onClose} type="button">
            Cancelar
          </Button>
          <Button type="submit" form="budget-form" loading={pending}>
            {isEdit ? 'Salvar alterações' : 'Criar orçamento'}
          </Button>
        </>
      }
    >
      <form id="budget-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <Select
          label="Categoria de despesa"
          placeholder="Selecione uma categoria"
          options={categoryOptions}
          error={errors.category_id?.message}
          disabled={isEdit}
          {...register('category_id')}
        />
        {categoryOptions.length === 0 && (
          <p className="text-xs text-amber-600 dark:text-amber-400">
            Crie ao menos uma categoria de despesa para definir orçamentos.
          </p>
        )}
        <Input
          label="Valor do orçamento"
          placeholder="0,00"
          inputMode="decimal"
          rightAddon="R$"
          error={errors.amount?.message}
          {...register('amount')}
        />
      </form>
    </Modal>
  );
}
