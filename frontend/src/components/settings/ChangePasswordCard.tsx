import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { KeyRound, Lock } from 'lucide-react';
import { Card, CardHeader, Button, Input, Modal } from '@/components/ui';
import { useChangePassword } from '@/hooks';

const schema = z
  .object({
    current_password: z.string().min(1, 'Informe a senha atual'),
    new_password: z.string().min(8, 'A nova senha deve ter ao menos 8 caracteres').max(72),
    confirm_password: z.string().min(1, 'Confirme a nova senha'),
  })
  .refine((v) => v.new_password === v.confirm_password, {
    message: 'As senhas não coincidem',
    path: ['confirm_password'],
  })
  .refine((v) => v.new_password !== v.current_password, {
    message: 'A nova senha deve ser diferente da atual',
    path: ['new_password'],
  });

type FormValues = z.infer<typeof schema>;

/**
 * Self-service password change for the authenticated user. Reuses the same
 * /me/change-password endpoint that the forced first-login flow uses, but
 * surfaced from the Settings page as a regular action (any logged-in user).
 */
export function ChangePasswordCard() {
  const [open, setOpen] = useState(false);
  const change = useChangePassword();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormValues>({ resolver: zodResolver(schema) });

  const close = () => {
    setOpen(false);
    reset();
  };

  const onSubmit = (values: FormValues) =>
    change.mutate(
      { current_password: values.current_password, new_password: values.new_password },
      { onSuccess: close },
    );

  return (
    <Card className="mt-4">
      <CardHeader
        eyebrow="Segurança"
        title="Senha"
        subtitle="Altere sua senha de acesso a qualquer momento."
        action={
          <Button onClick={() => setOpen(true)}>
            <KeyRound className="h-4 w-4" /> Alterar senha
          </Button>
        }
      />

      <Modal
        open={open}
        onClose={close}
        title="Alterar senha"
        description="Informe sua senha atual e a nova senha desejada."
        footer={
          <>
            <Button variant="ghost" onClick={close}>
              Cancelar
            </Button>
            <Button
              type="submit"
              form="change-password-form"
              loading={change.isPending}
            >
              Salvar nova senha
            </Button>
          </>
        }
      >
        <form
          id="change-password-form"
          onSubmit={handleSubmit(onSubmit)}
          className="space-y-4"
          noValidate
        >
          <Input
            label="Senha atual"
            type="password"
            autoComplete="current-password"
            leftIcon={<Lock className="h-4 w-4" />}
            error={errors.current_password?.message}
            {...register('current_password')}
          />
          <Input
            label="Nova senha"
            type="password"
            autoComplete="new-password"
            placeholder="Mínimo 8 caracteres"
            leftIcon={<Lock className="h-4 w-4" />}
            error={errors.new_password?.message}
            {...register('new_password')}
          />
          <Input
            label="Confirmar nova senha"
            type="password"
            autoComplete="new-password"
            leftIcon={<Lock className="h-4 w-4" />}
            error={errors.confirm_password?.message}
            {...register('confirm_password')}
          />
        </form>
      </Modal>
    </Card>
  );
}
