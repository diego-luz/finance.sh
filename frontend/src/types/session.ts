/** An active authentication session for the current user. */
export interface Session {
  id: string;
  user_agent: string;
  ip: string;
  /** ISO date string. */
  created_at: string;
  /** ISO date string. */
  expires_at: string;
}
