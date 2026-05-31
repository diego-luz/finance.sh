import { api, unwrap } from '@/lib/axios';
import type {
  AcceptInvitationPayload,
  ApiEnvelope,
  Invitation,
  InvitationPayload,
  Member,
  MemberRolePayload,
} from '@/types';

export const memberService = {
  // Members ----------------------------------------------------------------
  list: () => unwrap<Member[]>(api.get<ApiEnvelope<Member[]>>('/members')),

  updateRole: (id: string, payload: MemberRolePayload) =>
    unwrap<Member>(api.put<ApiEnvelope<Member>>(`/members/${id}`, payload)),

  remove: (id: string) => api.delete(`/members/${id}`),

  // Invitations ------------------------------------------------------------
  listInvitations: () =>
    unwrap<Invitation[]>(api.get<ApiEnvelope<Invitation[]>>('/invitations')),

  invite: (payload: InvitationPayload) =>
    unwrap<Invitation>(api.post<ApiEnvelope<Invitation>>('/invitations', payload)),

  revokeInvitation: (id: string) => api.delete(`/invitations/${id}`),

  acceptInvitation: (payload: AcceptInvitationPayload) =>
    unwrap<unknown>(api.post<ApiEnvelope<unknown>>('/invitations/accept', payload)),
};
