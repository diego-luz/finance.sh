import { api, unwrap } from '@/lib/axios';
import type { ApiEnvelope, ReportSummary, TransactionType } from '@/types';

export interface TransactionReportFilters {
  from?: string;
  to?: string;
  type?: TransactionType | '';
}

/** Period bounds for the summary / statement endpoints. */
export interface ReportPeriod {
  from?: string;
  to?: string;
}

function periodParams(filters: ReportPeriod): Record<string, string> {
  const params: Record<string, string> = {};
  if (filters.from) params.from = filters.from;
  if (filters.to) params.to = filters.to;
  return params;
}

/**
 * Report endpoints. The file endpoints return a raw Blob (not the JSON
 * envelope), so they request `responseType: 'blob'`. The summary endpoint
 * returns the usual envelope.
 */
export const reportService = {
  /** Aggregated report data for the Relatórios page. */
  summary: (filters: ReportPeriod = {}): Promise<ReportSummary> =>
    unwrap<ReportSummary>(
      api.get<ApiEnvelope<ReportSummary>>('/reports/summary', {
        params: periodParams(filters),
      }),
    ),

  /** Transactions CSV (existing endpoint). */
  transactionsCsv: async (filters: TransactionReportFilters = {}): Promise<Blob> => {
    const params: Record<string, string> = {};
    if (filters.from) params.from = filters.from;
    if (filters.to) params.to = filters.to;
    if (filters.type) params.type = filters.type;
    const res = await api.get('/reports/transactions.csv', {
      params,
      responseType: 'blob',
    });
    return res.data as Blob;
  },

  /** Full data export as an Excel workbook. */
  dataXlsx: async (): Promise<Blob> => {
    const res = await api.get('/export/data.xlsx', { responseType: 'blob' });
    return res.data as Blob;
  },

  /** Statement PDF for an optional date range. */
  statementPdf: async (filters: ReportPeriod = {}): Promise<Blob> => {
    const res = await api.get('/reports/statement.pdf', {
      params: periodParams(filters),
      responseType: 'blob',
    });
    return res.data as Blob;
  },

  /** Monthly report PDF for a given month/year. */
  monthlyPdf: async (month: number, year: number): Promise<Blob> => {
    const res = await api.get('/reports/monthly.pdf', {
      params: { month: String(month), year: String(year) },
      responseType: 'blob',
    });
    return res.data as Blob;
  },
};

/** Trigger a browser download for a Blob with the given filename. */
export function downloadBlob(blob: Blob, filename: string): void {
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  window.URL.revokeObjectURL(url);
}
