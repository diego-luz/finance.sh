import { useState, type MouseEvent } from 'react';
import {
  Download,
  ChevronDown,
  FileSpreadsheet,
  FileText,
  FileType,
  FileJson,
  Loader2,
} from 'lucide-react';
import { Button } from '@/components/ui';
import { reportService, downloadBlob } from '@/services';
import { useExportMyData } from '@/hooks';
import { useToast } from '@/contexts/ToastContext';
import { cn } from '@/lib/cn';

interface ExportMenuProps {
  /** Optional date range applied to the CSV + statement PDF exports. */
  range?: { from?: string; to?: string };
  /** Compact trigger (icon + short label) for dense toolbars. */
  compact?: boolean;
  className?: string;
}

type Job = 'xlsx' | 'statement' | 'csv' | null;

function stamp(): string {
  return new Date().toISOString().slice(0, 10);
}

/**
 * "Exportar" dropdown. Groups the file exports (Excel / PDF extrato / CSV) and,
 * under a separate labeled group, the LGPD data-portability JSON export.
 */
export function ExportMenu({ range, compact, className }: ExportMenuProps) {
  const toast = useToast();
  const exportMyData = useExportMyData();
  const [open, setOpen] = useState(false);
  const [job, setJob] = useState<Job>(null);

  const busy = job !== null || exportMyData.isPending;

  const run = async (kind: Exclude<Job, null>) => {
    setJob(kind);
    try {
      if (kind === 'xlsx') {
        const blob = await reportService.dataXlsx();
        downloadBlob(blob, `finance-sh-dados-${stamp()}.xlsx`);
      } else if (kind === 'statement') {
        const blob = await reportService.statementPdf(range ?? {});
        downloadBlob(blob, `extrato-${stamp()}.pdf`);
      } else {
        const blob = await reportService.transactionsCsv(range ?? {});
        downloadBlob(blob, `transacoes-${stamp()}.csv`);
      }
      toast.success('Exportação concluída. O download foi iniciado.');
      setOpen(false);
    } catch {
      toast.error('Não foi possível concluir a exportação.');
    } finally {
      setJob(null);
    }
  };

  const onLgpd = () => {
    exportMyData.mutate();
    setOpen(false);
  };

  return (
    <div className={cn('relative', className)}>
      <Button
        variant="outline"
        size={compact ? 'sm' : 'md'}
        onClick={() => setOpen((v) => !v)}
        onBlur={() => setTimeout(() => setOpen(false), 150)}
        disabled={busy}
        aria-haspopup="menu"
        aria-expanded={open}
      >
        {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : <Download className="h-4 w-4" />}
        Exportar
        <ChevronDown className="h-3.5 w-3.5 opacity-70" />
      </Button>

      {open && (
        <div
          role="menu"
          className="absolute right-0 z-20 mt-1.5 w-64 animate-fade-in overflow-hidden rounded-xl border border-gray-200 bg-white p-1.5 shadow-card dark:border-ink-border dark:bg-ink-surface"
        >
          <MenuItem
            icon={FileSpreadsheet}
            label="Excel (.xlsx)"
            description="Planilha completa dos seus dados"
            onMouseDown={(e) => e.preventDefault()}
            onClick={() => run('xlsx')}
            loading={job === 'xlsx'}
          />
          <MenuItem
            icon={FileType}
            label="PDF — Extrato"
            description="Relatório de movimentações"
            onMouseDown={(e) => e.preventDefault()}
            onClick={() => run('statement')}
            loading={job === 'statement'}
          />
          <MenuItem
            icon={FileText}
            label="CSV"
            description="Transações em planilha simples"
            onMouseDown={(e) => e.preventDefault()}
            onClick={() => run('csv')}
            loading={job === 'csv'}
          />

          <div className="my-1.5 border-t border-gray-100 dark:border-ink-border" />
          <p className="px-2.5 pb-1 pt-1 text-[11px] font-semibold uppercase tracking-wide text-gray-400">
            Portabilidade LGPD (JSON)
          </p>
          <MenuItem
            icon={FileJson}
            label="Baixar meus dados (JSON)"
            description="Cópia portátil dos seus dados pessoais"
            onMouseDown={(e) => e.preventDefault()}
            onClick={onLgpd}
            loading={exportMyData.isPending}
          />
        </div>
      )}
    </div>
  );
}

function MenuItem({
  icon: Icon,
  label,
  description,
  loading,
  onClick,
  onMouseDown,
}: {
  icon: typeof FileText;
  label: string;
  description: string;
  loading?: boolean;
  onClick: () => void;
  onMouseDown: (e: MouseEvent) => void;
}) {
  return (
    <button
      type="button"
      role="menuitem"
      onMouseDown={onMouseDown}
      onClick={onClick}
      disabled={loading}
      className="flex w-full items-center gap-3 rounded-lg px-2.5 py-2 text-left transition hover:bg-gray-100 disabled:cursor-not-allowed disabled:opacity-60 dark:hover:bg-ink-elevated"
    >
      <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
        {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Icon className="h-4 w-4" />}
      </span>
      <span className="min-w-0">
        <span className="block text-sm font-medium text-gray-900 dark:text-gray-100">{label}</span>
        <span className="block truncate text-xs text-gray-400">{description}</span>
      </span>
    </button>
  );
}
