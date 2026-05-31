import { useMutation, useQueryClient } from '@tanstack/react-query';
import { importService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { queryKeys } from './queryKeys';
import { ApiRequestError } from '@/lib/axios';
import type {
  CsvOptions,
  ImportCommitPayload,
  ImportCommitResult,
  ImportFormat,
  ImportPreview,
} from '@/types';

interface PreviewVars {
  file: File;
  accountId: string;
  format?: ImportFormat;
  csv?: CsvOptions;
}

/**
 * Parse a statement file into a deduplicated preview. Errors are surfaced via
 * `mutation.error` so the caller can toast them; we don't toast here because
 * the preview is just a step transition (no implicit side effect).
 */
export function useImportPreview() {
  return useMutation<ImportPreview, ApiRequestError, PreviewVars>({
    mutationFn: ({ file, accountId, format, csv }) =>
      importService.preview(file, accountId, format, csv),
  });
}

/**
 * Commit the selected rows. On success invalidates the same transactional
 * caches as a transaction create (transactions, payables, receivables,
 * forecast, dashboard, accounts) so the UI reflects the imported rows.
 */
export function useImportCommit() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  return useMutation<ImportCommitResult, ApiRequestError, ImportCommitPayload>({
    mutationFn: (payload) => importService.commit(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['transactions', orgId] });
      queryClient.invalidateQueries({ queryKey: ['payables', orgId] });
      queryClient.invalidateQueries({ queryKey: ['receivables', orgId] });
      queryClient.invalidateQueries({ queryKey: ['forecast', orgId] });
      queryClient.invalidateQueries({ queryKey: queryKeys.dashboard(orgId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.accounts(orgId) });
      queryClient.invalidateQueries({ queryKey: ['credit-cards', orgId] });
      queryClient.invalidateQueries({ queryKey: ['invoices', orgId] });
    },
  });
}
