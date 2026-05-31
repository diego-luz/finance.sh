import { useEffect } from 'react';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Modal, Input, Button, ColorPicker } from '@/components/ui';
import { useCreateCreditCard, useUpdateCreditCard } from '@/hooks';
import { centsToInput, parseCurrencyToCents } from '@/lib/currency';
import type { CreditCard } from '@/types';

const daySchema = z
  .string()
  .min(1, 'Informe o dia')
  .refine(
    (v) => {
      const n = Number(v);
      return Number.isInteger(n) && n >= 1 && n <= 31;
    },
    { message: 'Dia entre 1 e 31' },
  );

const schema = z.object({
  name: z.string().min(1, 'Informe o nome do cartão'),
  limit: z.string().min(1, 'Informe o limite'),
  closing_day: daySchema,
  due_day: daySchema,
  color: z.string().min(1),
});

type FormValues = z.infer<typeof schema>;

interface Props {
  open: boolean;
  onClose: () => void;
  creditCard?: CreditCard | null;
}

export function CreditCardFormModal({ open, onClose, creditCard }: Props) {
  const create = useCreateCreditCard();
  const update = useUpdateCreditCard();
  const isEdit = Boolean(creditCard);

  const {
    register,
    handleSubmit,
    control,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: '',
      limit: '',
      closing_day: '1',
      due_day: '10',
      color: '#6366f1',
    },
  });

  useEffect(() => {
    if (open) {
      reset({
        name: creditCard?.name ?? '',
        limit: creditCard ? centsToInput(creditCard.limit) : '',
        closing_day: String(creditCard?.closing_day ?? 1),
        due_day: String(creditCard?.due_day ?? 10),
        color: creditCard?.color ?? '#6366f1',
      });
    }
  }, [open, creditCard, reset]);

  const onSubmit = (values: FormValues) => {
    const payload = {
      name: values.name,
      limit: parseCurrencyToCents(values.limit),
      closing_day: Number(values.closing_day),
      due_day: Number(values.due_day),
      color: values.color,
    };
    const opts = { onSuccess: onClose };
    if (isEdit && creditCard) update.mutate({ id: creditCard.id, payload }, opts);
    else create.mutate(payload, opts);
  };

  const pending = create.isPending || update.isPending;

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={isEdit ? 'Editar cartão' : 'Novo cartão'}
      description="Defina limite, fechamento e vencimento do cartão de crédito."
      footer={
        <>
          <Button variant="ghost" onClick={onClose} type="button">
            Cancelar
          </Button>
          <Button type="submit" form="credit-card-form" loading={pending}>
            {isEdit ? 'Salvar alterações' : 'Criar cartão'}
          </Button>
        </>
      }
    >
      <form id="credit-card-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <Input
          label="Nome do cartão"
          placeholder="Ex.: Nubank PJ"
          error={errors.name?.message}
          {...register('name')}
        />
        <Input
          label="Limite"
          placeholder="0,00"
          inputMode="decimal"
          rightAddon="R$"
          error={errors.limit?.message}
          {...register('limit')}
        />
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Input
            label="Dia de fechamento"
            type="number"
            min={1}
            max={31}
            error={errors.closing_day?.message}
            {...register('closing_day')}
          />
          <Input
            label="Dia de vencimento"
            type="number"
            min={1}
            max={31}
            error={errors.due_day?.message}
            {...register('due_day')}
          />
        </div>

        <Controller
          control={control}
          name="color"
          render={({ field }) => (
            <ColorPicker label="Cor" value={field.value} onChange={field.onChange} />
          )}
        />
      </form>
    </Modal>
  );
}
