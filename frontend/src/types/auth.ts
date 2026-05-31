export interface User {
  id: string;
  name: string;
  email: string;
  email_verified: boolean;
  /** Whether the account has 2FA enabled (may be omitted by older backends). */
  two_factor_enabled?: boolean;
  avatar_url?: string;
  /**
   * Platform super-admin flag (distinct from per-org RBAC). Comes from /me and
   * the auth response. Gates the /admin back-office. Omitted by older backends.
   */
  super_admin?: boolean;
  /**
   * Forces a password change on the next authenticated request flow. Set by
   * the backend after admin provisioning; cleared by POST /me/change-password.
   * Omitted by older backends (treated as false).
   */
  must_change_password?: boolean;
}

export type OrgRole = 'owner' | 'admin' | 'member' | 'viewer' | string;

export interface Organization {
  id: string;
  name: string;
  slug: string;
  currency: string;
  role: OrgRole;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: User;
  organization: Organization;
}

/** Returned by /auth/login when 2FA is enabled for the account. */
export interface MfaChallenge {
  mfa_required: true;
  mfa_token: string;
}

/** Login can resolve to a full session or a pending 2FA challenge. */
export type LoginResult = AuthResponse | MfaChallenge;

/** Narrowing helper: true when the login response is a 2FA challenge. */
export function isMfaChallenge(res: LoginResult): res is MfaChallenge {
  return (res as MfaChallenge).mfa_required === true;
}

export interface MeResponse {
  user: User;
  organizations: Organization[];
}

export interface LoginPayload {
  email: string;
  password: string;
}

export interface RegisterPayload {
  name: string;
  email: string;
  password: string;
  organization_name: string;
  accepted_terms: boolean;
  /** 3-letter currency code for the new organization (default BRL). */
  currency?: string;
}

export interface RefreshPayload {
  refresh_token: string;
}

// ---------------------------------------------------------------------------
// Email verification
// ---------------------------------------------------------------------------
export interface VerifyEmailPayload {
  token: string;
}

export interface ResendVerificationPayload {
  email: string;
}

// ---------------------------------------------------------------------------
// Two-factor authentication
// ---------------------------------------------------------------------------
export interface TwoFactorVerifyPayload {
  mfa_token: string;
  code: string;
}

export interface TwoFactorSetupResponse {
  secret: string;
  otpauth_url: string;
}

export interface TwoFactorEnableResponse {
  recovery_codes: string[];
}

// ---------------------------------------------------------------------------
// LGPD / account privacy
// ---------------------------------------------------------------------------
export interface DeleteAccountPayload {
  password: string;
}

// ---------------------------------------------------------------------------
// Forced / self-service password change (authenticated)
// ---------------------------------------------------------------------------
export interface ChangePasswordPayload {
  current_password: string;
  new_password: string;
}
