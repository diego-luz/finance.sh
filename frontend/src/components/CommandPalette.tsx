import {
  useEffect,
  useMemo,
  useRef,
  useState,
  type KeyboardEvent as ReactKeyboardEvent,
  type ReactNode,
} from 'react';
import { createPortal } from 'react-dom';
import { useNavigate } from 'react-router-dom';
import {
  Search,
  ArrowLeftRight,
  Contact as ContactIcon,
  Tags,
  Landmark,
  CreditCard,
  Target,
  Loader2,
} from 'lucide-react';
import { useSearch } from '@/hooks';
import { formatCurrency } from '@/lib/currency';
import { cn } from '@/lib/cn';
import { resolveIcon } from '@/lib/icons';
import type { SearchResponse } from '@/types';

interface Props {
  open: boolean;
  onClose: () => void;
}

/** A single flat, navigable entry derived from the grouped search response. */
interface FlatItem {
  key: string;
  /** Where to navigate when the row is chosen. */
  to: string;
  /** Visual title. */
  title: string;
  /** Optional secondary line / aside. */
  subtitle?: string;
  /** A small leading visual (icon or colored dot wrapper). */
  leading: ReactNode;
}

const groupMeta = {
  transactions: { label: 'Transações', icon: ArrowLeftRight },
  contacts: { label: 'Contatos', icon: ContactIcon },
  categories: { label: 'Categorias', icon: Tags },
  accounts: { label: 'Contas', icon: Landmark },
  credit_cards: { label: 'Cartões', icon: CreditCard },
  goals: { label: 'Metas', icon: Target },
} as const;

type GroupKey = keyof typeof groupMeta;
const GROUP_ORDER: GroupKey[] = [
  'transactions',
  'contacts',
  'categories',
  'accounts',
  'credit_cards',
  'goals',
];

function IconBox({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <span
      className={cn(
        'flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-gray-100 text-gray-500 dark:bg-ink-elevated dark:text-gray-400',
        className,
      )}
    >
      {children}
    </span>
  );
}

/** Build the ordered, grouped rows + a flat index for keyboard navigation. */
function buildGroups(res: SearchResponse | undefined) {
  const groups: { key: GroupKey; label: string; items: FlatItem[] }[] = [];
  if (!res) return { groups, flat: [] as FlatItem[] };

  const r = res.results;

  for (const key of GROUP_ORDER) {
    const items: FlatItem[] = [];
    if (key === 'transactions') {
      for (const t of r.transactions ?? []) {
        const isIncome = t.type === 'income';
        items.push({
          key: `t-${t.id}`,
          to: `/transactions?search=${encodeURIComponent(t.description)}`,
          title: t.description,
          subtitle: `${isIncome ? '+' : t.type === 'expense' ? '−' : ''} ${formatCurrency(t.amount)}${t.category ? ` · ${t.category.name}` : ''}`,
          leading: (
            <IconBox
              className={cn(
                isIncome
                  ? 'bg-primary/10 text-primary'
                  : t.type === 'transfer'
                    ? 'bg-sky-500/10 text-sky-500'
                    : 'bg-red-500/10 text-red-500',
              )}
            >
              <ArrowLeftRight className="h-4 w-4" />
            </IconBox>
          ),
        });
      }
    } else if (key === 'contacts') {
      for (const c of r.contacts ?? []) {
        items.push({
          key: `c-${c.id}`,
          to: '/contatos',
          title: c.name,
          subtitle: c.type,
          leading: (
            <IconBox>
              <ContactIcon className="h-4 w-4" />
            </IconBox>
          ),
        });
      }
    } else if (key === 'categories') {
      for (const c of r.categories ?? []) {
        const Icon = resolveIcon(c.icon);
        items.push({
          key: `cat-${c.id}`,
          to: '/categories',
          title: c.name,
          subtitle: c.kind === 'income' ? 'Receita' : 'Despesa',
          leading: (
            <span
              className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg"
              style={{ backgroundColor: `${c.color}1a`, color: c.color }}
            >
              <Icon className="h-4 w-4" />
            </span>
          ),
        });
      }
    } else if (key === 'accounts') {
      for (const a of r.accounts ?? []) {
        items.push({
          key: `a-${a.id}`,
          to: '/accounts',
          title: a.name,
          subtitle: a.type,
          leading: (
            <IconBox>
              <Landmark className="h-4 w-4" />
            </IconBox>
          ),
        });
      }
    } else if (key === 'credit_cards') {
      for (const cc of r.credit_cards ?? []) {
        items.push({
          key: `cc-${cc.id}`,
          to: '/credit-cards',
          title: cc.name,
          leading: (
            <IconBox>
              <CreditCard className="h-4 w-4" />
            </IconBox>
          ),
        });
      }
    } else if (key === 'goals') {
      for (const g of r.goals ?? []) {
        items.push({
          key: `g-${g.id}`,
          to: '/goals',
          title: g.name,
          leading: (
            <IconBox>
              <Target className="h-4 w-4" />
            </IconBox>
          ),
        });
      }
    }
    if (items.length > 0) groups.push({ key, label: groupMeta[key].label, items });
  }

  const flat = groups.flatMap((g) => g.items);
  return { groups, flat };
}

export function CommandPalette({ open, onClose }: Props) {
  const navigate = useNavigate();
  const [query, setQuery] = useState('');
  const [active, setActive] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  const { data, isFetching } = useSearch(query);
  const { groups, flat } = useMemo(() => buildGroups(data), [data]);

  const trimmed = query.trim();
  const showResults = trimmed.length >= 2;

  // Reset state whenever the palette opens; focus the input.
  useEffect(() => {
    if (open) {
      setQuery('');
      setActive(0);
      // Defer focus until after the portal mounts.
      const id = window.setTimeout(() => inputRef.current?.focus(), 0);
      document.body.style.overflow = 'hidden';
      return () => {
        window.clearTimeout(id);
        document.body.style.overflow = '';
      };
    }
  }, [open]);

  // Keep the active index in range as results change.
  useEffect(() => {
    setActive((i) => (flat.length === 0 ? 0 : Math.min(i, flat.length - 1)));
  }, [flat.length]);

  if (!open) return null;

  const go = (to: string) => {
    onClose();
    navigate(to);
  };

  const onKeyDown = (e: ReactKeyboardEvent) => {
    if (e.key === 'Escape') {
      e.preventDefault();
      onClose();
      return;
    }
    if (flat.length === 0) return;
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setActive((i) => (i + 1) % flat.length);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setActive((i) => (i - 1 + flat.length) % flat.length);
    } else if (e.key === 'Enter') {
      e.preventDefault();
      const item = flat[active];
      if (item) go(item.to);
    }
  };

  // Map flat index -> active highlighting per group.
  let runningIndex = -1;

  return createPortal(
    <div
      className="fixed inset-0 z-[60] flex items-start justify-center px-4 pt-[12vh]"
      role="dialog"
      aria-modal="true"
      aria-label="Busca global"
      onKeyDown={onKeyDown}
    >
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={onClose}
        aria-hidden
      />
      <div className="relative z-10 flex max-h-[70vh] w-full max-w-xl animate-fade-in flex-col overflow-hidden rounded-2xl bg-white shadow-card dark:bg-ink-surface">
        {/* Search input */}
        <div className="flex items-center gap-3 border-b border-gray-100 px-4 py-3 dark:border-ink-border">
          <Search className="h-5 w-5 shrink-0 text-gray-400" />
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              setActive(0);
            }}
            placeholder="Buscar transações, contatos, categorias…"
            className="flex-1 bg-transparent text-sm text-gray-900 placeholder:text-gray-400 focus:outline-none dark:text-gray-100"
            aria-label="Buscar"
          />
          {isFetching && <Loader2 className="h-4 w-4 shrink-0 animate-spin text-primary" />}
        </div>

        {/* Results */}
        <div ref={listRef} className="money-sensitive min-h-0 flex-1 overflow-y-auto py-2">
          {!showResults ? (
            <p className="px-4 py-8 text-center text-sm text-gray-400">
              Digite ao menos 2 caracteres para buscar.
            </p>
          ) : flat.length === 0 ? (
            isFetching ? (
              <div className="space-y-2 px-3">
                {Array.from({ length: 4 }).map((_, i) => (
                  <div
                    key={i}
                    className="flex items-center gap-3 rounded-lg px-2 py-2"
                  >
                    <div className="skeleton h-8 w-8 rounded-lg" />
                    <div className="flex-1 space-y-1.5">
                      <div className="skeleton h-3.5 w-1/2" />
                      <div className="skeleton h-3 w-1/3" />
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="px-4 py-8 text-center text-sm text-gray-400">Nenhum resultado</p>
            )
          ) : (
            groups.map((group) => {
              const GroupIcon = groupMeta[group.key].icon;
              return (
                <div key={group.key} className="px-2 pb-2">
                  <p className="flex items-center gap-1.5 px-2 pb-1 pt-2 text-[11px] font-semibold uppercase tracking-wider text-gray-400">
                    <GroupIcon className="h-3.5 w-3.5" />
                    {group.label}
                  </p>
                  {group.items.map((item) => {
                    runningIndex += 1;
                    const isActive = runningIndex === active;
                    const idx = runningIndex;
                    return (
                      <button
                        key={item.key}
                        type="button"
                        onMouseEnter={() => setActive(idx)}
                        onClick={() => go(item.to)}
                        className={cn(
                          'flex w-full items-center gap-3 rounded-lg px-2 py-2 text-left transition',
                          isActive
                            ? 'bg-primary/10'
                            : 'hover:bg-gray-100 dark:hover:bg-ink-elevated',
                        )}
                      >
                        {item.leading}
                        <span className="min-w-0 flex-1">
                          <span className="block truncate text-sm font-medium text-gray-900 dark:text-gray-100">
                            {item.title}
                          </span>
                          {item.subtitle && (
                            <span className="block truncate text-xs text-gray-500 dark:text-gray-400">
                              {item.subtitle}
                            </span>
                          )}
                        </span>
                      </button>
                    );
                  })}
                </div>
              );
            })
          )}
        </div>

        {/* Footer hint */}
        <div className="flex items-center justify-between border-t border-gray-100 px-4 py-2.5 text-[11px] text-gray-400 dark:border-ink-border">
          <span>↑↓ navegar · Enter abrir · Esc fechar</span>
          {data && showResults && <span>{data.total} resultado(s)</span>}
        </div>
      </div>
    </div>,
    document.body,
  );
}
