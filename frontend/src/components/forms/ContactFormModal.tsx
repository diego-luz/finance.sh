import { useEffect } from 'react';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Modal, Input, Button } from '@/components/ui';
import { cn } from '@/lib/cn';
import { useCreateContact, useUpdateContact } from '@/hooks';
import type { Contact, ContactType } from '@/types';

const schema = z.object({
  name: z.string().min(1, 'Informe o nome do contato'),
  type: z.enum(['customer', 'supplier', 'both']),
  document: z.string().optional(),
  email: z.string().email('E-mail inválido').optional().or(z.literal('')),
  phone: z.string().optional(),
  notes: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

interface Props {
  open: boolean;
  onClose: () => void;
  contact?: Contact | null;
}

const typeTabs: { value: ContactType; label: string }[] = [
  { value: 'customer', label: 'Cliente' },
  { value: 'supplier', label: 'Fornecedor' },
  { value: 'both', label: 'Ambos' },
];

export function ContactFormModal({ open, onClose, contact }: Props) {
  const create = useCreateContact();
  const update = useUpdateContact();
  const isEdit = Boolean(contact);

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
      type: 'customer',
      document: '',
      email: '',
      phone: '',
      notes: '',
    },
  });

  useEffect(() => {
    if (open) {
      reset({
        name: contact?.name ?? '',
        type: contact?.type ?? 'customer',
        document: contact?.document ?? '',
        email: contact?.email ?? '',
        phone: contact?.phone ?? '',
        notes: contact?.notes ?? '',
      });
    }
  }, [open, contact, reset]);

  const onSubmit = (values: FormValues) => {
    const payload = {
      name: values.name,
      type: values.type,
      document: values.document || undefined,
      email: values.email || undefined,
      phone: values.phone || undefined,
      notes: values.notes || undefined,
    };
    const opts = { onSuccess: onClose };
    if (isEdit && contact) update.mutate({ id: contact.id, payload }, opts);
    else create.mutate(payload, opts);
  };

  const pending = create.isPending || update.isPending;

  return (
    <Modal
      open={open}
      onClose={onClose}
      variant="drawer"
      title={isEdit ? 'Editar contato' : 'Novo contato'}
      footer={
        <>
          <Button variant="ghost" onClick={onClose} type="button">
            Cancelar
          </Button>
          <Button type="submit" form="contact-form" loading={pending}>
            {isEdit ? 'Salvar' : 'Adicionar'}
          </Button>
        </>
      }
    >
      <form id="contact-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <Input
          label="Nome"
          placeholder="Ex.: Fornecedor Acme Ltda"
          error={errors.name?.message}
          {...register('name')}
        />

        <div>
          <span className="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
            Tipo
          </span>
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
        </div>

        <Input
          label="Documento (CPF/CNPJ)"
          placeholder="Opcional"
          error={errors.document?.message}
          {...register('document')}
        />

        <Input
          label="E-mail"
          type="email"
          placeholder="Opcional"
          error={errors.email?.message}
          {...register('email')}
        />

        <Input
          label="Telefone"
          placeholder="Opcional"
          error={errors.phone?.message}
          {...register('phone')}
        />

        <div>
          <label
            htmlFor="contact-notes"
            className="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300"
          >
            Observações
          </label>
          <textarea
            id="contact-notes"
            rows={3}
            className="input-base resize-none"
            placeholder="Notas adicionais (opcional)"
            {...register('notes')}
          />
        </div>
      </form>
    </Modal>
  );
}
