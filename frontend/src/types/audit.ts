/** The user who performed an audited action, when attributable. */
export interface AuditUser {
  id: string;
  name: string;
  email: string;
}

/**
 * A single audit-log entry. `action` is the HTTP method (POST/PUT/DELETE),
 * `entity` is a resource path segment (e.g. "transactions"), and `entity_id`
 * is the affected record's uuid (may be empty for collection-level actions).
 */
export interface AuditLog {
  id: string;
  /** ISO date-time string. */
  created_at: string;
  /** HTTP method: POST | PUT | DELETE. */
  action: string;
  /** Resource path segment, e.g. "transactions", "accounts", "members". */
  entity: string;
  /** Affected record uuid (may be empty). */
  entity_id: string;
  /** Originating IP address. */
  ip: string;
  /** Actor, or null when performed by the system / unauthenticated. */
  user: AuditUser | null;
  /** Free-form extra context captured at write time. */
  metadata?: Record<string, unknown> | null;
}

/** Query filters for GET /audit-logs. */
export interface AuditFilters {
  /** HTTP method filter (POST | PUT | DELETE). */
  action?: string;
  /** Entity path segment filter. */
  entity?: string;
  /** Restrict to actions performed by a given user. */
  user_id?: string;
  /** ISO date string — inclusive lower bound. */
  from?: string;
  /** ISO date string — inclusive upper bound. */
  to?: string;
  page?: number;
  per_page?: number;
}
