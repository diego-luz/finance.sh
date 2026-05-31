import {
  useEffect,
  useMemo,
  useRef,
  useState,
  type ChangeEvent,
  type DragEvent,
} from 'react';
import {
  ArrowLeft,
  CheckCircle2,
  ChevronDown,
  ChevronRight,
  FileSpreadsheet,
  FileUp,
  Loader2,
  Upload,
} from 'lucide-react';
import {
  Badge,
  Button,
  Input,
  Modal,
  Select,
  Table,
  TBody,
  Td,
  Th,
  THead,
  Tr,
} from '@/components/ui';
import { useToast } from '@/contexts/ToastContext';
import { useAccounts, useCategories, useImportCommit, useImportPreview } from '@/hooks';
import { formatCurrency } from '@/lib/currency';
import { formatDateShort } from '@/lib/date';
import { cn } from '@/lib/cn';
import type {
  CsvDelimiter,
  CsvOptions,
  DecimalSep,
  ImportPreview,
  ImportRow,
} from '@/types';

interface Props {
  open: boolean;
  onClose: () => void;
  /** Optional account pre-selected when launching from a filtered view. */
  defaultAccountId?: string;
}

type Step = 'upload' | 'preview' | 'done';

const ACCEPT = '.ofx,.csv,text/csv,application/x-ofx';
/** Defensive cap so a huge file never freezes the modal table. */
const MAX_RENDERED_ROWS = 1000;

const delimiterOptions: { value: CsvDelimiter; label: string }[] = [
  { value: '', label: 'Automático' },
  { value: ',', label: 'Vírgula ( , )' },
  { value: ';', label: 'Ponto e vírgula ( ; )' },
];

const decimalOptions: { value: DecimalSep; label: string }[] = [
  { value: ',', label: 'Vírgula ( , )' },
  { value: '.', label: 'Ponto ( . )' },
];

function hasCsvExtension(name: string): boolean {
  return name.toLowerCase().endsWith('.csv');
}

function RowTypeBadge({ type }: { type: ImportRow['type'] }) {
  return type === 'income' ? (
    <Badge variant="success" dot>
      Receita
    </Badge>
  ) : (
    <Badge variant="danger" dot>
      Despesa
    </Badge>
  );
}

export function ImportStatementModal({ open, onClose, defaultAccountId }: Props) {
  const toast = useToast();
  const { data: accounts } = useAccounts();
  const { data: categories } = useCategories();
  const preview = useImportPreview();
  const commit = useImportCommit();

  const inputRef = useRef<HTMLInputElement>(null);

  // --- Step state ---------------------------------------------------------
  const [step, setStep] = useState<Step>('upload');
  const [accountId, setAccountId] = useState('');
  const [categoryId, setCategoryId] = useState('');
  const [file, setFile] = useState<File | null>(null);
  const [dragOver, setDragOver] = useState(false);

  // CSV options (kept "auto" by default).
  const [showCsv, setShowCsv] = useState(false);
  const [delimiter, setDelimiter] = useState<CsvDelimiter>('');
  const [hasHeader, setHasHeader] = useState(true);
  const [dateFormat, setDateFormat] = useState('02/01/2006');
  const [decimalSep, setDecimalSep] = useState<DecimalSep>(',');
  const [dateCol, setDateCol] = useState('');
  const [descCol, setDescCol] = useState('');
  const [amountCol, setAmountCol] = useState('');

  // Preview/commit state.
  const [previewData, setPreviewData] = useState<ImportPreview | null>(null);
  const [selected, setSelected] = useState<Set<number>>(new Set());
  const [result, setResult] = useState<{ created: number; skipped: number } | null>(null);

  // Reset everything whenever the modal is (re)opened.
  useEffect(() => {
    if (!open) return;
    setStep('upload');
    setAccountId(defaultAccountId ?? '');
    setCategoryId('');
    setFile(null);
    setDragOver(false);
    setShowCsv(false);
    setDelimiter('');
    setHasHeader(true);
    setDateFormat('02/01/2006');
    setDecimalSep(',');
    setDateCol('');
    setDescCol('');
    setAmountCol('');
    setPreviewData(null);
    setSelected(new Set());
    setResult(null);
    preview.reset();
    commit.reset();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  const accountOptions = useMemo(
    () => (accounts ?? []).map((a) => ({ value: a.id, label: a.name })),
    [accounts],
  );
  const categoryOptions = useMemo(
    () => [
      { value: '', label: 'Sem categoria padrão' },
      ...(categories ?? []).map((c) => ({ value: c.id, label: c.name })),
    ],
    [categories],
  );

  const isCsv = file ? hasCsvExtension(file.name) : false;

  // --- File handling ------------------------------------------------------
  const acceptFile = (f: File | undefined) => {
    if (!f) return;
    const lower = f.name.toLowerCase();
    if (!lower.endsWith('.ofx') && !lower.endsWith('.csv')) {
      toast.error('Formato não suportado. Selecione um arquivo .ofx ou .csv.');
      return;
    }
    if (f.size === 0) {
      toast.error('O arquivo está vazio.');
      return;
    }
    setFile(f);
  };

  const onInputChange = (e: ChangeEvent<HTMLInputElement>) => {
    acceptFile(e.target.files?.[0]);
    e.target.value = '';
  };

  const onDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setDragOver(false);
    if (preview.isPending) return;
    acceptFile(e.dataTransfer.files?.[0]);
  };

  // --- Step 1 → 2: run preview -------------------------------------------
  const runPreview = () => {
    if (!accountId) {
      toast.error('Selecione a conta de destino.');
      return;
    }
    if (!file) {
      toast.error('Selecione um arquivo de extrato.');
      return;
    }

    const csv: CsvOptions | undefined = isCsv
      ? {
          delimiter,
          has_header: hasHeader,
          date_col: dateCol.trim() === '' ? -1 : Number(dateCol),
          desc_col: descCol.trim() === '' ? -1 : Number(descCol),
          amount_col: amountCol.trim() === '' ? -1 : Number(amountCol),
          date_format: dateFormat.trim() || '02/01/2006',
          decimal_sep: decimalSep,
        }
      : undefined;

    preview.mutate(
      { file, accountId, format: 'auto', csv },
      {
        onSuccess: (data) => {
          setPreviewData(data);
          // NEW rows checked by default; duplicates unchecked.
          setSelected(
            new Set(data.rows.filter((r) => !r.duplicate).map((r) => r.index)),
          );
          if (data.rows.length === 0) {
            toast.info('Nenhuma transação encontrada no arquivo.');
          }
          setStep('preview');
        },
        onError: (err) =>
          toast.error(err.message || 'Não foi possível ler o arquivo.'),
      },
    );
  };

  // --- Step 2 selection helpers ------------------------------------------
  const rows = previewData?.rows ?? [];
  const renderedRows = rows.slice(0, MAX_RENDERED_ROWS);
  const truncated = rows.length > MAX_RENDERED_ROWS;

  const allSelected = rows.length > 0 && rows.every((r) => selected.has(r.index));
  const toggleAll = () => {
    setSelected((prev) => {
      if (rows.every((r) => prev.has(r.index))) return new Set();
      return new Set(rows.map((r) => r.index));
    });
  };
  const toggleOne = (index: number) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(index)) next.delete(index);
      else next.add(index);
      return next;
    });
  };

  const selectedCount = selected.size;

  // --- Step 2 → 3: commit -------------------------------------------------
  const runCommit = () => {
    if (!previewData || selectedCount === 0) return;
    const chosen = previewData.rows.filter((r) => selected.has(r.index));
    commit.mutate(
      {
        account_id: accountId,
        category_id: categoryId || undefined,
        rows: chosen.map((r) => ({
          date: r.date,
          description: r.description,
          amount: r.amount,
          type: r.type,
          external_id: r.external_id,
        })),
      },
      {
        onSuccess: (res) => {
          setResult(res);
          setStep('done');
          toast.success(
            `${res.created} transaç${res.created === 1 ? 'ão importada' : 'ões importadas'}` +
              (res.skipped > 0 ? `, ${res.skipped} ignorada(s).` : '.'),
          );
        },
        onError: (err) =>
          toast.error(err.message || 'Não foi possível importar as transações.'),
      },
    );
  };

  // --- Footer per step ----------------------------------------------------
  const footer =
    step === 'upload' ? (
      <>
        <Button variant="ghost" type="button" onClick={onClose} disabled={preview.isPending}>
          Cancelar
        </Button>
        <Button
          type="button"
          onClick={runPreview}
          loading={preview.isPending}
          disabled={!accountId || !file}
        >
          Pré-visualizar <ChevronRight className="h-4 w-4" />
        </Button>
      </>
    ) : step === 'preview' ? (
      <>
        <Button
          variant="ghost"
          type="button"
          onClick={() => setStep('upload')}
          disabled={commit.isPending}
        >
          <ArrowLeft className="h-4 w-4" /> Voltar
        </Button>
        <Button
          type="button"
          onClick={runCommit}
          loading={commit.isPending}
          disabled={selectedCount === 0}
        >
          <Upload className="h-4 w-4" />
          {selectedCount > 0
            ? `Importar ${selectedCount} transaç${selectedCount === 1 ? 'ão' : 'ões'}`
            : 'Importar transações'}
        </Button>
      </>
    ) : (
      <Button type="button" onClick={onClose}>
        Concluir
      </Button>
    );

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Importar extrato"
      description="Importe movimentações de um arquivo OFX ou CSV do seu banco."
      size="lg"
      footer={footer}
    >
      {step === 'upload' && (
        <div className="space-y-4">
          <Select
            label="Conta de destino"
            placeholder="Selecione a conta"
            options={accountOptions}
            value={accountId}
            onChange={(e) => setAccountId(e.target.value)}
          />

          {/* Dropzone */}
          <div>
            <span className="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
              Arquivo do extrato
            </span>
            <input
              ref={inputRef}
              type="file"
              accept={ACCEPT}
              className="sr-only"
              onChange={onInputChange}
              disabled={preview.isPending}
              aria-label="Selecionar arquivo de extrato"
            />
            <div
              role="button"
              tabIndex={0}
              onClick={() => !preview.isPending && inputRef.current?.click()}
              onKeyDown={(e) => {
                if ((e.key === 'Enter' || e.key === ' ') && !preview.isPending) {
                  e.preventDefault();
                  inputRef.current?.click();
                }
              }}
              onDragOver={(e) => {
                e.preventDefault();
                setDragOver(true);
              }}
              onDragLeave={() => setDragOver(false)}
              onDrop={onDrop}
              aria-disabled={preview.isPending}
              className={cn(
                'flex cursor-pointer flex-col items-center justify-center gap-1.5 rounded-xl border border-dashed px-4 py-7 text-center transition',
                dragOver
                  ? 'border-primary bg-primary/5'
                  : 'border-gray-300 hover:border-primary/60 hover:bg-gray-50 dark:border-ink-border dark:hover:bg-ink-elevated',
                preview.isPending && 'pointer-events-none opacity-70',
              )}
            >
              {file ? (
                <>
                  <FileSpreadsheet className="h-6 w-6 text-primary" />
                  <span className="text-sm font-medium text-gray-900 dark:text-gray-100">
                    {file.name}
                  </span>
                  <span className="text-xs text-gray-400">
                    {isCsv ? 'CSV' : 'OFX'} · clique para trocar de arquivo
                  </span>
                </>
              ) : (
                <>
                  <FileUp className="h-6 w-6 text-gray-400" />
                  <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                    Arraste o arquivo ou clique para escolher
                  </span>
                  <span className="text-xs text-gray-400">
                    Formatos aceitos: .ofx e .csv
                  </span>
                </>
              )}
            </div>
          </div>

          <Select
            label="Categoria padrão (opcional)"
            options={categoryOptions}
            value={categoryId}
            onChange={(e) => setCategoryId(e.target.value)}
          />

          {/* CSV options — collapsible, only relevant for CSV files. */}
          {isCsv && (
            <div className="rounded-xl border border-gray-200 dark:border-ink-border">
              <button
                type="button"
                onClick={() => setShowCsv((v) => !v)}
                className="flex w-full items-center justify-between gap-2 px-4 py-3 text-left"
                aria-expanded={showCsv}
              >
                <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                  Opções de CSV
                </span>
                <ChevronDown
                  className={cn(
                    'h-4 w-4 text-gray-400 transition-transform',
                    showCsv && 'rotate-180',
                  )}
                />
              </button>
              {showCsv && (
                <div className="space-y-4 border-t border-gray-100 px-4 py-4 dark:border-ink-border">
                  <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                    <Select
                      label="Delimitador"
                      options={delimiterOptions}
                      value={delimiter}
                      onChange={(e) => setDelimiter(e.target.value as CsvDelimiter)}
                    />
                    <Select
                      label="Separador decimal"
                      options={decimalOptions}
                      value={decimalSep}
                      onChange={(e) => setDecimalSep(e.target.value as DecimalSep)}
                    />
                  </div>
                  <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                    <Input
                      label="Formato de data"
                      placeholder="dd/mm/aaaa"
                      hint="Ex.: 02/01/2006 = dd/mm/aaaa"
                      value={dateFormat}
                      onChange={(e) => setDateFormat(e.target.value)}
                    />
                    <label className="flex items-center gap-2 pt-7 text-sm text-gray-700 dark:text-gray-300">
                      <input
                        type="checkbox"
                        checked={hasHeader}
                        onChange={(e) => setHasHeader(e.target.checked)}
                        className="h-4 w-4 cursor-pointer rounded border-gray-300 text-primary focus:ring-primary/30 dark:border-ink-border dark:bg-ink-elevated"
                      />
                      Tem cabeçalho
                    </label>
                  </div>
                  <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
                    <Input
                      label="Coluna data"
                      type="number"
                      placeholder="auto"
                      hint="Índice (0-based)"
                      value={dateCol}
                      onChange={(e) => setDateCol(e.target.value)}
                    />
                    <Input
                      label="Coluna descrição"
                      type="number"
                      placeholder="auto"
                      hint="Índice (0-based)"
                      value={descCol}
                      onChange={(e) => setDescCol(e.target.value)}
                    />
                    <Input
                      label="Coluna valor"
                      type="number"
                      placeholder="auto"
                      hint="Índice (0-based)"
                      value={amountCol}
                      onChange={(e) => setAmountCol(e.target.value)}
                    />
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {step === 'preview' && previewData && (
        <div className="space-y-4">
          {/* Summary */}
          <div className="flex flex-wrap items-center gap-2 rounded-xl bg-gray-50 px-4 py-3 dark:bg-ink-elevated">
            <Badge variant="info">
              Formato: {previewData.format.toUpperCase()}
            </Badge>
            <span className="text-sm text-gray-600 dark:text-gray-300">
              <strong className="text-primary">{previewData.summary.new}</strong> nova(s),{' '}
              <strong className="text-amber-600 dark:text-amber-400">
                {previewData.summary.duplicates}
              </strong>{' '}
              duplicada(s) de{' '}
              <strong className="text-gray-900 dark:text-gray-100">
                {previewData.summary.total}
              </strong>
            </span>
          </div>

          {rows.length === 0 ? (
            <p className="rounded-xl border border-dashed border-gray-300 px-4 py-8 text-center text-sm text-gray-500 dark:border-ink-border dark:text-gray-400">
              Nenhuma transação foi encontrada no arquivo.
            </p>
          ) : (
            <>
              <div className="money-sensitive max-h-[48vh] overflow-y-auto rounded-xl border border-gray-200 dark:border-ink-border">
                <Table>
                  <THead>
                    <Th className="w-10">
                      <input
                        type="checkbox"
                        aria-label="Selecionar todas"
                        checked={allSelected}
                        onChange={toggleAll}
                        className="h-4 w-4 cursor-pointer rounded border-gray-300 text-primary focus:ring-primary/30 dark:border-ink-border dark:bg-ink-elevated"
                      />
                    </Th>
                    <Th>Data</Th>
                    <Th>Descrição</Th>
                    <Th>Tipo</Th>
                    <Th className="text-right">Valor</Th>
                  </THead>
                  <TBody>
                    {renderedRows.map((r) => {
                      const isChecked = selected.has(r.index);
                      const isIncome = r.type === 'income';
                      return (
                        <Tr
                          key={r.index}
                          className={cn(
                            isChecked && 'bg-primary/5 dark:bg-primary/10',
                            r.duplicate && !isChecked && 'opacity-70',
                          )}
                        >
                          <Td className="w-10">
                            <input
                              type="checkbox"
                              aria-label={`Selecionar ${r.description}`}
                              checked={isChecked}
                              onChange={() => toggleOne(r.index)}
                              className="h-4 w-4 cursor-pointer rounded border-gray-300 text-primary focus:ring-primary/30 dark:border-ink-border dark:bg-ink-elevated"
                            />
                          </Td>
                          <Td className="whitespace-nowrap text-gray-500 dark:text-gray-400">
                            {formatDateShort(r.date)}
                          </Td>
                          <Td>
                            <div className="min-w-0">
                              <p className="truncate font-medium text-gray-900 dark:text-gray-100">
                                {r.description || '—'}
                              </p>
                              {r.duplicate && (
                                <div className="mt-0.5 flex items-center gap-1.5">
                                  <Badge variant="warning">Duplicada</Badge>
                                  {r.reason && (
                                    <span className="truncate text-xs text-gray-400">
                                      {r.reason}
                                    </span>
                                  )}
                                </div>
                              )}
                            </div>
                          </Td>
                          <Td>
                            <RowTypeBadge type={r.type} />
                          </Td>
                          <Td className="text-right">
                            <span
                              className={cn(
                                'font-semibold tabular-nums',
                                isIncome ? 'text-primary' : 'text-red-500',
                              )}
                            >
                              {isIncome ? '+' : '−'} {formatCurrency(r.amount)}
                            </span>
                          </Td>
                        </Tr>
                      );
                    })}
                  </TBody>
                </Table>
              </div>
              {truncated && (
                <p className="text-xs text-gray-400">
                  Exibindo as primeiras {MAX_RENDERED_ROWS} de {rows.length} linhas. A
                  seleção e a importação consideram todas as linhas.
                </p>
              )}
            </>
          )}
        </div>
      )}

      {step === 'done' && result && (
        <div className="flex flex-col items-center justify-center gap-3 py-8 text-center">
          <span className="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/10 text-primary">
            <CheckCircle2 className="h-7 w-7" />
          </span>
          <h3 className="font-heading text-base font-semibold text-gray-900 dark:text-gray-100">
            Importação concluída
          </h3>
          <p className="max-w-sm text-sm text-gray-500 dark:text-gray-400">
            <strong className="text-primary">{result.created}</strong> transaç
            {result.created === 1 ? 'ão importada' : 'ões importadas'}
            {result.skipped > 0 && (
              <>
                {' '}e{' '}
                <strong className="text-gray-700 dark:text-gray-300">
                  {result.skipped}
                </strong>{' '}
                ignorada(s)
              </>
            )}
            .
          </p>
        </div>
      )}

      {/* Inline busy hint for the preview parse (in addition to footer spinner). */}
      {step === 'upload' && preview.isPending && (
        <div className="mt-4 flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
          <Loader2 className="h-4 w-4 animate-spin text-primary" /> Lendo o arquivo…
        </div>
      )}
    </Modal>
  );
}
