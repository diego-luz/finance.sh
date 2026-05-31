import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { attachmentService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';
import { queryKeys } from './queryKeys';
import { ApiRequestError } from '@/lib/axios';
import type { Attachment } from '@/types';

/** List the receipt attachments of a transaction (skipped while id is empty). */
export function useAttachments(transactionId: string | null | undefined) {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.attachments(orgId, transactionId ?? ''),
    queryFn: () => attachmentService.list(transactionId as string),
    enabled: Boolean(orgId) && Boolean(transactionId),
  });
}

/** Invalidate the attachments list of a transaction plus the transactions list
 *  (so the row's `attachment_count` indicator refreshes). */
function useInvalidateAttachments(transactionId: string) {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  return () => {
    queryClient.invalidateQueries({
      queryKey: queryKeys.attachments(orgId, transactionId),
    });
    queryClient.invalidateQueries({ queryKey: ['transactions', orgId] });
  };
}

/** Upload a file as an attachment of the given transaction. */
export function useUploadAttachment(transactionId: string) {
  const invalidate = useInvalidateAttachments(transactionId);
  const toast = useToast();
  return useMutation<Attachment, ApiRequestError, File>({
    mutationFn: (file: File) => attachmentService.upload(transactionId, file),
    onSuccess: () => {
      invalidate();
      toast.success('Comprovante anexado.');
    },
    onError: (err: ApiRequestError) =>
      toast.error(err.message || 'Não foi possível anexar o comprovante.'),
  });
}

/** Delete an attachment of the given transaction. */
export function useDeleteAttachment(transactionId: string) {
  const invalidate = useInvalidateAttachments(transactionId);
  const toast = useToast();
  return useMutation<unknown, ApiRequestError, string>({
    mutationFn: (id: string) => attachmentService.remove(id),
    onSuccess: () => {
      invalidate();
      toast.success('Comprovante removido.');
    },
    onError: (err: ApiRequestError) =>
      toast.error(err.message || 'Não foi possível remover o comprovante.'),
  });
}
