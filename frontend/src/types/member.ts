import type { OrgRole } from './auth';

export interface MemberUser {
  id: string;
  name: string;
  email: string;
  avatar_url?: string;
}

export interface Member {
  id: string;
  user: MemberUser;
  role: OrgRole;
}

export interface MemberRolePayload {
  role: OrgRole;
}
