import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { contactService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';
import { queryKeys } from './queryKeys';
import { ApiRequestError } from '@/lib/axios';
import type { ContactPayload } from '@/types';

export function useContacts() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.contacts(orgId),
    queryFn: () => contactService.list(),
    enabled: Boolean(orgId),
  });
}

export function useCreateContact() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: (payload: ContactPayload) => contactService.create(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.contacts(orgId) });
      toast.success('Contato criado.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useUpdateContact() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: ContactPayload }) =>
      contactService.update(id, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.contacts(orgId) });
      toast.success('Contato atualizado.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useDeleteContact() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: (id: string) => contactService.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.contacts(orgId) });
      toast.success('Contato removido.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}
