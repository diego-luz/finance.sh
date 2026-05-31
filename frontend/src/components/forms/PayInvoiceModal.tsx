import { useEffect, useState } from 'react';
import { Wallet } from 'lucide-react';
import { Modal, Select, Input, Button } from '@/components/ui';
import { Money } from '@/components/Money';
import { useAccounts, usePayInvoice } from '@/hooks';
import { toISODate, todayInput } from '@/lib/date';
import type { Invoice } from '@/types';

interface Props {
  open: boolean;
  onClose: () => void;
  cardId: string;
  invoice: Invoice | null;
}

export function PayInvoiceModal({ open, onClose, cardId, invoice }: Props) {
  const { data: accounts } = useAccounts();
  const pay = usePayInvoice();

  const [accountId, setAccountId] = useState('');
  const [paidAt, setPaidAt] = useState(todayInput());

  useEffect(() => {
    if (open) {
      setAccountId(accounts?.[0]?.id ?? '');
      setPaidAt(todayInput());
    }
  }, [open, accounts]);

  const accountOptions = (accounts ?? []).map((a) => ({ value: a.id, label: a.name }));

  const onConfirm = () => {
    if (!invoice || !accountId) return;
    pay.mutate(
      {
        cardId,
        reference: invoice.reference,
        payload: {
          account_id: accountId,
          paid_at: paidAt ? toISODate(paidAt) : undefined,
        },
      },
      { onSuccess: onClose },
    );
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Pagar fatura"
      description={invoice ? `Referência ${invoice.reference}` : undefined}
      size="sm"
      footer={
        <>
          <Button variant="ghost" type="button" onClick={onClose} disabled={pay.isPending}>
            Cancelar
          </Button>
          <Button type="button" onClick={onConfirm} loading={pay.isPending} disabled={!accountId}>
            <Wallet className="h-4 w-4" /> Confirmar pagamento
          </Button>
        </>
      }
    >
      <div className="space-y-4">
        {invoice && (
          <div className="rounded-xl border border-gray-100 bg-gray-50 px-4 py-3 dark:border-ink-border dark:bg-ink-elevated">
            <p className="text-xs text-gray-500 dark:text-gray-400">Valor em aberto</p>
            <Money
              value={invoice.open_total}
              className="mt-0.5 block font-heading text-xl font-bold text-gray-900 dark:text-gray-50"
            />
          </div>
        )}

        <Select
          label="Conta de pagamento"
          placeholder="Selecione uma conta"
          options={accountOptions}
          value={accountId}
          onChange={(e) => setAccountId(e.target.value)}
        />

        <Input
          label="Data do pagamento"
          type="date"
          value={paidAt}
          onChange={(e) => setPaidAt(e.target.value)}
        />
      </div>
    </Modal>
  );
}
