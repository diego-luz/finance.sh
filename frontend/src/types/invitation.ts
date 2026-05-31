import type { OrgRole } from './auth';

export interface Invitation {
  id: string;
  email: string;
  role: OrgRole;
  accepted: boolean;
  created_at: string;
  /** Returned only on creation (email delivery is not wired yet). */
  token?: string;
}

export interface InvitationPayload {
  email: string;
  role: OrgRole;
}

export interface AcceptInvitationPayload {
  token: string;
}
