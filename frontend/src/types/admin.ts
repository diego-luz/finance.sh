/**
 * Platform super-admin back-office types. These map the `/admin/*` endpoints,
 * which are platform-wide (no per-org tenant scoping) and only reachable by a
 * user whose `super_admin` flag is set. The back-office is read-only views plus
 * a couple of user-safety operations (disable, reset password).
 */

/** Aggregate counters returned by GET /admin/stats. */
export interface AdminStats {
  organizations: number;
  users: number;
  transactions: number;
}

/** A single row of GET /admin/organizations. */
export interface AdminOrg {
  id: string;
  name: string;
  slug: string;
  currency: string;
  owner: {
    id: string;
    name: string;
    email: string;
  };
  members: number;
  transactions: number;
  created_at: string;
}

/** A single row of GET /admin/users. */
export interface AdminUser {
  id: string;
  name: string;
  email: string;
  super_admin: boolean;
  disabled: boolean;
  email_verified: boolean;
  created_at: string;
  /** Organizations the user belongs to. */
  organizations: string[];
}

/** Shared filters for the paginated admin list endpoints. */
export interface AdminListFilters {
  search?: string;
  page?: number;
  per_page?: number;
}
