import { useMemo, useState } from 'react';
import { Plus, Tag as TagIcon, Pencil, Trash2 } from 'lucide-react';
import { PageHeader } from '@/components/PageHeader';
import { Button, Card, EmptyState, Skeleton } from '@/components/ui';
import { TagFormModal } from '@/components/forms/TagFormModal';
import { useTags, useDeleteTag, useIsViewer } from '@/hooks';
import type { Tag } from '@/types';

function TagRow({
  tag,
  readOnly,
  onEdit,
  onDelete,
}: {
  tag: Tag;
  readOnly: boolean;
  onEdit: () => void;
  onDelete: () => void;
}) {
  return (
    <div className="group flex items-center gap-3 rounded-xl border border-gray-100 bg-white p-3 transition hover:border-gray-200 hover:shadow-soft dark:border-ink-border dark:bg-ink-surface dark:hover:border-gray-600">
      <span
        className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-sm font-medium"
        style={{ backgroundColor: `${tag.color}1a`, color: tag.color }}
      >
        <span className="h-2 w-2 rounded-full" style={{ backgroundColor: tag.color }} />
        {tag.name}
      </span>
      {!readOnly && (
        <div className="ml-auto flex items-center gap-0.5 opacity-0 transition group-hover:opacity-100">
          <button
            onClick={onEdit}
            className="rounded-md p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-ink-elevated"
            aria-label="Editar tag"
          >
            <Pencil className="h-4 w-4" />
          </button>
          <button
            onClick={onDelete}
            className="rounded-md p-1.5 text-gray-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-500/10"
            aria-label="Excluir tag"
          >
            <Trash2 className="h-4 w-4" />
          </button>
        </div>
      )}
    </div>
  );
}

export function TagsPage() {
  const { data: tags, isLoading } = useTags();
  const deleteTag = useDeleteTag();
  const readOnly = useIsViewer();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Tag | null>(null);

  const list = useMemo(() => tags ?? [], [tags]);

  const openCreate = () => {
    setEditing(null);
    setModalOpen(true);
  };
  const openEdit = (t: Tag) => {
    setEditing(t);
    setModalOpen(true);
  };
  const handleDelete = (t: Tag) => {
    if (window.confirm(`Excluir a tag "${t.name}"?`)) {
      deleteTag.mutate(t.id);
    }
  };

  return (
    <div>
      <PageHeader
        eyebrow="Cadastros"
        title="Tags"
        description="Marque transações com etiquetas para filtros e relatórios transversais."
        action={
          !readOnly && (
            <Button onClick={openCreate}>
              <Plus className="h-4 w-4" /> Nova tag
            </Button>
          )
        }
      />

      {isLoading ? (
        <Card>
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
            {Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="h-12 rounded-xl" />
            ))}
          </div>
        </Card>
      ) : list.length === 0 ? (
        <EmptyState
          icon={TagIcon}
          title="Nenhuma tag cadastrada"
          description="Crie tags para marcar transações e cruzar informações entre categorias."
          action={
            !readOnly && (
              <Button onClick={openCreate}>
                <Plus className="h-4 w-4" /> Criar tag
              </Button>
            )
          }
        />
      ) : (
        <Card>
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
            {list.map((t) => (
              <TagRow
                key={t.id}
                tag={t}
                readOnly={readOnly}
                onEdit={() => openEdit(t)}
                onDelete={() => handleDelete(t)}
              />
            ))}
          </div>
        </Card>
      )}

      <TagFormModal open={modalOpen} onClose={() => setModalOpen(false)} tag={editing} />
    </div>
  );
}
