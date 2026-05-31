import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Copy, Check, Mail } from 'lucide-react';
import { Modal, Input, Select, Button } from '@/components/ui';
import { useCreateInvitation } from '@/hooks';
import type { Invitation, OrgRole } from '@/types';

const schema = z.object({
  email: z.string().min(1, 'Informe o e-mail').email('E-mail inválido'),
  role: z.enum(['admin', 'member', 'viewer']),
});

type FormValues = z.infer<typeof schema>;

const roleOptions: { value: OrgRole; label: string }[] = [
  { value: 'admin', label: 'Administrador' },
  { value: 'member', label: 'Membro' },
  { value: 'viewer', label: 'Visualizador' },
];

interface Props {
  open: boolean;
  onClose: () => void;
}

export function InviteMemberModal({ open, onClose }: Props) {
  const create = useCreateInvitation();
  const [created, setCreated] = useState<Invitation | null>(null);
  const [copied, setCopied] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { email: '', role: 'member' },
  });

  useEffect(() => {
    if (open) {
      reset({ email: '', role: 'member' });
      setCreated(null);
      setCopied(false);
    }
  }, [open, reset]);

  const onSubmit = (values: FormValues) => {
    create.mutate(values, {
      onSuccess: (inv) => setCreated(inv),
    });
  };

  const copyToken = async () => {
    if (!created?.token) return;
    try {
      await navigator.clipboard.writeText(created.token);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard may be unavailable; the value is still selectable in the field.
    }
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Convidar membro"
      description="Envie um convite por e-mail para colaborar nesta organização."
      footer={
        created ? (
          <Button onClick={onClose} type="button">
            Concluir
          </Button>
        ) : (
          <>
            <Button variant="ghost" onClick={onClose} type="button">
              Cancelar
            </Button>
            <Button type="submit" form="invite-form" loading={create.isPending}>
              Enviar convite
            </Button>
          </>
        )
      }
    >
      {created ? (
        <div className="space-y-4">
          <div className="rounded-xl border border-primary/30 bg-primary/5 p-4">
            <div className="mb-2 flex h-10 w-10 items-center justify-center rounded-xl bg-primary/10 text-primary">
              <Mail className="h-5 w-5" />
            </div>
            <p className="text-sm text-gray-700 dark:text-gray-300">
              Convite criado para{' '}
              <span className="font-medium text-gray-900 dark:text-gray-100">{created.email}</span>.
              O envio por e-mail ainda não está habilitado — compartilhe o token abaixo com a pessoa
              convidada.
            </p>
          </div>
          {created.token && (
            <div>
              <span className="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
                Token do convite
              </span>
              <div className="flex items-center gap-2">
                <Input readOnly value={created.token} className="font-mono text-xs" />
                <Button type="button" variant="outline" size="icon" onClick={copyToken} aria-label="Copiar token">
                  {copied ? <Check className="h-4 w-4 text-primary" /> : <Copy className="h-4 w-4" />}
                </Button>
              </div>
            </div>
          )}
        </div>
      ) : (
        <form id="invite-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
          <Input
            label="E-mail"
            type="email"
            placeholder="colega@empresa.com.br"
            leftIcon={<Mail className="h-4 w-4" />}
            error={errors.email?.message}
            {...register('email')}
          />
          <Select label="Função" options={roleOptions} {...register('role')} />
        </form>
      )}
    </Modal>
  );
}
