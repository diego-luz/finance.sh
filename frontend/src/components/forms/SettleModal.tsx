import { useEffect, useState } from 'react';
import { Modal, Select, Input, Button } from '@/components/ui';
import { Money } from '@/components/Money';
import { useAccounts, useSettleTransaction } from '@/hooks';
import { toISODate, todayInput } from '@/lib/date';
import type { Transaction } from '@/types';

interface Props {
  open: boolean;
  onClose: () => void;
  transaction: Transaction | null;
  /** Controls labels: 'pay' for payables, 'receive' for receivables. */
  mode: 'pay' | 'receive';
}

export function SettleModal({ open, onClose, transaction, mode }: Props) {
  const { data: accounts } = useAccounts();
  const settle = useSettleTransaction();

  const [accountId, setAccountId] = useState('');
  const [paidAt, setPaidAt] = useState(todayInput());

  useEffect(() => {
    if (open) {
      setAccountId(transaction?.account_id ?? accounts?.[0]?.id ?? '');
      setPaidAt(todayInput());
    }
  }, [open, transaction, accounts]);

  const accountOptions = (accounts ?? []).map((a) => ({ value: a.id, label: a.name }));

  const onConfirm = () => {
    if (!transaction) return;
    settle.mutate(
      {
        id: transaction.id,
        payload: {
          account_id: accountId || undefined,
          paid_at: paidAt ? toISODate(paidAt) : undefined,
        },
      },
      { onSuccess: onClose },
    );
  };

  const title = mode === 'pay' ? 'Registrar pagamento' : 'Registrar recebimento';
  const cta = mode === 'pay' ? 'Pagar' : 'Receber';
  const dateLabel = mode === 'pay' ? 'Data do pagamento' : 'Data do recebimento';

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={title}
      description={transaction?.description}
      size="sm"
      footer={
        <>
          <Button variant="ghost" onClick={onClose} type="button">
            Cancelar
          </Button>
          <Button type="button" onClick={onConfirm} loading={settle.isPending}>
            {cta}
          </Button>
        </>
      }
    >
      <div className="space-y-4">
        {transaction && (
          <div className="rounded-xl border border-gray-100 bg-gray-50 px-4 py-3 dark:border-ink-border dark:bg-ink-elevated">
            <p className="text-xs text-gray-500 dark:text-gray-400">Valor</p>
            <Money
              value={transaction.amount}
              className="mt-0.5 block font-heading text-xl font-bold text-gray-900 dark:text-gray-50"
            />
          </div>
        )}

        <Select
          label="Conta"
          placeholder="Selecione uma conta"
          options={accountOptions}
          value={accountId}
          onChange={(e) => setAccountId(e.target.value)}
        />

        <Input
          label={dateLabel}
          type="date"
          value={paidAt}
          onChange={(e) => setPaidAt(e.target.value)}
        />
      </div>
    </Modal>
  );
}
