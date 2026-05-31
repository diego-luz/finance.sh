import { useEffect, useState, type ReactNode } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { authService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { useActiveCurrency } from '@/hooks/useCurrentOrg';
import { setActiveCurrency } from '@/lib/currency';
import { queryKeys } from '@/hooks/queryKeys';
import { Spinner } from '@/components/ui';
import { Logo } from '@/components/Logo';

/**
 * Hydrates the authenticated session: when an access token exists in the
 * persisted store, fetches /me to refresh the user + organization list.
 * Shows a brief splash while hydrating so we don't flash protected UI.
 */
export function AuthProvider({ children }: { children: ReactNode }) {
  const accessToken = useAuthStore((s) => s.accessToken);
  const setUser = useAuthStore((s) => s.setUser);
  const setOrganizations = useAuthStore((s) => s.setOrganizations);
  const queryClient = useQueryClient();
  const [hydrating, setHydrating] = useState(Boolean(accessToken));

  // Keep the module-level active currency in sync with the current org so every
  // formatCurrency() call site reflects the org currency on load and on switch.
  useActiveCurrency();

  useEffect(() => {
    let active = true;
    if (!accessToken) {
      setHydrating(false);
      return;
    }
    setHydrating(true);
    authService
      .me()
      .then((me) => {
        if (!active) return;
        setUser(me.user);
        setOrganizations(me.organizations);
        queryClient.setQueryData(queryKeys.me, me);
        // Apply immediately (synchronously) once orgs resolve, so the first
        // post-hydration render already formats in the org currency.
        const orgId = useAuthStore.getState().currentOrgId;
        const org = me.organizations.find((o) => o.id === orgId) ?? me.organizations[0];
        setActiveCurrency(org?.currency);
      })
      .catch(() => {
        // On failure the axios interceptor handles refresh/logout.
      })
      .finally(() => {
        if (active) setHydrating(false);
      });
    return () => {
      active = false;
    };
    // Only run on token changes (login/logout), not on every render.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [accessToken]);

  if (hydrating) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-gray-100 dark:bg-ink-base">
        <Logo />
        <Spinner />
      </div>
    );
  }

  return <>{children}</>;
}
