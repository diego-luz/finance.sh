import { useState } from 'react';
import { Check, Plus, X } from 'lucide-react';
import { cn } from '@/lib/cn';
import { useTags, useCreateTag } from '@/hooks';
import { TagFormModal } from './TagFormModal';
import type { Tag } from '@/types';

interface Props {
  /** Currently selected tag ids. */
  value: string[];
  onChange: (ids: string[]) => void;
  label?: string;
}

/**
 * Compact multi-select for tags rendered as toggleable colored chips. Lets the
 * user quick-create a new tag inline (by typing a name and pressing Enter) or
 * open the full TagFormModal to pick a color. Newly created tags are selected
 * automatically.
 */
export function TagPicker({ value, onChange, label = 'Tags' }: Props) {
  const { data: tags } = useTags();
  const createTag = useCreateTag();
  const [inlineName, setInlineName] = useState('');
  const [modalOpen, setModalOpen] = useState(false);

  const list = tags ?? [];
  const selected = new Set(value);

  const toggle = (id: string) => {
    const next = new Set(selected);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    onChange(Array.from(next));
  };

  const selectTag = (tag: Tag) => {
    if (selected.has(tag.id)) return;
    onChange([...value, tag.id]);
  };

  const quickCreate = () => {
    const name = inlineName.trim();
    if (!name || createTag.isPending) return;
    createTag.mutate(
      { name },
      {
        onSuccess: (tag) => {
          selectTag(tag);
          setInlineName('');
        },
      },
    );
  };

  return (
    <div>
      <span className="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
        {label}
      </span>

      {list.length > 0 && (
        <div className="mb-2 flex flex-wrap gap-1.5">
          {list.map((tag) => {
            const isSelected = selected.has(tag.id);
            return (
              <button
                key={tag.id}
                type="button"
                onClick={() => toggle(tag.id)}
                className={cn(
                  'inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-medium transition',
                  isSelected
                    ? 'border-transparent'
                    : 'border-gray-200 text-gray-600 hover:bg-gray-50 dark:border-ink-border dark:text-gray-300 dark:hover:bg-ink-elevated',
                )}
                style={
                  isSelected
                    ? { backgroundColor: `${tag.color}1a`, color: tag.color }
                    : undefined
                }
                aria-pressed={isSelected}
              >
                <span className="h-2 w-2 rounded-full" style={{ backgroundColor: tag.color }} />
                {tag.name}
                {isSelected && <Check className="h-3 w-3" />}
              </button>
            );
          })}
        </div>
      )}

      <div className="flex items-center gap-2">
        <input
          type="text"
          value={inlineName}
          onChange={(e) => setInlineName(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') {
              e.preventDefault();
              quickCreate();
            }
          }}
          placeholder="Nova tag…"
          className="input-base h-9 flex-1 text-sm"
          aria-label="Nome da nova tag"
        />
        {inlineName.trim() ? (
          <button
            type="button"
            onClick={quickCreate}
            disabled={createTag.isPending}
            className="inline-flex h-9 shrink-0 items-center gap-1 rounded-lg bg-primary/10 px-3 text-xs font-medium text-primary transition hover:bg-primary/20 disabled:opacity-50"
          >
            <Plus className="h-3.5 w-3.5" /> Criar
          </button>
        ) : (
          <button
            type="button"
            onClick={() => setModalOpen(true)}
            className="inline-flex h-9 shrink-0 items-center gap-1 rounded-lg border border-gray-200 px-3 text-xs font-medium text-gray-600 transition hover:bg-gray-50 dark:border-ink-border dark:text-gray-300 dark:hover:bg-ink-elevated"
          >
            <Plus className="h-3.5 w-3.5" /> Nova tag
          </button>
        )}
        {value.length > 0 && (
          <button
            type="button"
            onClick={() => onChange([])}
            className="inline-flex h-9 shrink-0 items-center gap-1 rounded-lg px-2 text-xs text-gray-400 transition hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-ink-elevated"
            aria-label="Limpar tags"
          >
            <X className="h-3.5 w-3.5" /> Limpar
          </button>
        )}
      </div>

      <TagFormModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onSaved={(tag) => selectTag(tag)}
      />
    </div>
  );
}
