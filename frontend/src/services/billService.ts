import { api, ApiRequestError } from '@/lib/axios';
import type {
  ApiEnvelope,
  BillFilters,
  BillListResponse,
  BillSummary,
  Transaction,
} from '@/types';

/** Envelope shape for bill lists: summary may ride alongside data/meta. */
interface BillEnvelope extends ApiEnvelope<Transaction[]> {
  summary?: BillSummary;
}

function buildParams(filters: BillFilters): Record<string, string | number> {
  const params: Record<string, string | number> = {};
  if (filters.status) params.status = filters.status;
  if (filters.from) params.from = filters.from;
  if (filters.to) params.to = filters.to;
  if (filters.contact_id) params.contact_id = filters.contact_id;
  if (filters.page) params.page = filters.page;
  if (filters.per_page) params.per_page = filters.per_page;
  return params;
}

const EMPTY_SUMMARY: BillSummary = {
  total_open: 0,
  total_overdue: 0,
  due_next_7d: 0,
  count_open: 0,
};

async function fetchBills(
  path: '/payables' | '/receivables',
  filters: BillFilters,
): Promise<BillListResponse> {
  const res = await api.get<BillEnvelope>(path, { params: buildParams(filters) });
  const envelope = res.data;
  if (!envelope || envelope.success === false) {
    throw new ApiRequestError(
      envelope?.error ?? { code: 'unknown', message: 'Erro inesperado' },
      res.status,
    );
  }
  const meta = envelope.meta ?? {};
  // The summary may be returned at the top level or nested inside meta.
  const summary =
    envelope.summary ?? ((meta.summary as BillSummary | undefined) ?? EMPTY_SUMMARY);
  const data = envelope.data ?? [];
  return {
    data,
    meta: {
      page: Number(meta.page ?? 1),
      per_page: Number(meta.per_page ?? data.length),
      total: Number(meta.total ?? data.length),
      pages: Number(meta.pages ?? 1),
    },
    summary: {
      total_open: Number(summary.total_open ?? 0),
      total_overdue: Number(summary.total_overdue ?? 0),
      due_next_7d: Number(summary.due_next_7d ?? 0),
      count_open: Number(summary.count_open ?? 0),
    },
  };
}

export const billService = {
  payables: (filters: BillFilters) => fetchBills('/payables', filters),
  receivables: (filters: BillFilters) => fetchBills('/receivables', filters),
};
