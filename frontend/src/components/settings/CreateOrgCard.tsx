import { useState } from 'react';
import { Plus, Building2 } from 'lucide-react';
import { Card, CardHeader, Button, Input } from '@/components/ui';
import { CurrencySelect } from '@/components/CurrencySelect';
import { useCreateOrganization } from '@/hooks';

/**
 * Self-service: create an ADDITIONAL organization owned by the user — e.g. one
 * org for personal finances ("Casa") and another for the microempresa. On
 * success the app switches to the new org automatically.
 */
export function CreateOrgCard() {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState('');
  const [currency, setCurrency] = useState('BRL');
  const create = useCreateOrganization();

  function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim()) return;
    create.mutate(
      { name: name.trim(), currency },
      {
        onSuccess: () => {
          setName('');
          setCurrency('BRL');
          setOpen(false);
        },
      },
    );
  }

  return (
    <Card className="mt-4">
      <CardHeader
        eyebrow="Multi-organização"
        title="Nova organização"
        subtitle="Crie outra organização na mesma conta (ex.: Casa e Microempresa). Troque entre elas no seletor do topo."
        action={<Building2 className="h-5 w-5 text-gray-400" />}
      />
      {open ? (
        <form onSubmit={submit} className="space-y-3">
          <Input
            autoFocus
            label="Nome"
            placeholder="Ex.: Microempresa"
            value={name}
            onChange={(e) => setName(e.target.value)}
            maxLength={120}
          />
          <CurrencySelect value={currency} onChange={(e) => setCurrency(e.target.value)} />
          <div className="flex gap-2">
            <Button type="submit" loading={create.isPending} disabled={!name.trim()}>
              <Plus className="h-4 w-4" /> Criar e entrar
            </Button>
            <Button type="button" variant="ghost" onClick={() => setOpen(false)}>
              Cancelar
            </Button>
          </div>
        </form>
      ) : (
        <Button variant="outline" onClick={() => setOpen(true)}>
          <Plus className="h-4 w-4" /> Nova organização
        </Button>
      )}
    </Card>
  );
}
