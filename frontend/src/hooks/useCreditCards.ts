import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { creditCardService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';
import { queryKeys } from './queryKeys';
import { ApiRequestError } from '@/lib/axios';
import type { CreditCardPayload, InvoicePayPayload } from '@/types';

export function useCreditCards() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.creditCards(orgId),
    queryFn: () => creditCardService.list(),
    enabled: Boolean(orgId),
  });
}

export function useCreateCreditCard() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: (payload: CreditCardPayload) => creditCardService.create(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.creditCards(orgId) });
      toast.success('Cartão criado com sucesso.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useUpdateCreditCard() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: CreditCardPayload }) =>
      creditCardService.update(id, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.creditCards(orgId) });
      toast.success('Cartão atualizado.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useDeleteCreditCard() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: (id: string) => creditCardService.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.creditCards(orgId) });
      toast.success('Cartão removido.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useInvoices(cardId: string, enabled = true) {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.invoices(orgId, cardId),
    queryFn: () => creditCardService.listInvoices(cardId),
    enabled: Boolean(orgId) && Boolean(cardId) && enabled,
  });
}

export function useInvoice(cardId: string, reference: string, enabled = true) {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.invoice(orgId, cardId, reference),
    queryFn: () => creditCardService.getInvoice(cardId, reference),
    enabled: Boolean(orgId) && Boolean(cardId) && Boolean(reference) && enabled,
  });
}

export function usePayInvoice() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: ({
      cardId,
      reference,
      payload,
    }: {
      cardId: string;
      reference: string;
      payload: InvoicePayPayload;
    }) => creditCardService.payInvoice(cardId, reference, payload),
    onSuccess: (_data, { cardId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.creditCards(orgId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.invoices(orgId, cardId) });
      queryClient.invalidateQueries({ queryKey: ['invoice', orgId, cardId] });
      queryClient.invalidateQueries({ queryKey: ['transactions', orgId] });
      queryClient.invalidateQueries({ queryKey: ['payables', orgId] });
      queryClient.invalidateQueries({ queryKey: queryKeys.dashboard(orgId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.accounts(orgId) });
      toast.success('Fatura paga com sucesso.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}
