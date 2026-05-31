import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { memberService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';
import { queryKeys } from './queryKeys';
import { ApiRequestError } from '@/lib/axios';
import type { InvitationPayload, MemberRolePayload } from '@/types';

export function useMembers() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.members(orgId),
    queryFn: () => memberService.list(),
    enabled: Boolean(orgId),
  });
}

export function useUpdateMemberRole() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: MemberRolePayload }) =>
      memberService.updateRole(id, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.members(orgId) });
      toast.success('Função atualizada.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useRemoveMember() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: (id: string) => memberService.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.members(orgId) });
      toast.success('Membro removido.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useInvitations() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.invitations(orgId),
    queryFn: () => memberService.listInvitations(),
    enabled: Boolean(orgId),
  });
}

export function useCreateInvitation() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: (payload: InvitationPayload) => memberService.invite(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.invitations(orgId) });
      toast.success('Convite enviado.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}

export function useRevokeInvitation() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: (id: string) => memberService.revokeInvitation(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.invitations(orgId) });
      toast.success('Convite revogado.');
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}
