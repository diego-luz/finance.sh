export type NotificationType =
  | 'info'
  | 'success'
  | 'warning'
  | 'error'
  | 'bill'
  | 'budget'
  | 'goal'
  | string;

export interface Notification {
  id: string;
  type: NotificationType;
  title: string;
  message: string;
  read: boolean;
  created_at: string;
}
