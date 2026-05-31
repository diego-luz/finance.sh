import { useState } from 'react';
import {
  Plus,
  CreditCard as CreditCardIcon,
  MoreVertical,
  Pencil,
  Trash2,
  Receipt,
} from 'lucide-react';
import { PageHeader } from '@/components/PageHeader';
import { Button, Card, EmptyState, SkeletonCard } from '@/components/ui';
import { CreditCardFormModal } from '@/components/forms/CreditCardFormModal';
import { InvoicesModal } from '@/components/forms/InvoicesModal';
import { Money } from '@/components/Money';
import { useCreditCards, useDeleteCreditCard, useIsViewer } from '@/hooks';
import { cn } from '@/lib/cn';
import type { CreditCard } from '@/types';

function CreditCardItem({
  card,
  readOnly,
  onEdit,
  onDelete,
  onInvoices,
}: {
  card: CreditCard;
  readOnly: boolean;
  onEdit: () => void;
  onDelete: () => void;
  onInvoices: () => void;
}) {
  const [menu, setMenu] = useState(false);
  const percent = card.limit > 0 ? Math.min(100, Math.round((card.used / card.limit) * 100)) : 0;
  const over = card.used > card.limit;

  return (
    <Card className="relative overflow-hidden">
      <div
        className="absolute inset-x-0 top-0 h-1"
        style={{ backgroundColor: card.color }}
        aria-hidden
      />
      <div className="flex items-start justify-between">
        <span
          className="flex h-12 w-12 items-center justify-center rounded-xl"
          style={{ backgroundColor: `${card.color}1a`, color: card.color }}
        >
          <CreditCardIcon className="h-6 w-6" />
        </span>

        {!readOnly && (
          <div className="relative">
            <button
              onClick={() => setMenu((v) => !v)}
              onBlur={() => setTimeout(() => setMenu(false), 150)}
              className="rounded-lg p-1.5 text-gray-400 transition hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-ink-elevated"
              aria-label="Ações do cartão"
            >
              <MoreVertical className="h-4 w-4" />
            </button>
            {menu && (
              <div className="absolute right-0 z-10 mt-1 w-40 animate-fade-in overflow-hidden rounded-lg border border-gray-200 bg-white p-1 shadow-card dark:border-ink-border dark:bg-ink-surface">
                <button
                  onClick={onEdit}
                  className="flex w-full items-center gap-2 rounded-md px-2.5 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-ink-elevated"
                >
                  <Pencil className="h-4 w-4" /> Editar
                </button>
                <button
                  onClick={onDelete}
                  className="flex w-full items-center gap-2 rounded-md px-2.5 py-2 text-sm text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-500/10"
                >
                  <Trash2 className="h-4 w-4" /> Excluir
                </button>
              </div>
            )}
          </div>
        )}
      </div>

      <div className="mt-4">
        <h3 className="font-heading text-base font-semibold text-gray-900 dark:text-gray-100">
          {card.name}
        </h3>
        <p className="mt-0.5 text-xs text-gray-400">
          Fecha dia {card.closing_day} · vence dia {card.due_day}
        </p>
      </div>

      <div className="mt-4">
        <div className="mb-1.5 flex items-center justify-between text-xs">
          <span className="text-gray-500 dark:text-gray-400">
            Usado <Money value={card.used} />
          </span>
          <span className={cn('font-medium', over ? 'text-red-500' : 'text-gray-500 dark:text-gray-400')}>
            {percent}%
          </span>
        </div>
        <div className="h-2 w-full overflow-hidden rounded-full bg-gray-100 dark:bg-ink-elevated">
          <div
            className="h-full rounded-full transition-all"
            style={{
              width: `${percent}%`,
              backgroundColor: over ? '#ef4444' : card.color,
            }}
          />
        </div>
        <div className="mt-3 flex items-center justify-between">
          <div>
            <p className="text-[11px] uppercase tracking-wide text-gray-400">Disponível</p>
            <p className="font-heading text-lg font-bold tabular-nums text-gray-900 dark:text-gray-50">
              <Money value={card.available} />
            </p>
          </div>
          <div className="text-right">
            <p className="text-[11px] uppercase tracking-wide text-gray-400">Limite</p>
            <p className="text-sm font-medium tabular-nums text-gray-600 dark:text-gray-300">
              <Money value={card.limit} />
            </p>
          </div>
        </div>
      </div>

      <div className="mt-4 border-t border-gray-100 pt-3 dark:border-ink-border">
        <Button variant="outline" size="sm" className="w-full" onClick={onInvoices}>
          <Receipt className="h-4 w-4" /> Ver faturas
        </Button>
      </div>
    </Card>
  );
}

export function CreditCardsPage() {
  const { data: cards, isLoading } = useCreditCards();
  const deleteCard = useDeleteCreditCard();
  const readOnly = useIsViewer();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<CreditCard | null>(null);
  const [invoicesOpen, setInvoicesOpen] = useState(false);
  const [invoicesCard, setInvoicesCard] = useState<CreditCard | null>(null);

  const openInvoices = (card: CreditCard) => {
    setInvoicesCard(card);
    setInvoicesOpen(true);
  };

  const openCreate = () => {
    setEditing(null);
    setModalOpen(true);
  };
  const openEdit = (card: CreditCard) => {
    setEditing(card);
    setModalOpen(true);
  };
  const handleDelete = (card: CreditCard) => {
    if (window.confirm(`Excluir o cartão "${card.name}"? Esta ação não pode ser desfeita.`)) {
      deleteCard.mutate(card.id);
    }
  };

  return (
    <div>
      <PageHeader
        eyebrow="Cadastros"
        title="Cartões"
        description="Acompanhe o limite, uso e disponibilidade dos seus cartões de crédito."
        action={
          !readOnly && (
            <Button onClick={openCreate}>
              <Plus className="h-4 w-4" /> Novo cartão
            </Button>
          )
        }
      />

      {isLoading ? (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <SkeletonCard key={i} />
          ))}
        </div>
      ) : !cards || cards.length === 0 ? (
        <EmptyState
          icon={CreditCardIcon}
          title="Nenhum cartão cadastrado"
          description="Cadastre seus cartões de crédito para acompanhar limites e faturas."
          action={
            !readOnly && (
              <Button onClick={openCreate}>
                <Plus className="h-4 w-4" /> Criar cartão
              </Button>
            )
          }
        />
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {cards.map((card) => (
            <CreditCardItem
              key={card.id}
              card={card}
              readOnly={readOnly}
              onEdit={() => openEdit(card)}
              onDelete={() => handleDelete(card)}
              onInvoices={() => openInvoices(card)}
            />
          ))}
        </div>
      )}

      <CreditCardFormModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        creditCard={editing}
      />

      <InvoicesModal
        open={invoicesOpen}
        onClose={() => setInvoicesOpen(false)}
        card={invoicesCard}
      />
    </div>
  );
}
