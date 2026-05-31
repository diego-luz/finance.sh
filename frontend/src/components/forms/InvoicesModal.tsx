import { useEffect, useState } from 'react';
import {
  ChevronDown,
  ChevronRight,
  FileText,
  Receipt,
  Wallet,
} from 'lucide-react';
import { Modal, Badge, Button, EmptyState, Skeleton } from '@/components/ui';
import type { BadgeVariant } from '@/components/ui';
import { Money } from '@/components/Money';
import { useInvoices, useInvoice, useIsViewer } from '@/hooks';
import { formatDate, formatDateShort } from '@/lib/date';
import { cn } from '@/lib/cn';
import type { CreditCard, Invoice, InvoiceStatus } from '@/types';
import { PayInvoiceModal } from './PayInvoiceModal';

const STATUS_META: Record<InvoiceStatus, { label: string; variant: BadgeVariant }> = {
  open: { label: 'Aberta', variant: 'info' },
  closed: { label: 'Fechada', variant: 'warning' },
  paid: { label: 'Paga', variant: 'success' },
  overdue: { label: 'Vencida', variant: 'danger' },
};

/** Human-friendly "Maio 2026" from a "YYYY-MM" reference. */
function formatReference(reference: string): string {
  const label = formatDate(`${reference}-01`, 'MMMM yyyy');
  if (!label) return reference;
  return label.charAt(0).toUpperCase() + label.slice(1);
}

interface Props {
  open: boolean;
  onClose: () => void;
  card: CreditCard | null;
}

export function InvoicesModal({ open, onClose, card }: Props) {
  const cardId = card?.id ?? '';
  const { data: invoices, isLoading, isError } = useInvoices(cardId, open && Boolean(cardId));
  const [expanded, setExpanded] = useState<string | null>(null);
  const [payTarget, setPayTarget] = useState<Invoice | null>(null);

  useEffect(() => {
    if (open) {
      setExpanded(null);
      setPayTarget(null);
    }
  }, [open, cardId]);

  return (
    <>
      <Modal
        open={open}
        onClose={onClose}
        variant="drawer"
        title={card ? `Faturas · ${card.name}` : 'Faturas'}
        description="Acompanhe e pague as faturas do cartão."
      >
        {isLoading ? (
          <div className="space-y-3">
            {Array.from({ length: 4 }).map((_, i) => (
              <div
                key={i}
                className="space-y-2 rounded-xl border border-gray-100 p-4 dark:border-ink-border"
              >
                <Skeleton className="h-3 w-1/3" />
                <Skeleton className="h-6 w-1/2" />
                <Skeleton className="h-3 w-2/3" />
              </div>
            ))}
          </div>
        ) : isError ? (
          <EmptyState
            icon={Receipt}
            title="Não foi possível carregar as faturas"
            description="Tente novamente em instantes."
          />
        ) : !invoices || invoices.length === 0 ? (
          <EmptyState
            icon={Receipt}
            title="Nenhuma fatura encontrada"
            description="As faturas aparecem aqui conforme você registra despesas neste cartão."
          />
        ) : (
          <div className="space-y-3">
            {invoices.map((invoice) => (
              <InvoiceRow
                key={invoice.reference}
                cardId={cardId}
                invoice={invoice}
                expanded={expanded === invoice.reference}
                onToggle={() =>
                  setExpanded((prev) =>
                    prev === invoice.reference ? null : invoice.reference,
                  )
                }
                onPay={() => setPayTarget(invoice)}
              />
            ))}
          </div>
        )}
      </Modal>

      <PayInvoiceModal
        open={Boolean(payTarget)}
        onClose={() => setPayTarget(null)}
        cardId={cardId}
        invoice={payTarget}
      />
    </>
  );
}

function InvoiceRow({
  cardId,
  invoice,
  expanded,
  onToggle,
  onPay,
}: {
  cardId: string;
  invoice: Invoice;
  expanded: boolean;
  onToggle: () => void;
  onPay: () => void;
}) {
  const readOnly = useIsViewer();
  const status = STATUS_META[invoice.status];
  const payable = invoice.status !== 'paid' && invoice.open_total > 0;

  const { data: detail, isLoading: detailLoading } = useInvoice(
    cardId,
    invoice.reference,
    expanded,
  );

  return (
    <div
      className={cn(
        'overflow-hidden rounded-xl border transition',
        invoice.status === 'overdue'
          ? 'border-red-300 dark:border-red-500/40'
          : 'border-gray-200 dark:border-ink-border',
      )}
    >
      <button
        type="button"
        onClick={onToggle}
        className="flex w-full items-center gap-3 px-4 py-3 text-left transition hover:bg-gray-50 dark:hover:bg-ink-elevated"
        aria-expanded={expanded}
      >
        <span className="text-gray-400">
          {expanded ? (
            <ChevronDown className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
        </span>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <p className="truncate font-medium text-gray-900 dark:text-gray-100">
              {formatReference(invoice.reference)}
            </p>
            <Badge variant={status.variant} dot>
              {status.label}
            </Badge>
          </div>
          <p className="mt-0.5 text-xs text-gray-400">
            {formatDateShort(invoice.period_start)} – {formatDateShort(invoice.period_end)} ·
            vence {formatDateShort(invoice.due_date)}
          </p>
        </div>
        <div className="text-right">
          <Money
            value={invoice.total}
            className="block font-heading text-base font-bold text-gray-900 dark:text-gray-50"
          />
          {invoice.open_total > 0 && invoice.open_total !== invoice.total && (
            <p className="text-xs text-gray-400">
              em aberto <Money value={invoice.open_total} />
            </p>
          )}
        </div>
      </button>

      {expanded && (
        <div className="border-t border-gray-100 px-4 py-3 dark:border-ink-border">
          <div className="mb-3 grid grid-cols-3 gap-2 text-center">
            <div className="rounded-lg bg-gray-50 px-2 py-2 dark:bg-ink-elevated">
              <p className="text-[11px] uppercase tracking-wide text-gray-400">Total</p>
              <Money
                value={invoice.total}
                className="block text-sm font-semibold text-gray-900 dark:text-gray-100"
              />
            </div>
            <div className="rounded-lg bg-gray-50 px-2 py-2 dark:bg-ink-elevated">
              <p className="text-[11px] uppercase tracking-wide text-gray-400">Pago</p>
              <Money
                value={invoice.paid_total}
                className="block text-sm font-semibold text-primary"
              />
            </div>
            <div className="rounded-lg bg-gray-50 px-2 py-2 dark:bg-ink-elevated">
              <p className="text-[11px] uppercase tracking-wide text-gray-400">Em aberto</p>
              <Money
                value={invoice.open_total}
                className="block text-sm font-semibold text-gray-900 dark:text-gray-100"
              />
            </div>
          </div>

          {detailLoading ? (
            <div className="space-y-2">
              {Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-9" />
              ))}
            </div>
          ) : !detail || detail.transactions.length === 0 ? (
            <div className="flex items-center gap-2 py-2 text-sm text-gray-400">
              <FileText className="h-4 w-4" /> Sem lançamentos nesta fatura.
            </div>
          ) : (
            <ul className="divide-y divide-gray-100 dark:divide-ink-border">
              {detail.transactions.map((t) => (
                <li key={t.id} className="flex items-center justify-between gap-3 py-2">
                  <div className="min-w-0">
                    <div className="flex items-center gap-1.5">
                      <p className="truncate text-sm text-gray-700 dark:text-gray-200">
                        {t.description}
                      </p>
                      {(t.installment_total ?? 0) > 1 && (
                        <Badge variant="info">
                          {t.installment_number}/{t.installment_total}
                        </Badge>
                      )}
                    </div>
                    <p className="text-xs text-gray-400">{formatDateShort(t.date)}</p>
                  </div>
                  <Money
                    value={t.amount}
                    className="shrink-0 text-sm font-semibold text-red-500"
                  />
                </li>
              ))}
            </ul>
          )}

          {!readOnly && payable && (
            <div className="mt-3 flex justify-end">
              <Button size="sm" onClick={onPay}>
                <Wallet className="h-4 w-4" /> Pagar fatura
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
