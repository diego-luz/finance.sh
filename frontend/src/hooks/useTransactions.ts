import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query';
import { transactionService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';
import { queryKeys } from './queryKeys';
import { ApiRequestError } from '@/lib/axios';
import type {
  BulkCategorizePayload,
  BulkSettlePayload,
  DeleteScope,
  SettlePayload,
  TransactionFilters,
  TransactionPayload,
} from '@/types';

export function useTransactions(filters: TransactionFilters) {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.transactions(orgId, filters),
    queryFn: () => transactionService.list(filters),
    enabled: Boolean(orgId),
    placeholderData: keepPreviousData,
  });
}

function useInvalidateTransactional() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  return () => {
    queryClient.invalidateQueries({ queryKey: ['transactions', orgId] });
    queryClient.invalidateQueries({ queryKey: ['payables', orgId] });
    queryClient.invalidateQueries({ queryKey: ['receivables', orgId] });
    queryClient.invalidateQueries({ queryKey: ['forecast', orgId] });
    queryClient.invalidateQueries({ queryKey: queryKeys.dashboard(orgId) });
    queryClient.invalidateQueries({ queryKey: queryKeys.accounts(orgId) });
    // Card-linked transactions affect card usage + invoice views.
    queryClient.invalidateQueries({ queryKey: ['credit-cards', orgId] });
    queryClient.invalidateQueries({ queryKey: ['invoices', orgId] });
  };
}

export function useCreateTransaction() {
  const invalidate = useInvalidateTransactional();
  const toast = useToast();
  return useMutation({
    mutationFn: (payload: TransactionPayload) => transactionService.create(payload),
    onSuccess: () => {
      invalidate();
      toast.success('Transação registrada.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useUpdateTransaction() {
  const invalidate = useInvalidateTransactional();
  const toast = useToast();
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: TransactionPayload }) =>
      transactionService.update(id, payload),
    onSuccess: () => {
      invalidate();
      toast.success('Transação atualizada.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useDeleteTransaction() {
  const invalidate = useInvalidateTransactional();
  const toast = useToast();
  return useMutation({
    mutationFn: ({ id, scope = 'one' }: { id: string; scope?: DeleteScope }) =>
      transactionService.remove(id, scope),
    onSuccess: (_data, { scope }) => {
      invalidate();
      toast.success(scope === 'all' ? 'Parcelas removidas.' : 'Transação removida.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useSettleTransaction() {
  const invalidate = useInvalidateTransactional();
  const toast = useToast();
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: SettlePayload }) =>
      transactionService.settle(id, payload),
    onSuccess: () => {
      invalidate();
      toast.success('Baixa registrada.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useUnsettleTransaction() {
  const invalidate = useInvalidateTransactional();
  const toast = useToast();
  return useMutation({
    mutationFn: (id: string) => transactionService.unsettle(id),
    onSuccess: () => {
      invalidate();
      toast.success('Baixa desfeita.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useBulkSettle() {
  const invalidate = useInvalidateTransactional();
  const toast = useToast();
  return useMutation({
    mutationFn: (payload: BulkSettlePayload) => transactionService.bulkSettle(payload),
    onSuccess: ({ updated }) => {
      invalidate();
      toast.success(`${updated} transaç${updated === 1 ? 'ão marcada' : 'ões marcadas'} como paga(s).`);
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useBulkCategorize() {
  const invalidate = useInvalidateTransactional();
  const toast = useToast();
  return useMutation({
    mutationFn: (payload: BulkCategorizePayload) => transactionService.bulkCategorize(payload),
    onSuccess: ({ updated }) => {
      invalidate();
      toast.success(`${updated} transaç${updated === 1 ? 'ão categorizada' : 'ões categorizadas'}.`);
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useBulkDelete() {
  const invalidate = useInvalidateTransactional();
  const toast = useToast();
  return useMutation({
    mutationFn: (ids: string[]) => transactionService.bulkDelete(ids),
    onSuccess: ({ updated }) => {
      invalidate();
      toast.success(`${updated} transaç${updated === 1 ? 'ão removida' : 'ões removidas'}.`);
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}
