import { useMemo, useState } from 'react';
import { Plus, Tags, Pencil, Trash2, ArrowDownCircle, ArrowUpCircle } from 'lucide-react';
import { PageHeader } from '@/components/PageHeader';
import { Button, Card, EmptyState, Skeleton } from '@/components/ui';
import { CategoryFormModal } from '@/components/forms/CategoryFormModal';
import { useCategories, useDeleteCategory, useIsViewer } from '@/hooks';
import { resolveIcon } from '@/lib/icons';
import type { Category, CategoryKind } from '@/types';

function CategoryChip({
  category,
  readOnly,
  onEdit,
  onDelete,
}: {
  category: Category;
  readOnly: boolean;
  onEdit: () => void;
  onDelete: () => void;
}) {
  const Icon = resolveIcon(category.icon);
  return (
    <div className="group flex items-center gap-3 rounded-xl border border-gray-100 bg-white p-3 transition hover:border-gray-200 hover:shadow-soft dark:border-ink-border dark:bg-ink-surface dark:hover:border-gray-600">
      <span
        className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg"
        style={{ backgroundColor: `${category.color}1a`, color: category.color }}
      >
        <Icon className="h-5 w-5" />
      </span>
      <span className="min-w-0 flex-1 truncate text-sm font-medium text-gray-900 dark:text-gray-100">
        {category.name}
      </span>
      {!readOnly && (
        <div className="flex items-center gap-0.5 opacity-0 transition group-hover:opacity-100">
          <button
            onClick={onEdit}
            className="rounded-md p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-ink-elevated"
            aria-label="Editar categoria"
          >
            <Pencil className="h-4 w-4" />
          </button>
          <button
            onClick={onDelete}
            className="rounded-md p-1.5 text-gray-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-500/10"
            aria-label="Excluir categoria"
          >
            <Trash2 className="h-4 w-4" />
          </button>
        </div>
      )}
    </div>
  );
}

function CategorySection({
  title,
  kind,
  categories,
  readOnly,
  onEdit,
  onDelete,
}: {
  title: string;
  kind: CategoryKind;
  categories: Category[];
  readOnly: boolean;
  onEdit: (c: Category) => void;
  onDelete: (c: Category) => void;
}) {
  const items = categories.filter((c) => c.kind === kind);
  const Icon = kind === 'income' ? ArrowUpCircle : ArrowDownCircle;
  return (
    <Card>
      <div className="mb-4 flex items-center gap-2">
        <Icon className={kind === 'income' ? 'h-5 w-5 text-primary' : 'h-5 w-5 text-red-500'} />
        <h3 className="font-heading text-base font-semibold text-gray-900 dark:text-gray-100">
          {title}
        </h3>
        <span className="ml-auto text-xs text-gray-400">{items.length}</span>
      </div>
      {items.length === 0 ? (
        <p className="py-6 text-center text-sm text-gray-400">Nenhuma categoria deste tipo.</p>
      ) : (
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
          {items.map((c) => (
            <CategoryChip
              key={c.id}
              category={c}
              readOnly={readOnly}
              onEdit={() => onEdit(c)}
              onDelete={() => onDelete(c)}
            />
          ))}
        </div>
      )}
    </Card>
  );
}

export function CategoriesPage() {
  const { data: categories, isLoading } = useCategories();
  const deleteCategory = useDeleteCategory();
  const readOnly = useIsViewer();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Category | null>(null);

  const list = useMemo(() => categories ?? [], [categories]);

  const openCreate = () => {
    setEditing(null);
    setModalOpen(true);
  };
  const openEdit = (c: Category) => {
    setEditing(c);
    setModalOpen(true);
  };
  const handleDelete = (c: Category) => {
    if (window.confirm(`Excluir a categoria "${c.name}"?`)) {
      deleteCategory.mutate(c.id);
    }
  };

  return (
    <div>
      <PageHeader
        eyebrow="Cadastros"
        title="Categorias"
        description="Classifique suas receitas e despesas para relatórios mais ricos."
        action={
          !readOnly && (
            <Button onClick={openCreate}>
              <Plus className="h-4 w-4" /> Nova categoria
            </Button>
          )
        }
      />

      {isLoading ? (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          {Array.from({ length: 2 }).map((_, i) => (
            <Card key={i}>
              <Skeleton className="mb-4 h-5 w-32" />
              <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                {Array.from({ length: 4 }).map((__, j) => (
                  <Skeleton key={j} className="h-16 rounded-xl" />
                ))}
              </div>
            </Card>
          ))}
        </div>
      ) : list.length === 0 ? (
        <EmptyState
          icon={Tags}
          title="Nenhuma categoria cadastrada"
          description="Crie categorias para organizar suas transações por tipo."
          action={
            !readOnly && (
              <Button onClick={openCreate}>
                <Plus className="h-4 w-4" /> Criar categoria
              </Button>
            )
          }
        />
      ) : (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          <CategorySection
            title="Despesas"
            kind="expense"
            categories={list}
            readOnly={readOnly}
            onEdit={openEdit}
            onDelete={handleDelete}
          />
          <CategorySection
            title="Receitas"
            kind="income"
            categories={list}
            readOnly={readOnly}
            onEdit={openEdit}
            onDelete={handleDelete}
          />
        </div>
      )}

      <CategoryFormModal open={modalOpen} onClose={() => setModalOpen(false)} category={editing} />
    </div>
  );
}
