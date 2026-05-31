import { useEffect, useMemo, useState } from 'react';
import { Building2, Save } from 'lucide-react';
import { Card, CardHeader, Button, Input, Badge, Skeleton } from '@/components/ui';
import { CurrencySelect } from '@/components/CurrencySelect';
import {
  useCurrentOrg,
  useIsAdmin,
  useCurrencies,
  useUpdateOrganization,
} from '@/hooks';
import { SUPPORTED_CURRENCIES } from '@/lib/currency';

const roleLabels: Record<string, string> = {
  owner: 'Proprietário',
  admin: 'Administrador',
  member: 'Membro',
  viewer: 'Visualizador',
};

function ReadOnlyField({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-gray-100 px-3.5 py-2.5 dark:border-ink-border">
      <p className="text-[11px] font-medium uppercase tracking-wide text-gray-400">{label}</p>
      <p className="mt-0.5 text-sm font-medium text-gray-900 dark:text-gray-100">{value}</p>
    </div>
  );
}

export function OrganizationCard() {
  const org = useCurrentOrg();
  const isAdmin = useIsAdmin();
  const update = useUpdateOrganization();
  // Currencies require auth; the static list is a graceful fallback.
  const { data: currencies, isLoading: currenciesLoading } = useCurrencies();

  const [name, setName] = useState(org?.name ?? '');
  const [currency, setCurrency] = useState(org?.currency ?? 'BRL');

  // Reset local form when the active org changes (e.g. via the OrgSwitcher).
  useEffect(() => {
    setName(org?.name ?? '');
    setCurrency(org?.currency ?? 'BRL');
  }, [org?.id, org?.name, org?.currency]);

  const currencyName = useMemo(() => {
    const list = currencies && currencies.length > 0 ? currencies : SUPPORTED_CURRENCIES;
    const match = list.find((c) => c.code === org?.currency);
    return match ? `${match.code} — ${match.name} (${match.symbol})` : (org?.currency ?? '—');
  }, [currencies, org?.currency]);

  const dirty = Boolean(org) && (name.trim() !== org!.name || currency !== org!.currency);

  const onSave = () => {
    const trimmed = name.trim();
    if (!trimmed) return;
    update.mutate({ name: trimmed, currency });
  };

  if (!org) {
    return (
      <Card>
        <CardHeader
          eyebrow="02 · Organização"
          title="Empresa atual"
          subtitle="Dados da organização selecionada."
        />
        <p className="text-sm text-gray-400">Nenhuma organização selecionada.</p>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader
        eyebrow="02 · Organização"
        title="Empresa atual"
        subtitle={
          isAdmin
            ? 'Edite o nome e a moeda da organização.'
            : 'Dados da organização selecionada.'
        }
      />

      <div className="space-y-4">
        <div className="flex items-center gap-3">
          <span className="flex h-11 w-11 items-center justify-center rounded-xl bg-primary/10 text-primary">
            <Building2 className="h-5 w-5" />
          </span>
          <div>
            <p className="font-heading text-base font-semibold text-gray-900 dark:text-gray-100">
              {org.name}
            </p>
            <Badge variant="neutral">{roleLabels[org.role] ?? org.role}</Badge>
          </div>
        </div>

        {isAdmin ? (
          <div className="space-y-3">
            <Input
              label="Nome da organização"
              value={name}
              onChange={(e) => setName(e.target.value)}
              maxLength={120}
            />
            {currenciesLoading ? (
              <div>
                <p className="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
                  Moeda
                </p>
                <Skeleton className="h-11 rounded-xl" />
              </div>
            ) : (
              <CurrencySelect
                label="Moeda"
                currencies={currencies}
                value={currency}
                onChange={(e) => setCurrency(e.target.value)}
              />
            )}
            <p className="text-xs text-gray-400">
              Ao alterar a moeda, todos os valores passam a ser exibidos com o novo
              símbolo e formato.
            </p>
            <div className="flex justify-end">
              <Button
                size="sm"
                onClick={onSave}
                disabled={!dirty || !name.trim()}
                loading={update.isPending}
              >
                <Save className="h-4 w-4" /> Salvar alterações
              </Button>
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-2 gap-3">
            <ReadOnlyField label="Slug" value={org.slug} />
            <ReadOnlyField label="Moeda" value={currencyName} />
          </div>
        )}
      </div>
    </Card>
  );
}
