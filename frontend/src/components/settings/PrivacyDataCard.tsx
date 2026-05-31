import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Download, Trash2, ShieldAlert, FileJson } from 'lucide-react';
import { Card, CardHeader, Button, Input, Modal } from '@/components/ui';
import { useExportMyData, useDeleteMyAccount } from '@/hooks';

const CONFIRM_WORD = 'EXCLUIR';

export function PrivacyDataCard() {
  const exportData = useExportMyData();
  const deleteAccount = useDeleteMyAccount();

  const [open, setOpen] = useState(false);
  const [password, setPassword] = useState('');
  const [confirmation, setConfirmation] = useState('');

  const canDelete = password.length > 0 && confirmation.trim().toUpperCase() === CONFIRM_WORD;

  const closeModal = () => {
    if (deleteAccount.isPending) return;
    setOpen(false);
    setPassword('');
    setConfirmation('');
  };

  const confirmDelete = () => {
    if (!canDelete) return;
    deleteAccount.mutate(password);
  };

  return (
    <Card className="mt-4">
      <CardHeader
        eyebrow="Privacidade e Dados"
        title="Seus dados (LGPD)"
        subtitle="Exporte uma cópia dos seus dados ou solicite a exclusão da sua conta."
      />

      <div className="space-y-3">
        {/* Export */}
        <div className="flex flex-col gap-3 rounded-lg border border-gray-100 px-4 py-3 dark:border-ink-border sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-3">
            <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
              <FileJson className="h-[18px] w-[18px]" />
            </span>
            <div>
              <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                Exportar meus dados
              </p>
              <p className="text-xs text-gray-400">
                Baixe um arquivo JSON com todos os seus dados pessoais.
              </p>
            </div>
          </div>
          <Button
            variant="outline"
            size="sm"
            loading={exportData.isPending}
            onClick={() => exportData.mutate()}
          >
            <Download className="h-4 w-4" /> Baixar meus dados (LGPD · JSON)
          </Button>
        </div>

        {/* Delete (danger) */}
        <div className="flex flex-col gap-3 rounded-lg border border-red-200 bg-red-50/50 px-4 py-3 dark:border-red-500/30 dark:bg-red-500/5 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-3">
            <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-red-500/10 text-red-500">
              <ShieldAlert className="h-[18px] w-[18px]" />
            </span>
            <div>
              <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                Excluir minha conta
              </p>
              <p className="text-xs text-gray-400">
                Esta ação é permanente e remove ou anonimiza seus dados.
              </p>
            </div>
          </div>
          <Button variant="danger" size="sm" onClick={() => setOpen(true)}>
            <Trash2 className="h-4 w-4" /> Excluir conta
          </Button>
        </div>

        <p className="text-xs text-gray-400">
          Saiba como tratamos seus dados na{' '}
          <Link to="/privacidade" className="text-primary hover:text-primary-600">
            Política de Privacidade
          </Link>
          .
        </p>
      </div>

      <Modal
        open={open}
        onClose={closeModal}
        title="Excluir conta permanentemente"
        description="Esta ação não pode ser desfeita."
        size="sm"
        footer={
          <>
            <Button variant="ghost" onClick={closeModal} disabled={deleteAccount.isPending}>
              Cancelar
            </Button>
            <Button
              variant="danger"
              onClick={confirmDelete}
              loading={deleteAccount.isPending}
              disabled={!canDelete}
            >
              <Trash2 className="h-4 w-4" /> Excluir definitivamente
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <div className="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-300">
            Ao confirmar, sua conta será excluída ou anonimizada e você perderá o acesso aos dados
            associados.
          </div>
          <Input
            label="Sua senha"
            type="password"
            autoComplete="current-password"
            placeholder="••••••••"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          <Input
            label={`Digite "${CONFIRM_WORD}" para confirmar`}
            placeholder={CONFIRM_WORD}
            value={confirmation}
            onChange={(e) => setConfirmation(e.target.value)}
          />
        </div>
      </Modal>
    </Card>
  );
}
