import { useMemo, useState } from 'react';
import {
  Plus,
  Users,
  Pencil,
  Trash2,
  Search,
  Mail,
  Phone,
  FileText,
} from 'lucide-react';
import { PageHeader } from '@/components/PageHeader';
import { Button, Card, Input, EmptyState, Badge, Skeleton } from '@/components/ui';
import type { BadgeVariant } from '@/components/ui';
import { ContactFormModal } from '@/components/forms/ContactFormModal';
import { useContacts, useDeleteContact, useIsViewer } from '@/hooks';
import type { Contact, ContactType } from '@/types';

const typeMeta: Record<ContactType, { label: string; variant: BadgeVariant }> = {
  customer: { label: 'Cliente', variant: 'info' },
  supplier: { label: 'Fornecedor', variant: 'warning' },
  both: { label: 'Cliente e fornecedor', variant: 'neutral' },
};

function ContactCard({
  contact,
  readOnly,
  onEdit,
  onDelete,
}: {
  contact: Contact;
  readOnly: boolean;
  onEdit: () => void;
  onDelete: () => void;
}) {
  const meta = typeMeta[contact.type];
  return (
    <div className="group flex flex-col rounded-xl border border-gray-100 bg-white p-4 transition hover:border-gray-200 hover:shadow-soft dark:border-ink-border dark:bg-ink-surface dark:hover:border-gray-600">
      <div className="flex items-start justify-between gap-3">
        <div className="flex min-w-0 items-center gap-3">
          <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
            <Users className="h-5 w-5" />
          </span>
          <div className="min-w-0">
            <p className="truncate text-sm font-semibold text-gray-900 dark:text-gray-100">
              {contact.name}
            </p>
            <Badge variant={meta.variant}>{meta.label}</Badge>
          </div>
        </div>
        {!readOnly && (
          <div className="flex items-center gap-0.5 opacity-0 transition group-hover:opacity-100">
            <button
              onClick={onEdit}
              className="rounded-md p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-ink-elevated"
              aria-label="Editar contato"
            >
              <Pencil className="h-4 w-4" />
            </button>
            <button
              onClick={onDelete}
              className="rounded-md p-1.5 text-gray-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-500/10"
              aria-label="Excluir contato"
            >
              <Trash2 className="h-4 w-4" />
            </button>
          </div>
        )}
      </div>

      {(contact.email || contact.phone || contact.document) && (
        <div className="mt-3 space-y-1.5 border-t border-gray-100 pt-3 text-xs text-gray-500 dark:border-ink-border dark:text-gray-400">
          {contact.document && (
            <p className="flex items-center gap-1.5">
              <FileText className="h-3.5 w-3.5 shrink-0" />
              <span className="truncate">{contact.document}</span>
            </p>
          )}
          {contact.email && (
            <p className="flex items-center gap-1.5">
              <Mail className="h-3.5 w-3.5 shrink-0" />
              <span className="truncate">{contact.email}</span>
            </p>
          )}
          {contact.phone && (
            <p className="flex items-center gap-1.5">
              <Phone className="h-3.5 w-3.5 shrink-0" />
              <span className="truncate">{contact.phone}</span>
            </p>
          )}
        </div>
      )}
    </div>
  );
}

export function ContactsPage() {
  const { data: contacts, isLoading } = useContacts();
  const deleteContact = useDeleteContact();
  const readOnly = useIsViewer();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Contact | null>(null);
  const [search, setSearch] = useState('');

  const filtered = useMemo(() => {
    const list = contacts ?? [];
    const q = search.trim().toLowerCase();
    if (!q) return list;
    return list.filter(
      (c) =>
        c.name.toLowerCase().includes(q) ||
        c.email?.toLowerCase().includes(q) ||
        c.document?.toLowerCase().includes(q),
    );
  }, [contacts, search]);

  const openCreate = () => {
    setEditing(null);
    setModalOpen(true);
  };
  const openEdit = (c: Contact) => {
    setEditing(c);
    setModalOpen(true);
  };
  const handleDelete = (c: Contact) => {
    if (window.confirm(`Excluir o contato "${c.name}"?`)) {
      deleteContact.mutate(c.id);
    }
  };

  return (
    <div>
      <PageHeader
        eyebrow="Cadastros"
        title="Contatos"
        description="Cadastre clientes e fornecedores para vincular às suas contas."
        action={
          !readOnly && (
            <Button onClick={openCreate}>
              <Plus className="h-4 w-4" /> Novo contato
            </Button>
          )
        }
      />

      <Card className="mb-4 p-4">
        <Input
          placeholder="Buscar por nome, e-mail ou documento..."
          leftIcon={<Search className="h-4 w-4" />}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </Card>

      {isLoading ? (
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Card key={i}>
              <Skeleton className="mb-3 h-10 w-10 rounded-lg" />
              <Skeleton className="mb-2 h-4 w-2/3" />
              <Skeleton className="h-3 w-1/2" />
            </Card>
          ))}
        </div>
      ) : filtered.length === 0 ? (
        <EmptyState
          icon={Users}
          title={search ? 'Nenhum contato encontrado' : 'Nenhum contato cadastrado'}
          description={
            search
              ? 'Ajuste a busca para encontrar o contato.'
              : 'Cadastre clientes e fornecedores para usar em contas a pagar e receber.'
          }
          action={
            !readOnly &&
            !search && (
              <Button onClick={openCreate}>
                <Plus className="h-4 w-4" /> Criar contato
              </Button>
            )
          }
        />
      ) : (
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {filtered.map((c) => (
            <ContactCard
              key={c.id}
              contact={c}
              readOnly={readOnly}
              onEdit={() => openEdit(c)}
              onDelete={() => handleDelete(c)}
            />
          ))}
        </div>
      )}

      <ContactFormModal open={modalOpen} onClose={() => setModalOpen(false)} contact={editing} />
    </div>
  );
}
