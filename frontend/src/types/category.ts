export type CategoryKind = 'income' | 'expense';

export interface Category {
  id: string;
  name: string;
  kind: CategoryKind;
  color: string;
  icon: string;
}

export interface CategoryPayload {
  name: string;
  kind: CategoryKind;
  color: string;
  icon: string;
}
