import { useState } from 'react';
import { Plus, Landmark, MoreVertical, Pencil, Trash2, Archive } from 'lucide-react';
import { PageHeader } from '@/components/PageHeader';
import { Button, Card, EmptyState, SkeletonCard, Badge } from '@/components/ui';
import { AccountFormModal } from '@/components/forms/AccountFormModal';
import { Money } from '@/components/Money';
import { useAccounts, useDeleteAccount, useIsViewer } from '@/hooks';
import { resolveIcon, accountTypeMeta } from '@/lib/icons';
import type { Account } from '@/types';

function AccountCard({
  account,
  readOnly,
  onEdit,
  onDelete,
}: {
  account: Account;
  readOnly: boolean;
  onEdit: () => void;
  onDelete: () => void;
}) {
  const [menu, setMenu] = useState(false);
  const Icon = resolveIcon(account.icon);

  return (
    <Card className="relative overflow-hidden">
      <div
        className="absolute inset-x-0 top-0 h-1"
        style={{ backgroundColor: account.color }}
        aria-hidden
      />
      <div className="flex items-start justify-between">
        <span
          className="flex h-12 w-12 items-center justify-center rounded-xl"
          style={{ backgroundColor: `${account.color}1a`, color: account.color }}
        >
          <Icon className="h-6 w-6" />
        </span>

        {!readOnly && (
          <div className="relative">
            <button
              onClick={() => setMenu((v) => !v)}
              onBlur={() => setTimeout(() => setMenu(false), 150)}
              className="rounded-lg p-1.5 text-gray-400 transition hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-ink-elevated"
              aria-label="Ações da conta"
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
        <div className="flex items-center gap-2">
          <h3 className="font-heading text-base font-semibold text-gray-900 dark:text-gray-100">
            {account.name}
          </h3>
          {account.archived && (
            <Badge variant="neutral">
              <Archive className="h-3 w-3" /> Arquivada
            </Badge>
          )}
        </div>
        <Badge variant="success" className="mt-1.5">
          {accountTypeMeta[account.type]?.label ?? account.type}
        </Badge>
      </div>

      <div className="mt-4">
        <p className="text-xs text-gray-400">Saldo atual</p>
        <p className="font-heading text-2xl font-bold tabular-nums text-gray-900 dark:text-gray-50">
          <Money value={account.balance} />
        </p>
      </div>
    </Card>
  );
}

export function AccountsPage() {
  const { data: accounts, isLoading } = useAccounts();
  const deleteAccount = useDeleteAccount();
  const readOnly = useIsViewer();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Account | null>(null);

  const openCreate = () => {
    setEditing(null);
    setModalOpen(true);
  };
  const openEdit = (account: Account) => {
    setEditing(account);
    setModalOpen(true);
  };
  const handleDelete = (account: Account) => {
    if (window.confirm(`Excluir a conta "${account.name}"? Esta ação não pode ser desfeita.`)) {
      deleteAccount.mutate(account.id);
    }
  };

  return (
    <div>
      <PageHeader
        eyebrow="Cadastros"
        title="Contas"
        description="Gerencie suas contas bancárias, carteiras e cartões."
        action={
          !readOnly && (
            <Button onClick={openCreate}>
              <Plus className="h-4 w-4" /> Nova conta
            </Button>
          )
        }
      />

      {isLoading ? (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <SkeletonCard key={i} />
          ))}
        </div>
      ) : !accounts || accounts.length === 0 ? (
        <EmptyState
          icon={Landmark}
          title="Nenhuma conta cadastrada"
          description="Cadastre sua primeira conta para começar a controlar suas finanças."
          action={
            !readOnly && (
              <Button onClick={openCreate}>
                <Plus className="h-4 w-4" /> Criar conta
              </Button>
            )
          }
        />
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {accounts.map((account) => (
            <AccountCard
              key={account.id}
              account={account}
              readOnly={readOnly}
              onEdit={() => openEdit(account)}
              onDelete={() => handleDelete(account)}
            />
          ))}
        </div>
      )}

      <AccountFormModal open={modalOpen} onClose={() => setModalOpen(false)} account={editing} />
    </div>
  );
}
