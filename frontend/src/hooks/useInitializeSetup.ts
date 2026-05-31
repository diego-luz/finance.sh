import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  initializeSetup,
  type SetupInitializePayload,
} from '@/lib/api/setup';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';
import { ApiRequestError } from '@/lib/axios';
import { setActiveCurrency } from '@/lib/currency';
import { setupStatusQueryKey } from './useSetupStatus';
import type { AuthResponse } from '@/types';

/**
 * Mutation that posts the first-run bootstrap payload. On success persists
 * the returned tokens via the auth store (same path useLogin takes), flips
 * the cached setup status to `{ needs_setup: false }`, clears any stale
 * tenant data and routes the brand-new owner to the dashboard.
 *
 * Error handling is delegated to the caller because the wizard wants to
 * distinguish 409 (already-initialized → redirect to /login) from 400
 * (validation → inline messages).
 */
export function useInitializeSetup() {
  const setAuth = useAuthStore((s) => s.setAuth);
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const toast = useToast();
  const { t } = useTranslation();

  return useMutation<AuthResponse, ApiRequestError, SetupInitializePayload>({
    mutationFn: (payload) => initializeSetup(payload),
    onSuccess: (res) => {
      setAuth(res);
      // Reflect the new org's currency immediately so the first render of
      // the dashboard formats values correctly.
      setActiveCurrency(res.organization?.currency);
      // Drop any tenant cache that may have leaked in from a previous boot,
      // then bake the new probe result so any other gate observer skips the
      // wizard. Order matters — removeQueries must run BEFORE setQueryData,
      // otherwise the seeded value gets wiped.
      queryClient.removeQueries({ predicate: () => true });
      queryClient.setQueryData(setupStatusQueryKey, { needs_setup: false });
      toast.success(t('setup.success'));
      navigate('/', { replace: true });
    },
  });
}
