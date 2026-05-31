import { useEffect } from 'react';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Modal, Input, Button, ColorPicker } from '@/components/ui';
import { useCreateTag, useUpdateTag } from '@/hooks';
import type { Tag } from '@/types';

const schema = z.object({
  name: z.string().min(1, 'Informe o nome da tag'),
  color: z.string().min(1),
});

type FormValues = z.infer<typeof schema>;

interface Props {
  open: boolean;
  onClose: () => void;
  tag?: Tag | null;
  /** Called with the persisted tag after a successful create/update. */
  onSaved?: (tag: Tag) => void;
}

export function TagFormModal({ open, onClose, tag, onSaved }: Props) {
  const create = useCreateTag();
  const update = useUpdateTag();
  const isEdit = Boolean(tag);

  const {
    register,
    handleSubmit,
    control,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: '', color: '#10b981' },
  });

  useEffect(() => {
    if (open) {
      reset({
        name: tag?.name ?? '',
        color: tag?.color ?? '#10b981',
      });
    }
  }, [open, tag, reset]);

  const onSubmit = (values: FormValues) => {
    const opts = {
      onSuccess: (saved: Tag) => {
        onSaved?.(saved);
        onClose();
      },
    };
    if (isEdit && tag) update.mutate({ id: tag.id, payload: values }, opts);
    else create.mutate(values, opts);
  };

  const pending = create.isPending || update.isPending;

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={isEdit ? 'Editar tag' : 'Nova tag'}
      description="Use tags para marcar e filtrar transações de forma transversal."
      size="sm"
      footer={
        <>
          <Button variant="ghost" onClick={onClose} type="button">
            Cancelar
          </Button>
          <Button type="submit" form="tag-form" loading={pending}>
            {isEdit ? 'Salvar alterações' : 'Criar tag'}
          </Button>
        </>
      }
    >
      <form id="tag-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <Input
          label="Nome da tag"
          placeholder="Ex.: Reembolsável"
          error={errors.name?.message}
          {...register('name')}
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
