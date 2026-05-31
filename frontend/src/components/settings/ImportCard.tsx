import { useRef, useState } from 'react';
import { Upload } from 'lucide-react';
import { Card, CardHeader, Button } from '@/components/ui';
import { useToast } from '@/contexts/ToastContext';
import { accountPrivacyService, type ImportSummary } from '@/services/accountPrivacyService';

/**
 * Imports a previously-exported finance.sh JSON into a BRAND-NEW organization
 * (data portability between instances). The user picks the file; on success a
 * summary of restored rows is shown. Switch to the new org from the org switcher.
 */
export function ImportCard() {
  const toast = useToast();
  const inputRef = useRef<HTMLInputElement>(null);
  const [busy, setBusy] = useState(false);
  const [summary, setSummary] = useState<ImportSummary | null>(null);

  async function onFile(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    e.target.value = ''; // allow re-selecting the same file
    if (!file) return;
    setBusy(true);
    setSummary(null);
    try {
      const sum = await accountPrivacyService.importData(file);
      setSummary(sum);
      toast.success(`Importado em "${sum.organization_name}".`);
    } catch (err: any) {
      toast.error(err?.response?.data?.error?.message || 'Falha ao importar o arquivo.');
    } finally {
      setBusy(false);
    }
  }

  return (
    <Card className="mt-4">
      <CardHeader
        eyebrow="Portabilidade"
        title="Importar dados"
        subtitle="Restaure um arquivo de export (.json) em uma organização nova."
        action={<Upload className="h-5 w-5 text-gray-400" />}
      />
      <input
        ref={inputRef}
        type="file"
        accept="application/json,.json"
        className="hidden"
        onChange={onFile}
      />
      <Button variant="outline" onClick={() => inputRef.current?.click()} loading={busy}>
        <Upload className="h-4 w-4" /> Escolher arquivo .json
      </Button>
      {summary && (
        <p className="mt-3 text-sm text-gray-500 dark:text-gray-400">
          Importado em <strong>{summary.organization_name}</strong>: {summary.accounts} contas,{' '}
          {summary.categories} categorias, {summary.contacts} contatos, {summary.credit_cards} cartões,{' '}
          {summary.transactions} lançamentos
          {summary.skipped > 0 ? ` (${summary.skipped} ignorados)` : ''}. Troque para ela no seletor de organização.
        </p>
      )}
    </Card>
  );
}
