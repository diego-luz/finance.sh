import { useState } from 'react';
import { QRCodeSVG } from 'qrcode.react';
import { ShieldCheck, ShieldOff, KeyRound, Copy, Check } from 'lucide-react';
import { Card, CardHeader, Button, Input, Badge, Modal } from '@/components/ui';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';
import {
  useTwoFactorSetup,
  useEnableTwoFactor,
  useDisableTwoFactor,
} from '@/hooks';
import type { TwoFactorSetupResponse } from '@/types';
import { cn } from '@/lib/cn';

/** Reads the 2FA flag off the user; tolerant if the backend omits it. */
function useTwoFactorEnabled(): boolean {
  const user = useAuthStore((s) => s.user);
  return Boolean(user?.two_factor_enabled);
}

function CopyButton({ value, label }: { value: string; label?: string }) {
  const [copied, setCopied] = useState(false);
  const toast = useToast();
  const onCopy = async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      toast.success('Copiado para a área de transferência.');
      window.setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error('Não foi possível copiar.');
    }
  };
  return (
    <Button variant="outline" size="sm" type="button" onClick={onCopy}>
      {copied ? <Check className="h-4 w-4 text-primary" /> : <Copy className="h-4 w-4" />}
      {label ?? (copied ? 'Copiado' : 'Copiar')}
    </Button>
  );
}

export function TwoFactorCard() {
  const enabled = useTwoFactorEnabled();
  const setup = useTwoFactorSetup();
  const enable = useEnableTwoFactor();
  const disable = useDisableTwoFactor();

  const [setupData, setSetupData] = useState<TwoFactorSetupResponse | null>(null);
  const [enableCode, setEnableCode] = useState('');
  const [recoveryCodes, setRecoveryCodes] = useState<string[] | null>(null);

  const [disableOpen, setDisableOpen] = useState(false);
  const [disableCode, setDisableCode] = useState('');

  const startSetup = () => {
    setRecoveryCodes(null);
    setEnableCode('');
    setup.mutate(undefined, {
      onSuccess: (data) => setSetupData(data),
    });
  };

  const confirmEnable = () => {
    if (enableCode.length !== 6) return;
    enable.mutate(enableCode, {
      onSuccess: (data) => {
        setRecoveryCodes(data.recovery_codes);
        setSetupData(null);
        setEnableCode('');
      },
    });
  };

  const confirmDisable = () => {
    if (disableCode.length !== 6) return;
    disable.mutate(disableCode, {
      onSuccess: () => {
        setDisableOpen(false);
        setDisableCode('');
      },
    });
  };

  return (
    <>
      <Card>
        <CardHeader
          eyebrow="Segurança"
          title="Autenticação em dois fatores"
          subtitle="Adicione uma camada extra de proteção à sua conta."
          action={
            enabled ? (
              <Badge variant="success" dot>
                Ativada
              </Badge>
            ) : (
              <Badge variant="neutral" dot>
                Desativada
              </Badge>
            )
          }
        />

        {/* Recovery codes block (shown right after enabling) */}
        {recoveryCodes && (
          <div className="mb-4 rounded-xl border border-primary/30 bg-primary/5 p-4">
            <p className="flex items-center gap-2 text-sm font-semibold text-gray-900 dark:text-gray-100">
              <KeyRound className="h-4 w-4 text-primary" /> Códigos de recuperação
            </p>
            <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
              Guarde estes códigos em local seguro. Cada um pode ser usado uma única vez caso você
              perca o acesso ao aplicativo autenticador.
            </p>
            <div className="mt-3 grid grid-cols-2 gap-2 rounded-lg bg-white p-3 font-mono text-sm text-gray-800 dark:bg-ink-elevated dark:text-gray-200 sm:grid-cols-3">
              {recoveryCodes.map((code) => (
                <span key={code} className="tabular-nums">
                  {code}
                </span>
              ))}
            </div>
            <div className="mt-3 flex gap-2">
              <CopyButton value={recoveryCodes.join('\n')} label="Copiar códigos" />
              <Button variant="ghost" size="sm" onClick={() => setRecoveryCodes(null)}>
                Concluir
              </Button>
            </div>
          </div>
        )}

        {/* Enrollment flow (setup -> QR + confirm) */}
        {setupData ? (
          <div className="space-y-4">
            <p className="text-sm text-gray-600 dark:text-gray-300">
              Escaneie o QR Code abaixo com um aplicativo autenticador (Google Authenticator,
              Authy, 1Password) e informe o código gerado para confirmar.
            </p>
            <div className="flex flex-col items-center gap-4 sm:flex-row sm:items-start">
              <div className="rounded-xl border border-gray-200 bg-white p-3 dark:border-ink-border">
                <QRCodeSVG value={setupData.otpauth_url} size={160} level="M" />
              </div>
              <div className="min-w-0 flex-1 space-y-3">
                <div>
                  <p className="text-xs font-medium uppercase tracking-wide text-gray-400">
                    Chave manual
                  </p>
                  <div className="mt-1 flex items-center gap-2">
                    <code className="block min-w-0 flex-1 truncate rounded-lg bg-gray-100 px-3 py-2 font-mono text-xs text-gray-700 dark:bg-ink-elevated dark:text-gray-200">
                      {setupData.secret}
                    </code>
                    <CopyButton value={setupData.secret} label="" />
                  </div>
                </div>
                <Input
                  label="Código de confirmação"
                  inputMode="numeric"
                  autoComplete="one-time-code"
                  maxLength={6}
                  placeholder="000000"
                  className="text-center tracking-[0.4em]"
                  value={enableCode}
                  onChange={(e) => setEnableCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                />
              </div>
            </div>
            <div className="flex gap-2">
              <Button
                onClick={confirmEnable}
                loading={enable.isPending}
                disabled={enableCode.length !== 6}
              >
                <ShieldCheck className="h-4 w-4" /> Ativar
              </Button>
              <Button variant="ghost" onClick={() => setSetupData(null)} disabled={enable.isPending}>
                Cancelar
              </Button>
            </div>
          </div>
        ) : (
          !recoveryCodes && (
            <div className="flex items-center justify-between rounded-lg border border-gray-100 px-4 py-3 dark:border-ink-border">
              <div className="flex items-center gap-3">
                <span
                  className={cn(
                    'flex h-9 w-9 items-center justify-center rounded-lg',
                    enabled ? 'bg-primary/10 text-primary' : 'bg-gray-100 text-gray-400 dark:bg-ink-elevated',
                  )}
                >
                  {enabled ? <ShieldCheck className="h-[18px] w-[18px]" /> : <ShieldOff className="h-[18px] w-[18px]" />}
                </span>
                <div>
                  <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                    {enabled ? 'Proteção ativada' : 'Proteção desativada'}
                  </p>
                  <p className="text-xs text-gray-400">
                    {enabled
                      ? 'Um código será solicitado a cada login.'
                      : 'Recomendamos ativar para maior segurança.'}
                  </p>
                </div>
              </div>
              {enabled ? (
                <Button variant="outline" size="sm" onClick={() => setDisableOpen(true)}>
                  Desativar
                </Button>
              ) : (
                <Button size="sm" loading={setup.isPending} onClick={startSetup}>
                  Ativar
                </Button>
              )}
            </div>
          )
        )}
      </Card>

      {/* Disable confirmation */}
      <Modal
        open={disableOpen}
        onClose={() => setDisableOpen(false)}
        title="Desativar autenticação em dois fatores"
        description="Informe um código atual do seu aplicativo para confirmar."
        size="sm"
        footer={
          <>
            <Button variant="ghost" onClick={() => setDisableOpen(false)} disabled={disable.isPending}>
              Cancelar
            </Button>
            <Button
              variant="danger"
              onClick={confirmDisable}
              loading={disable.isPending}
              disabled={disableCode.length !== 6}
            >
              Desativar
            </Button>
          </>
        }
      >
        <Input
          label="Código de verificação"
          inputMode="numeric"
          autoComplete="one-time-code"
          maxLength={6}
          autoFocus
          placeholder="000000"
          className="text-center tracking-[0.4em]"
          value={disableCode}
          onChange={(e) => setDisableCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
        />
      </Modal>
    </>
  );
}
