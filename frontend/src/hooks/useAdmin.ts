import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query';
import { adminService, authService } from '@/services';
import { useToast } from '@/contexts/ToastContext';
import { queryKeys } from './queryKeys';
import { ApiRequestError } from '@/lib/axios';
import type { AdminListFilters } from '@/types';

/** Invalidates every `['admin', ...]` query (lists + stats). */
function invalidateAdmin(queryClient: ReturnType<typeof useQueryClient>) {
  queryClient.invalidateQueries({ queryKey: ['admin'] });
}

// ---------------------------------------------------------------------------
// Stats
// ---------------------------------------------------------------------------
export function useAdminStats() {
  return useQuery({
    queryKey: queryKeys.adminStats,
    queryFn: () => adminService.stats(),
  });
}

// ---------------------------------------------------------------------------
// Organizations (read-only)
// ---------------------------------------------------------------------------
export function useAdminOrganizations(filters: AdminListFilters) {
  return useQuery({
    queryKey: queryKeys.adminOrganizations(filters),
    queryFn: () => adminService.listOrganizations(filters),
    placeholderData: keepPreviousData,
  });
}

// ---------------------------------------------------------------------------
// Users
// ---------------------------------------------------------------------------
export function useAdminUsers(filters: AdminListFilters) {
  return useQuery({
    queryKey: queryKeys.adminUsers(filters),
    queryFn: () => adminService.listUsers(filters),
    placeholderData: keepPreviousData,
  });
}

export function useDisableUser() {
  const queryClient = useQueryClient();
  const toast = useToast();
  return useMutation({
    mutationFn: (id: string) => adminService.disableUser(id),
    onSuccess: () => {
      invalidateAdmin(queryClient);
      toast.success('Usuário desativado.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useEnableUser() {
  const queryClient = useQueryClient();
  const toast = useToast();
  return useMutation({
    mutationFn: (id: string) => adminService.enableUser(id),
    onSuccess: () => {
      invalidateAdmin(queryClient);
      toast.success('Usuário reativado.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

/**
 * Super-admin: reset a user's password. The user is forced to change it on
 * their next login and existing sessions are revoked server-side.
 */
export function useResetUserPassword() {
  const toast = useToast();
  return useMutation({
    mutationFn: ({ userId, new_password }: { userId: string; new_password: string }) =>
      adminService.resetUserPassword(userId, new_password),
    onSuccess: () => {
      toast.success('Senha redefinida. Informe-a ao usuário — ele será obrigado a trocá-la no próximo acesso.');
    },
  });
}

// ---------------------------------------------------------------------------
// Public: is self-service registration open?
// ---------------------------------------------------------------------------
/**
 * Whether self-service registration is open. Fails gracefully: on error it
 * falls back to `{ open: true }` so the UI never blocks signup spuriously.
 */
export function useRegistrationOpen() {
  return useQuery({
    queryKey: queryKeys.registrationOpen,
    queryFn: () => authService.registrationOpen(),
    retry: false,
    staleTime: 5 * 60_000,
  });
}
