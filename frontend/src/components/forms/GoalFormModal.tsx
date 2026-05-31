import { useEffect } from 'react';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Modal, Input, Button, ColorPicker } from '@/components/ui';
import { useCreateGoal, useUpdateGoal } from '@/hooks';
import { centsToInput, parseCurrencyToCents } from '@/lib/currency';
import { toInputDate, toISODate } from '@/lib/date';
import type { Goal } from '@/types';

const schema = z.object({
  name: z.string().min(1, 'Informe o nome da meta'),
  target_amount: z.string().min(1, 'Informe o valor alvo'),
  current_amount: z.string(),
  deadline: z.string().optional(),
  color: z.string().min(1),
});

type FormValues = z.infer<typeof schema>;

interface Props {
  open: boolean;
  onClose: () => void;
  goal?: Goal | null;
}

export function GoalFormModal({ open, onClose, goal }: Props) {
  const create = useCreateGoal();
  const update = useUpdateGoal();
  const isEdit = Boolean(goal);

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
      target_amount: '',
      current_amount: '',
      deadline: '',
      color: '#0ea5e9',
    },
  });

  useEffect(() => {
    if (open) {
      reset({
        name: goal?.name ?? '',
        target_amount: goal ? centsToInput(goal.target_amount) : '',
        current_amount: goal ? centsToInput(goal.current_amount) : '',
        deadline: goal?.deadline ? toInputDate(goal.deadline) : '',
        color: goal?.color ?? '#0ea5e9',
      });
    }
  }, [open, goal, reset]);

  const onSubmit = (values: FormValues) => {
    const payload = {
      name: values.name,
      target_amount: parseCurrencyToCents(values.target_amount),
      current_amount: parseCurrencyToCents(values.current_amount),
      deadline: values.deadline ? toISODate(values.deadline) : undefined,
      color: values.color,
    };
    const opts = { onSuccess: onClose };
    if (isEdit && goal) update.mutate({ id: goal.id, payload }, opts);
    else create.mutate(payload, opts);
  };

  const pending = create.isPending || update.isPending;

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={isEdit ? 'Editar meta' : 'Nova meta'}
      description="Defina um objetivo financeiro e acompanhe seu progresso."
      footer={
        <>
          <Button variant="ghost" onClick={onClose} type="button">
            Cancelar
          </Button>
          <Button type="submit" form="goal-form" loading={pending}>
            {isEdit ? 'Salvar alterações' : 'Criar meta'}
          </Button>
        </>
      }
    >
      <form id="goal-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <Input
          label="Nome da meta"
          placeholder="Ex.: Reserva de emergência"
          error={errors.name?.message}
          {...register('name')}
        />
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Input
            label="Valor alvo"
            placeholder="0,00"
            inputMode="decimal"
            rightAddon="R$"
            error={errors.target_amount?.message}
            {...register('target_amount')}
          />
          <Input
            label="Valor atual"
            placeholder="0,00"
            inputMode="decimal"
            rightAddon="R$"
            error={errors.current_amount?.message}
            {...register('current_amount')}
          />
        </div>
        <Input
          label="Prazo (opcional)"
          type="date"
          error={errors.deadline?.message}
          {...register('deadline')}
        />

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
