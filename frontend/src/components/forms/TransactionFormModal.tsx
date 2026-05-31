import { useEffect, useRef, useState } from 'react';
import { Minus, Plus, Sparkles, X as XIcon } from 'lucide-react';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Modal, Input, Select, Button } from '@/components/ui';
import { AttachmentsSection } from '@/components/AttachmentsSection';
import { TagPicker } from '@/components/forms/TagPicker';
import { cn } from '@/lib/cn';
import {
  useAccounts,
  useCategories,
  useContacts,
  useCreateTransaction,
  useDebouncedValue,
  useUpdateTransaction,
} from '@/hooks';
import { categorizationService } from '@/services';
import { centsToInput, formatCurrency, parseCurrencyToCents } from '@/lib/currency';
import { toInputDate, toISODate } from '@/lib/date';
import type { Transaction, TransactionType } from '@/types';

const schema = z
  .object({
    type: z.enum(['income', 'expense', 'transfer']),
    account_id: z.string().min(1, 'Selecione a conta'),
    transfer_account_id: z.string().optional(),
    category_id: z.string().optional(),
    amount: z.string().min(1, 'Informe o valor'),
    description: z.string().min(1, 'Informe uma descrição'),
    date: z.string().min(1, 'Informe a data'),
    due_date: z.string().optional(),
    contact_id: z.string().optional(),
    paid: z.boolean(),
    notes: z.string().optional(),
    installments: z.coerce.number().int().min(1).max(360),
    tag_ids: z.array(z.string()),
  })
  .refine(
    (v) => v.type !== 'transfer' || (v.transfer_account_id && v.transfer_account_id.length > 0),
    { message: 'Selecione a conta de destino', path: ['transfer_account_id'] },
  )
  .refine((v) => v.type === 'transfer' || parseCurrencyToCents(v.amount) > 0, {
    message: 'O valor deve ser maior que zero',
    path: ['amount'],
  });

type FormValues = z.infer<typeof schema>;

interface Props {
  open: boolean;
  onClose: () => void;
  transaction?: Transaction | null;
}

const typeTabs: { value: TransactionType; label: string }[] = [
  { value: 'expense', label: 'Despesa' },
  { value: 'income', label: 'Receita' },
  { value: 'transfer', label: 'Transferência' },
];

export function TransactionFormModal({ open, onClose, transaction }: Props) {
  const { data: accounts } = useAccounts();
  const { data: categories } = useCategories();
  const { data: contacts } = useContacts();
  const create = useCreateTransaction();
  const update = useUpdateTransaction();
  const isEdit = Boolean(transaction);

  const {
    register,
    handleSubmit,
    control,
    watch,
    reset,
    setValue,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      type: 'expense',
      account_id: '',
      transfer_account_id: '',
      category_id: '',
      amount: '',
      description: '',
      date: toInputDate(),
      due_date: '',
      contact_id: '',
      paid: true,
      notes: '',
      installments: 1,
      tag_ids: [],
    },
  });

  const type = watch('type');
  const installments = Number(watch('installments')) || 1;
  const amountInput = watch('amount');
  const description = watch('description');
  const categoryId = watch('category_id');

  // ----- Automatic category suggestion ------------------------------------
  // While the user types a description (debounced) and hasn't picked a
  // category yet, ask the backend for a likely category and softly pre-select
  // it. Failures are ignored so the form never breaks.
  const [suggestedName, setSuggestedName] = useState<string | null>(null);
  const debouncedDescription = useDebouncedValue(description, 400);
  // Tracks the category id we auto-applied, so we can tell it apart from a
  // user's manual choice and avoid re-suggesting over their selection.
  const autoAppliedRef = useRef<string | null>(null);

  const dismissSuggestion = () => {
    // Clear only the value we auto-applied; leave a manual pick untouched.
    if (categoryId && categoryId === autoAppliedRef.current) {
      setValue('category_id', '', { shouldDirty: false });
    }
    setSuggestedName(null);
    autoAppliedRef.current = null;
  };

  // Clear the hint as soon as the user changes the category themselves.
  useEffect(() => {
    if (categoryId && categoryId !== autoAppliedRef.current) {
      setSuggestedName(null);
    }
  }, [categoryId]);

  useEffect(() => {
    if (!open) return;
    const text = debouncedDescription?.trim() ?? '';
    // Only suggest for income/expense, when no category is chosen yet.
    if (type === 'transfer' || text.length < 2) return;
    if (categoryId && categoryId !== autoAppliedRef.current) return;

    const suggestType: 'income' | 'expense' = type;
    let cancelled = false;
    categorizationService
      .suggest(text, suggestType)
      .then((res) => {
        if (cancelled || !res.category_id || !res.category) return;
        // Never override a category the user explicitly picked.
        if (categoryId && categoryId !== autoAppliedRef.current) return;
        autoAppliedRef.current = res.category_id;
        setValue('category_id', res.category_id, { shouldDirty: false });
        setSuggestedName(res.category.name);
      })
      .catch(() => {
        /* ignore suggestion failures */
      });

    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debouncedDescription, type, open]);
  // The "Parcelar" control only makes sense for new expenses.
  const canInstallment = !isEdit && type === 'expense';
  const amountCents = parseCurrencyToCents(amountInput ?? '');
  const perInstallment =
    installments > 1 && amountCents > 0 ? Math.floor(amountCents / installments) : 0;

  useEffect(() => {
    if (open) {
      setSuggestedName(null);
      autoAppliedRef.current = null;
      reset({
        type: transaction?.type ?? 'expense',
        account_id: transaction?.account_id ?? accounts?.[0]?.id ?? '',
        transfer_account_id: transaction?.transfer_account_id ?? '',
        category_id: transaction?.category?.id ?? '',
        amount: transaction ? centsToInput(transaction.amount) : '',
        description: transaction?.description ?? '',
        date: toInputDate(transaction?.date),
        due_date: transaction?.due_date ? toInputDate(transaction.due_date) : '',
        contact_id: transaction?.contact_id ?? transaction?.contact?.id ?? '',
        paid: transaction?.paid ?? true,
        notes: transaction?.notes ?? '',
        installments: 1,
        tag_ids: transaction?.tags?.map((t) => t.id) ?? [],
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, transaction, accounts]);

  const accountOptions = (accounts ?? []).map((a) => ({ value: a.id, label: a.name }));
  const categoryOptions = (categories ?? [])
    .filter((c) => (type === 'income' ? c.kind === 'income' : c.kind === 'expense'))
    .map((c) => ({ value: c.id, label: c.name }));
  const contactOptions = (contacts ?? []).map((c) => ({ value: c.id, label: c.name }));

  const onSubmit = (values: FormValues) => {
    const payload = {
      account_id: values.account_id,
      category_id: values.type === 'transfer' || !values.category_id ? undefined : values.category_id,
      transfer_account_id:
        values.type === 'transfer' ? values.transfer_account_id : undefined,
      type: values.type,
      amount: parseCurrencyToCents(values.amount),
      description: values.description,
      date: toISODate(values.date),
      due_date: values.due_date ? toISODate(values.due_date) : undefined,
      contact_id: values.type === 'transfer' || !values.contact_id ? undefined : values.contact_id,
      paid: values.paid,
      notes: values.notes || undefined,
      installments:
        !isEdit && values.type === 'expense' && values.installments > 1
          ? values.installments
          : undefined,
      tag_ids: values.tag_ids,
    };
    const opts = { onSuccess: onClose };
    if (isEdit && transaction) update.mutate({ id: transaction.id, payload }, opts);
    else create.mutate(payload, opts);
  };

  const pending = create.isPending || update.isPending;

  return (
    <Modal
      open={open}
      onClose={onClose}
      variant="drawer"
      title={isEdit ? 'Editar transação' : 'Nova transação'}
      footer={
        <>
          <Button variant="ghost" onClick={onClose} type="button">
            Cancelar
          </Button>
          <Button type="submit" form="transaction-form" loading={pending}>
            {isEdit ? 'Salvar' : 'Adicionar'}
          </Button>
        </>
      }
    >
      <form id="transaction-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        {/* Type tabs */}
        <Controller
          control={control}
          name="type"
          render={({ field }) => (
            <div className="grid grid-cols-3 gap-1 rounded-lg bg-gray-100 p-1 dark:bg-ink-elevated">
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
          placeholder="Ex.: Pagamento de fornecedor"
          error={errors.description?.message}
          {...register('description')}
        />

        <Select
          label={type === 'transfer' ? 'Conta de origem' : 'Conta'}
          placeholder="Selecione uma conta"
          options={accountOptions}
          error={errors.account_id?.message}
          {...register('account_id')}
        />

        {type === 'transfer' ? (
          <Select
            label="Conta de destino"
            placeholder="Selecione a conta de destino"
            options={accountOptions}
            error={errors.transfer_account_id?.message}
            {...register('transfer_account_id')}
          />
        ) : (
          <div>
            <Select
              label="Categoria"
              placeholder="Sem categoria"
              options={categoryOptions}
              {...register('category_id')}
            />
            {suggestedName && categoryId && (
              <div className="mt-1.5 flex items-center gap-1.5 text-xs text-primary">
                <Sparkles className="h-3.5 w-3.5 shrink-0" />
                <span>Sugerido automaticamente</span>
                <button
                  type="button"
                  onClick={dismissSuggestion}
                  className="ml-auto inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-gray-400 transition hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-ink-elevated"
                  aria-label="Dispensar sugestão"
                >
                  <XIcon className="h-3.5 w-3.5" /> Limpar
                </button>
              </div>
            )}
          </div>
        )}

        <div className="grid grid-cols-2 gap-3">
          <Input label="Data" type="date" error={errors.date?.message} {...register('date')} />
          <Input
            label="Vencimento"
            type="date"
            error={errors.due_date?.message}
            {...register('due_date')}
          />
        </div>

        {type !== 'transfer' && (
          <Select
            label="Contato"
            placeholder="Sem contato"
            options={contactOptions}
            {...register('contact_id')}
          />
        )}

        <Controller
          control={control}
          name="tag_ids"
          render={({ field }) => (
            <TagPicker value={field.value} onChange={field.onChange} />
          )}
        />

        {canInstallment && (
          <Controller
            control={control}
            name="installments"
            render={({ field }) => {
              const value = Number(field.value) || 1;
              const setValue = (n: number) =>
                field.onChange(Math.max(1, Math.min(360, n)));
              return (
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
                    Parcelas
                  </label>
                  <div className="flex items-center gap-2">
                    <button
                      type="button"
                      onClick={() => setValue(value - 1)}
                      disabled={value <= 1}
                      className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border border-gray-300 text-gray-600 transition hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-ink-border dark:text-gray-300 dark:hover:bg-ink-elevated"
                      aria-label="Diminuir parcelas"
                    >
                      <Minus className="h-4 w-4" />
                    </button>
                    <input
                      type="number"
                      inputMode="numeric"
                      min={1}
                      max={360}
                      value={value}
                      onChange={(e) => setValue(parseInt(e.target.value, 10) || 1)}
                      className="input-base w-20 text-center tabular-nums"
                      aria-label="Número de parcelas"
                    />
                    <button
                      type="button"
                      onClick={() => setValue(value + 1)}
                      disabled={value >= 360}
                      className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border border-gray-300 text-gray-600 transition hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-ink-border dark:text-gray-300 dark:hover:bg-ink-elevated"
                      aria-label="Aumentar parcelas"
                    >
                      <Plus className="h-4 w-4" />
                    </button>
                    {installments > 1 && perInstallment > 0 && (
                      <span className="money-sensitive ml-1 text-sm font-medium text-primary">
                        {installments}x de {formatCurrency(perInstallment)}
                      </span>
                    )}
                  </div>
                  {installments > 1 && (
                    <p className="mt-1 text-xs text-gray-400">
                      Serão criadas {installments} parcelas mensais.
                    </p>
                  )}
                </div>
              );
            }}
          />
        )}

        <div className="space-y-2">
          <label className="flex cursor-pointer items-center gap-2.5 text-sm text-gray-700 dark:text-gray-300">
            <input
              type="checkbox"
              className="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary dark:border-ink-border dark:bg-ink-elevated"
              {...register('paid')}
            />
            {type === 'income' ? 'Recebido' : 'Pago'}
          </label>
        </div>

        <div>
          <label
            htmlFor="notes"
            className="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300"
          >
            Observações
          </label>
          <textarea
            id="notes"
            rows={3}
            className="input-base resize-none"
            placeholder="Notas adicionais (opcional)"
            {...register('notes')}
          />
        </div>
      </form>

      {/* Receipt attachments — only available once the transaction exists. */}
      {isEdit && transaction?.id && (
        <div className="mt-6 border-t border-gray-100 pt-5 dark:border-ink-border">
          <h3 className="mb-3 font-heading text-sm font-semibold text-gray-900 dark:text-gray-100">
            Comprovantes
          </h3>
          <AttachmentsSection transactionId={transaction.id} />
        </div>
      )}
    </Modal>
  );
}
