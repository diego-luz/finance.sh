import { useEffect, useMemo } from 'react';
import { useAuthStore } from '@/stores/authStore';
import { getActiveCurrency, setActiveCurrency } from '@/lib/currency';
import type { Organization, OrgRole } from '@/types';

/** Returns the currently selected organization (or undefined). */
export function useCurrentOrg(): Organization | undefined {
  const organizations = useAuthStore((s) => s.organizations);
  const currentOrgId = useAuthStore((s) => s.currentOrgId);
  return useMemo(
    () => organizations.find((o) => o.id === currentOrgId) ?? organizations[0],
    [organizations, currentOrgId],
  );
}

/**
 * Returns the active organization's currency code and keeps the module-level
 * active currency in {@link import('@/lib/currency')} in sync. Mount this once
 * high in the tree (it is wired into {@link AuthProvider}) so every
 * `formatCurrency()` call site reflects the org currency on load and switch.
 */
export function useActiveCurrency(): string {
  const org = useCurrentOrg();
  const currency = org?.currency ?? getActiveCurrency();

  useEffect(() => {
    setActiveCurrency(currency);
  }, [currency]);

  return currency;
}

/** The role of the current user in the active organization. */
export function useCurrentRole(): OrgRole | undefined {
  return useCurrentOrg()?.role;
}

/** True when the active member is a read-only viewer. */
export function useIsViewer(): boolean {
  return useCurrentRole() === 'viewer';
}

/** True when the active member can manage the team / billing (owner or admin). */
export function useIsAdmin(): boolean {
  const role = useCurrentRole();
  return role === 'owner' || role === 'admin';
}

/** True only for the organization owner. */
export function useIsOwner(): boolean {
  return useCurrentRole() === 'owner';
}
