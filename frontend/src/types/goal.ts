export interface Goal {
  id: string;
  name: string;
  /** Target amount in cents (int64). */
  target_amount: number;
  /** Currently saved amount in cents (int64). */
  current_amount: number;
  /** Optional ISO deadline date. */
  deadline?: string;
  color: string;
  /** Progress ratio in the 0..1 range. */
  progress: number;
}

export interface GoalPayload {
  name: string;
  target_amount: number;
  current_amount: number;
  deadline?: string;
  color: string;
}
