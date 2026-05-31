import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { organizationService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';
import { setActiveCurrency } from '@/lib/currency';
import { ApiRequestError } from '@/lib/axios';
import type { Organization, UpdateOrganizationPayload } from '@/types';
import { queryKeys } from './queryKeys';

/** Query keys (prefixes) for every cache that renders monetary values. */
const MONEY_QUERY_PREFIXES = [
  'transactions',
  'dashboard',
  'accounts',
  'budgets',
  'goals',
  'payables',
  'receivables',
  'forecast',
  'report-summary',
  'credit-cards',
  'invoices',
  'invoice',
] as const;

/** Supported currencies for the current session (requires auth). */
export function useCurrencies() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.currencies,
    queryFn: () => organizationService.getCurrencies(),
    enabled: Boolean(orgId),
    staleTime: 1000 * 60 * 60, // currencies rarely change
  });
}

/** The current organization (server-authoritative). */
export function useOrganization() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.organization(orgId),
    queryFn: () => organizationService.getOrganization(),
    enabled: Boolean(orgId),
  });
}

/**
 * Update the current organization (name/currency). On success: persist the new
 * details into the auth store, sync the active currency, invalidate every
 * money-bearing query so amounts reformat, and toast.
 */
export function useUpdateOrganization() {
  const queryClient = useQueryClient();
  const toast = useToast();
  const setOrgDetails = useAuthStore((s) => s.setOrgDetails);

  return useMutation<Organization, ApiRequestError, UpdateOrganizationPayload>({
    mutationFn: (payload: UpdateOrganizationPayload) =>
      organizationService.updateOrganization(payload),
    onSuccess: (org) => {
      setOrgDetails(org);
      setActiveCurrency(org.currency);
      queryClient.invalidateQueries({ queryKey: queryKeys.organization(org.id) });
      queryClient.invalidateQueries({ queryKey: queryKeys.me });
      for (const prefix of MONEY_QUERY_PREFIXES) {
        queryClient.invalidateQueries({ queryKey: [prefix] });
      }
      toast.success('Organização atualizada com sucesso.');
    },
    onError: (err: ApiRequestError) =>
      toast.error(err.message || 'Não foi possível atualizar a organização.'),
  });
}

/**
 * Create an ADDITIONAL organization owned by the user (e.g. a personal "Casa"
 * org plus a "Microempresa" org). On success: add it to the store, SWITCH to it
 * (so the X-Organization-ID header points at the new org), sync currency and
 * invalidate every money-bearing query so the app reloads for the new tenant.
 */
export function useCreateOrganization() {
  const queryClient = useQueryClient();
  const toast = useToast();
  const setOrganizations = useAuthStore((s) => s.setOrganizations);
  const setOrg = useAuthStore((s) => s.setOrg);

  return useMutation<Organization, ApiRequestError, { name: string; currency?: string }>({
    mutationFn: (payload) => organizationService.createOrganization(payload),
    onSuccess: (org) => {
      const current = useAuthStore.getState().organizations;
      setOrganizations([...current, org]);
      setOrg(org.id);
      setActiveCurrency(org.currency);
      queryClient.invalidateQueries({ queryKey: queryKeys.me });
      for (const prefix of MONEY_QUERY_PREFIXES) {
        queryClient.invalidateQueries({ queryKey: [prefix] });
      }
      toast.success(`Organização "${org.name}" criada. Você já está nela.`);
    },
    onError: (err: ApiRequestError) =>
      toast.error(err.message || 'Não foi possível criar a organização.'),
  });
}
