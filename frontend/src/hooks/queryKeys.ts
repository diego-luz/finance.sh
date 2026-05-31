import type {
  AdminListFilters,
  AuditFilters,
  BillFilters,
  BudgetFilters,
  TransactionFilters,
} from '@/types';

/** Centralized React Query keys, scoped by organization where relevant. */
export const queryKeys = {
  me: ['me'] as const,
  currencies: ['currencies'] as const,
  organization: (orgId: string | null) => ['organization', orgId] as const,
  accounts: (orgId: string | null) => ['accounts', orgId] as const,
  categories: (orgId: string | null) => ['categories', orgId] as const,
  tags: (orgId: string | null) => ['tags', orgId] as const,
  search: (orgId: string | null, query: string) => ['search', orgId, query] as const,
  contacts: (orgId: string | null) => ['contacts', orgId] as const,
  transactions: (orgId: string | null, filters: TransactionFilters) =>
    ['transactions', orgId, filters] as const,
  attachments: (orgId: string | null, transactionId: string) =>
    ['attachments', orgId, transactionId] as const,
  payables: (orgId: string | null, filters: BillFilters) =>
    ['payables', orgId, filters] as const,
  receivables: (orgId: string | null, filters: BillFilters) =>
    ['receivables', orgId, filters] as const,
  forecast: (orgId: string | null, months: number) =>
    ['forecast', orgId, months] as const,
  dashboard: (orgId: string | null) => ['dashboard', orgId] as const,
  creditCards: (orgId: string | null) => ['credit-cards', orgId] as const,
  invoices: (orgId: string | null, cardId: string) =>
    ['invoices', orgId, cardId] as const,
  invoice: (orgId: string | null, cardId: string, reference: string) =>
    ['invoice', orgId, cardId, reference] as const,
  goals: (orgId: string | null) => ['goals', orgId] as const,
  budgets: (orgId: string | null, filters: BudgetFilters) =>
    ['budgets', orgId, filters] as const,
  members: (orgId: string | null) => ['members', orgId] as const,
  invitations: (orgId: string | null) => ['invitations', orgId] as const,
  notifications: (orgId: string | null) => ['notifications', orgId] as const,
  reportSummary: (orgId: string | null, range: { from?: string; to?: string }) =>
    ['report-summary', orgId, range] as const,
  sessions: ['sessions'] as const,
  categorizationRules: (orgId: string | null) =>
    ['categorization-rules', orgId] as const,
  recurrences: (orgId: string | null) => ['recurrences', orgId] as const,
  auditLogs: (orgId: string | null, filters: AuditFilters) =>
    ['audit-logs', orgId, filters] as const,
  // Platform super-admin (not org-scoped).
  adminStats: ['admin', 'stats'] as const,
  adminOrganizations: (filters: AdminListFilters) =>
    ['admin', 'organizations', filters] as const,
  adminUsers: (filters: AdminListFilters) => ['admin', 'users', filters] as const,
  registrationOpen: ['registration-open'] as const,
};
