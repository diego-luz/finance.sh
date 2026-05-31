import { useEffect } from 'react';
import { CalendarClock } from 'lucide-react';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Modal, Input, Select, Button } from '@/components/ui';
import { cn } from '@/lib/cn';
import {
  useAccounts,
  useCategories,
  useContacts,
  useCreateRecurrence,
  useUpdateRecurrence,
} from '@/hooks';
import { centsToInput, parseCurrencyToCents } from '@/lib/currency';
import { toInputDate, toISODate } from '@/lib/date';
import { summarizeSchedule } from '@/lib/recurrence';
import type { RecurrenceFrequency, RecurrenceRule, RecurrenceType } from '@/types';

const schema = z
  .object({
    type: z.enum(['income', 'expense']),
    description: z.string().min(1, 'Informe uma descrição'),
    amount: z.string().min(1, 'Informe o valor'),
    account_id: z.string().min(1, 'Selecione a conta'),
    category_id: z.string().optional(),
    contact_id: z.string().optional(),
    frequency: z.enum(['daily', 'weekly', 'monthly', 'yearly']),
    interval: z.coerce.number().int().min(1).max(999),
    start_date: z.string().min(1, 'Informe a data de início'),
    end_date: z.string().optional(),
    max_occurrences: z.coerce.number().int().min(0).max(9999),
    paid: z.boolean(),
    active: z.boolean(),
  })
  .refine((v) => parseCurrencyToCents(v.amount) > 0, {
    message: 'O valor deve ser maior que zero',
    path: ['amount'],
  });

type FormValues = z.infer<typeof schema>;

interface Props {
  open: boolean;
  onClose: () => void;
  recurrence?: RecurrenceRule | null;
}

const typeTabs: { value: RecurrenceType; label: string }[] = [
  { value: 'expense', label: 'Despesa' },
  { value: 'income', label: 'Receita' },
];

const frequencyOptions: { value: RecurrenceFrequency; label: string }[] = [
  { value: 'daily', label: 'Diária' },
  { value: 'weekly', label: 'Semanal' },
  { value: 'monthly', label: 'Mensal' },
  { value: 'yearly', label: 'Anual' },
];

export function RecurrenceFormModal({ open, onClose, recurrence }: Props) {
  const { data: accounts } = useAccounts();
  const { data: categories } = useCategories();
  const { data: contacts } = useContacts();
  const create = useCreateRecurrence();
  const update = useUpdateRecurrence();
  const isEdit = Boolean(recurrence);

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
      type: 'expense',
      description: '',
      amount: '',
      account_id: '',
      category_id: '',
      contact_id: '',
      frequency: 'monthly',
      interval: 1,
      start_date: toInputDate(),
      end_date: '',
      max_occurrences: 0,
      paid: true,
      active: true,
    },
  });

  useEffect(() => {
    if (open) {
      reset({
        type: recurrence?.type ?? 'expense',
        description: recurrence?.description ?? '',
        amount: recurrence ? centsToInput(recurrence.amount) : '',
        account_id: recurrence?.account_id ?? accounts?.[0]?.id ?? '',
        category_id: recurrence?.category_id ?? recurrence?.category?.id ?? '',
        contact_id: recurrence?.contact_id ?? '',
        frequency: recurrence?.frequency ?? 'monthly',
        interval: recurrence?.interval ?? 1,
        start_date: toInputDate(recurrence?.start_date),
        end_date: recurrence?.end_date ? toInputDate(recurrence.end_date) : '',
        max_occurrences: recurrence?.max_occurrences ?? 0,
        paid: recurrence?.paid ?? true,
        active: recurrence?.active ?? true,
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, recurrence, accounts]);

  const type = watch('type');
  const frequency = watch('frequency');
  const interval = Number(watch('interval')) || 1;
  const startDate = watch('start_date');

  const accountOptions = (accounts ?? []).map((a) => ({ value: a.id, label: a.name }));
  const categoryOptions = (categories ?? [])
    .filter((c) => (type === 'income' ? c.kind === 'income' : c.kind === 'expense'))
    .map((c) => ({ value: c.id, label: c.name }));
  const contactOptions = (contacts ?? []).map((c) => ({ value: c.id, label: c.name }));

  const onSubmit = (values: FormValues) => {
    const payload = {
      type: values.type,
      description: values.description,
      amount: parseCurrencyToCents(values.amount),
      account_id: values.account_id,
      category_id: values.category_id || undefined,
      contact_id: values.contact_id || undefined,
      frequency: values.frequency,
      interval: values.interval,
      start_date: toISODate(values.start_date),
      end_date: values.end_date ? toISODate(values.end_date) : undefined,
      max_occurrences: values.max_occurrences,
      paid: values.paid,
      active: values.active,
    };
    const opts = { onSuccess: onClose };
    if (isEdit && recurrence) update.mutate({ id: recurrence.id, payload }, opts);
    else create.mutate(payload, opts);
  };

  const pending = create.isPending || update.isPending;

  return (
    <Modal
      open={open}
      onClose={onClose}
      variant="drawer"
      title={isEdit ? 'Editar recorrência' : 'Nova recorrência'}
      description="Gere transações automaticamente em um intervalo definido."
      footer={
        <>
          <Button variant="ghost" onClick={onClose} type="button">
            Cancelar
          </Button>
          <Button type="submit" form="recurrence-form" loading={pending}>
            {isEdit ? 'Salvar alterações' : 'Criar recorrência'}
          </Button>
        </>
      }
    >
      <form id="recurrence-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        {/* Type tabs */}
        <Controller
          control={control}
          name="type"
          render={({ field }) => (
            <div className="grid grid-cols-2 gap-1 rounded-lg bg-gray-100 p-1 dark:bg-ink-elevated">
              {typeTabs.map((t) => (
                <button
                  key={t.value}
                  type="button"
                  onClick={() => field.onChange(t.value)}
                  className={cn(
                    'rounded-md px-2 py-1.5 text-sm font-medium transition',
                    field.value === t.value
                      ? 'bg-white text-gray-900 shadow-sm dark:bg-ink-surface dark:text-gray-100'
                      : 'text-gray-500 hover:text-gray-700 dark:text-gray-400',
                  )}
                >
                  {t.label}
                </button>
              ))}
            </div>
          )}
        />

        <Input
          label="Valor"
          placeholder="0,00"
          inputMode="decimal"
          rightAddon="R$"
          error={errors.amount?.message}
          {...register('amount')}
        />

        <Input
          label="Descrição"
          placeholder="Ex.: Aluguel"
          error={errors.description?.message}
          {...register('description')}
        />

        <Select
          label="Conta"
          placeholder="Selecione uma conta"
          options={accountOptions}
          error={errors.account_id?.message}
          {...register('account_id')}
        />

        <Select
          label="Categoria"
          placeholder="Sem categoria"
          options={categoryOptions}
          {...register('category_id')}
        />

        <Select
          label="Contato"
          placeholder="Sem contato"
          options={contactOptions}
          {...register('contact_id')}
        />

        <div className="grid grid-cols-2 gap-3">
          <Select
            label="Frequência"
            options={frequencyOptions}
            error={errors.frequency?.message}
            {...register('frequency')}
          />
          <Input
            label="A cada"
            type="number"
            min={1}
            max={999}
            inputMode="numeric"
            error={errors.interval?.message}
            {...register('interval')}
          />
        </div>

        <div className="grid grid-cols-2 gap-3">
          <Input
            label="Início"
            type="date"
            error={errors.start_date?.message}
            {...register('start_date')}
          />
          <Input
            label="Término (opcional)"
            type="date"
            error={errors.end_date?.message}
            {...register('end_date')}
          />
        </div>

        <Input
          label="Máximo de ocorrências"
          type="number"
          min={0}
          inputMode="numeric"
          error={errors.max_occurrences?.message}
          hint="0 ou vazio = ilimitado."
          {...register('max_occurrences')}
        />

        {/* Human summary helper */}
        <div className="flex items-start gap-2 rounded-lg bg-primary/5 px-3 py-2.5 text-xs text-primary-700 dark:text-primary-300">
          <CalendarClock className="mt-0.5 h-3.5 w-3.5 shrink-0" />
          <span>{summarizeSchedule(frequency, interval, startDate)}</span>
        </div>

        <Controller
          control={control}
          name="paid"
          render={({ field }) => (
            <button
              type="button"
              role="switch"
              aria-checked={field.value}
              onClick={() => field.onChange(!field.value)}
              className="flex w-full items-center justify-between rounded-lg border border-gray-200 px-4 py-3 text-left transition hover:bg-gray-50 dark:border-ink-border dark:hover:bg-ink-elevated"
            >
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                Lançar como {type === 'income' ? 'recebido' : 'pago'}
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
                Recorrência ativa
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
