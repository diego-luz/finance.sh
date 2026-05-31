import type { Category, CategoryKind } from './category';

/** How a categorization rule matches a transaction description. */
export type MatchType = 'contains' | 'prefix' | 'regex';

/** A single automatic-categorization rule. */
export interface CategoryRule {
  id: string;
  /** Keyword or pattern matched against the transaction description. */
  pattern: string;
  match_type: MatchType;
  category_id: string;
  /** Expanded category summary returned by the backend. */
  category: {
    id: string;
    name: string;
    color: string;
    icon: string;
    kind: CategoryKind;
  };
  /** Higher priority rules win when several match. */
  priority: number;
  active: boolean;
}

/** Body for creating/updating a categorization rule. */
export interface CategoryRulePayload {
  pattern: string;
  category_id: string;
  match_type?: MatchType;
  priority?: number;
  active?: boolean;
}

/** Source of an automatic suggestion. */
export type SuggestSource = 'rule' | 'history' | 'none';

/** Response of GET /categorization/suggest. */
export interface SuggestResponse {
  category_id?: string;
  source: SuggestSource;
  category?: Pick<Category, 'id' | 'name' | 'color' | 'icon' | 'kind'>;
}
