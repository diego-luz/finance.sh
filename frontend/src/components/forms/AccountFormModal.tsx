import { useEffect } from 'react';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Modal, Input, Select, Button, ColorPicker, IconPicker } from '@/components/ui';
import { useCreateAccount, useUpdateAccount } from '@/hooks';
import { centsToInput, parseCurrencyToCents } from '@/lib/currency';
import type { Account, AccountType } from '@/types';

const schema = z.object({
  name: z.string().min(1, 'Informe o nome da conta'),
  type: z.enum(['bank', 'wallet', 'investment', 'credit_card']),
  initial_balance: z.string(),
  color: z.string().min(1),
  icon: z.string().min(1),
});

type FormValues = z.infer<typeof schema>;

const typeOptions: { value: AccountType; label: string }[] = [
  { value: 'bank', label: 'Conta bancária' },
  { value: 'wallet', label: 'Carteira' },
  { value: 'investment', label: 'Investimento' },
  { value: 'credit_card', label: 'Cartão de crédito' },
];

interface Props {
  open: boolean;
  onClose: () => void;
  account?: Account | null;
}

export function AccountFormModal({ open, onClose, account }: Props) {
  const create = useCreateAccount();
  const update = useUpdateAccount();
  const isEdit = Boolean(account);

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
      type: 'bank',
      initial_balance: '',
      color: '#10b981',
      icon: 'landmark',
    },
  });

  useEffect(() => {
    if (open) {
      reset({
        name: account?.name ?? '',
        type: account?.type ?? 'bank',
        initial_balance: account ? centsToInput(account.initial_balance) : '',
        color: account?.color ?? '#10b981',
        icon: account?.icon ?? 'landmark',
      });
    }
  }, [open, account, reset]);

  const onSubmit = (values: FormValues) => {
    const payload = {
      name: values.name,
      type: values.type,
      initial_balance: parseCurrencyToCents(values.initial_balance),
      color: values.color,
      icon: values.icon,
    };
    const opts = { onSuccess: onClose };
    if (isEdit && account) update.mutate({ id: account.id, payload }, opts);
    else create.mutate(payload, opts);
  };

  const pending = create.isPending || update.isPending;

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={isEdit ? 'Editar conta' : 'Nova conta'}
      description="Defina nome, tipo e saldo inicial da conta."
      footer={
        <>
          <Button variant="ghost" onClick={onClose} type="button">
            Cancelar
          </Button>
          <Button type="submit" form="account-form" loading={pending}>
            {isEdit ? 'Salvar alterações' : 'Criar conta'}
          </Button>
        </>
      }
    >
      <form id="account-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <Input
          label="Nome da conta"
          placeholder="Ex.: Banco Inter PJ"
          error={errors.name?.message}
          {...register('name')}
        />
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Select label="Tipo" options={typeOptions} {...register('type')} />
          <Input
            label="Saldo inicial"
            placeholder="0,00"
            inputMode="decimal"
            rightAddon="R$"
            error={errors.initial_balance?.message}
            {...register('initial_balance')}
          />
        </div>

        <Controller
          control={control}
          name="color"
          render={({ field }) => (
            <ColorPicker label="Cor" value={field.value} onChange={field.onChange} />
          )}
        />

        <Controller
          control={control}
          name="icon"
          render={({ field }) => (
            <IconPicker label="Ícone" value={field.value} onChange={field.onChange} />
          )}
        />
      </form>
    </Modal>
  );
}
