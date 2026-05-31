import { useEffect, useRef, useState } from 'react';
import { Building2, Check, ChevronsUpDown } from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';
import { setActiveCurrency } from '@/lib/currency';
import { cn } from '@/lib/cn';

const roleLabels: Record<string, string> = {
  owner: 'Proprietário',
  admin: 'Administrador',
  member: 'Membro',
  viewer: 'Visualizador',
};

export function OrgSwitcher() {
  const organizations = useAuthStore((s) => s.organizations);
  const currentOrgId = useAuthStore((s) => s.currentOrgId);
  const setOrg = useAuthStore((s) => s.setOrg);
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  const current = organizations.find((o) => o.id === currentOrgId) ?? organizations[0];

  useEffect(() => {
    const onClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', onClick);
    return () => document.removeEventListener('mousedown', onClick);
  }, []);

  if (!current) return null;

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((v) => !v)}
        className="flex items-center gap-2.5 rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-left transition hover:bg-gray-50 dark:border-ink-border dark:bg-ink-surface dark:hover:bg-ink-elevated"
      >
        <span className="flex h-7 w-7 items-center justify-center rounded-md bg-primary/10 text-primary">
          <Building2 className="h-4 w-4" />
        </span>
        <span className="hidden min-w-0 sm:block">
          <span className="block max-w-[140px] truncate text-sm font-medium text-gray-900 dark:text-gray-100">
            {current.name}
          </span>
          <span className="block text-[11px] text-gray-400">
            {roleLabels[current.role] ?? current.role}
          </span>
        </span>
        <ChevronsUpDown className="h-4 w-4 text-gray-400" />
      </button>

      {open && (
        <div className="absolute left-0 z-30 mt-2 w-64 animate-fade-in overflow-hidden rounded-xl border border-gray-200 bg-white p-1.5 shadow-card dark:border-ink-border dark:bg-ink-surface">
          <p className="px-2.5 py-1.5 text-[11px] font-semibold uppercase tracking-wider text-gray-400">
            Suas organizações
          </p>
          {organizations.map((org) => (
            <button
              key={org.id}
              onClick={() => {
                setOrg(org.id);
                // Sync the active currency synchronously so amounts reformat
                // immediately on switch (org-scoped queries also refetch).
                setActiveCurrency(org.currency);
                setOpen(false);
              }}
              className={cn(
                'flex w-full items-center gap-2.5 rounded-lg px-2.5 py-2 text-left text-sm transition hover:bg-gray-100 dark:hover:bg-ink-elevated',
                org.id === current.id && 'bg-gray-50 dark:bg-ink-elevated',
              )}
            >
              <span className="flex h-7 w-7 items-center justify-center rounded-md bg-primary/10 text-primary">
                <Building2 className="h-3.5 w-3.5" />
              </span>
              <span className="min-w-0 flex-1">
                <span className="block truncate font-medium text-gray-900 dark:text-gray-100">
                  {org.name}
                </span>
                <span className="block text-[11px] text-gray-400">
                  {roleLabels[org.role] ?? org.role}
                </span>
              </span>
              {org.id === current.id && <Check className="h-4 w-4 text-primary" />}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
