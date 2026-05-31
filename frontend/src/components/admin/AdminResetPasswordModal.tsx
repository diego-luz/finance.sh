import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Lock, KeyRound, Wand2, Copy, Check } from 'lucide-react';
import { Modal, Button, Input } from '@/components/ui';
import { useResetUserPassword } from '@/hooks';
import { suggestPassword } from '@/lib/password';
import { copyToClipboard } from '@/lib/clipboard';
import { useToast } from '@/contexts/ToastContext';
import type { AdminUser } from '@/types';

const schema = z
  .object({
    new_password: z.string().min(8, 'Mínimo 8 caracteres').max(72),
    confirm_password: z.string().min(1, 'Confirme a senha'),
  })
  .refine((v) => v.new_password === v.confirm_password, {
    message: 'As senhas não coincidem',
    path: ['confirm_password'],
  });

type FormValues = z.infer<typeof schema>;

/**
 * Super-admin: define a new password for a target user. On success the backend
 * sets must_change_password=true and revokes the user's sessions, so the next
 * login forces another change.
 */
export function AdminResetPasswordModal({
  user,
  onClose,
}: {
  user: AdminUser | null;
  onClose: () => void;
}) {
  const reset = useResetUserPassword();
  const toast = useToast();
  const [copied, setCopied] = useState(false);
  const {
    register,
    handleSubmit,
    reset: resetForm,
    setValue,
    getValues,
    formState: { errors },
  } = useForm<FormValues>({ resolver: zodResolver(schema) });

  const suggest = () => {
    const pw = suggestPassword(16);
    setValue('new_password', pw, { shouldValidate: true, shouldDirty: true });
    setValue('confirm_password', pw, { shouldValidate: true, shouldDirty: true });
  };

  const copy = async () => {
    const pw = getValues('new_password');
    if (!pw) {
      toast.error('Preencha ou sugira uma senha antes de copiar.');
      return;
    }
    const ok = await copyToClipboard(pw);
    if (ok) {
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } else {
      toast.error('Não foi possível copiar. Selecione e copie manualmente.');
    }
  };

  const close = () => {
    resetForm();
    onClose();
  };

  if (!user) return null;

  const onSubmit = (values: FormValues) =>
    reset.mutate(
      { userId: user.id, new_password: values.new_password },
      { onSuccess: close },
    );

  return (
    <Modal
      open={Boolean(user)}
      onClose={close}
      title="Resetar senha"
      description={`Definir uma nova senha para ${user.name} (${user.email}).`}
      footer={
        <>
          <Button variant="ghost" onClick={close}>
            Cancelar
          </Button>
          <Button type="submit" form="admin-reset-pw-form" loading={reset.isPending}>
            <KeyRound className="h-4 w-4" /> Definir senha
          </Button>
        </>
      }
    >
      <form
        id="admin-reset-pw-form"
        onSubmit={handleSubmit(onSubmit)}
        className="space-y-4"
        noValidate
      >
        <div className="flex items-center justify-between gap-2">
          <Button type="button" variant="outline" size="sm" onClick={suggest}>
            <Wand2 className="h-4 w-4" /> Sugerir senha segura
          </Button>
          <Button type="button" variant="ghost" size="sm" onClick={copy}>
            {copied ? <Check className="h-4 w-4 text-primary" /> : <Copy className="h-4 w-4" />}
            {copied ? 'Copiada' : 'Copiar'}
          </Button>
        </div>
        <Input
          label="Nova senha"
          type="text"
          autoComplete="new-password"
          placeholder="Mínimo 8 caracteres"
          leftIcon={<Lock className="h-4 w-4" />}
          error={errors.new_password?.message}
          hint="O usuário será obrigado a alterá-la no próximo acesso."
          {...register('new_password')}
        />
        <Input
          label="Confirmar senha"
          type="password"
          autoComplete="new-password"
          leftIcon={<Lock className="h-4 w-4" />}
          error={errors.confirm_password?.message}
          {...register('confirm_password')}
        />
      </form>
    </Modal>
  );
}
