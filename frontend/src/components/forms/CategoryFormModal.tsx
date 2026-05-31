import { useEffect } from 'react';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Modal, Input, Button, ColorPicker, IconPicker } from '@/components/ui';
import { cn } from '@/lib/cn';
import { useCreateCategory, useUpdateCategory } from '@/hooks';
import type { Category, CategoryKind } from '@/types';

const schema = z.object({
  name: z.string().min(1, 'Informe o nome da categoria'),
  kind: z.enum(['income', 'expense']),
  color: z.string().min(1),
  icon: z.string().min(1),
});

type FormValues = z.infer<typeof schema>;

interface Props {
  open: boolean;
  onClose: () => void;
  category?: Category | null;
}

export function CategoryFormModal({ open, onClose, category }: Props) {
  const create = useCreateCategory();
  const update = useUpdateCategory();
  const isEdit = Boolean(category);

  const {
    register,
    handleSubmit,
    control,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: '', kind: 'expense', color: '#10b981', icon: 'receipt' },
  });

  useEffect(() => {
    if (open) {
      reset({
        name: category?.name ?? '',
        kind: category?.kind ?? 'expense',
        color: category?.color ?? '#10b981',
        icon: category?.icon ?? 'receipt',
      });
    }
  }, [open, category, reset]);

  const onSubmit = (values: FormValues) => {
    const opts = { onSuccess: onClose };
    if (isEdit && category) update.mutate({ id: category.id, payload: values }, opts);
    else create.mutate(values, opts);
  };

  const pending = create.isPending || update.isPending;

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={isEdit ? 'Editar categoria' : 'Nova categoria'}
      description="Organize suas transações por categorias de receita ou despesa."
      footer={
        <>
          <Button variant="ghost" onClick={onClose} type="button">
            Cancelar
          </Button>
          <Button type="submit" form="category-form" loading={pending}>
            {isEdit ? 'Salvar alterações' : 'Criar categoria'}
          </Button>
        </>
      }
    >
      <form id="category-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <Input
          label="Nome da categoria"
          placeholder="Ex.: Aluguel"
          error={errors.name?.message}
          {...register('name')}
        />

        <div>
          <span className="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
            Tipo
          </span>
          <Controller
            control={control}
            name="kind"
            render={({ field }) => (
              <div className="grid grid-cols-2 gap-2">
                {(['expense', 'income'] as CategoryKind[]).map((kind) => (
                  <button
                    key={kind}
                    type="button"
                    onClick={() => field.onChange(kind)}
                    className={cn(
                      'rounded-lg border px-4 py-2.5 text-sm font-medium transition',
                      field.value === kind
                        ? kind === 'income'
                          ? 'border-primary bg-primary/10 text-primary-700 dark:text-primary-300'
                          : 'border-red-400 bg-red-500/10 text-red-600 dark:text-red-400'
                        : 'border-gray-200 text-gray-600 hover:bg-gray-50 dark:border-ink-border dark:text-gray-400 dark:hover:bg-ink-elevated',
                    )}
                  >
                    {kind === 'income' ? 'Receita' : 'Despesa'}
                  </button>
                ))}
              </div>
            )}
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
