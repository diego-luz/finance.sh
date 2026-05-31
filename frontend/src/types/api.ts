/** Standard API response envelope returned by the Go backend on every call. */
export interface ApiError {
  code: string;
  message: string;
  fields?: Record<string, string>;
}

export interface ApiMeta {
  page?: number;
  per_page?: number;
  total?: number;
  pages?: number;
  [key: string]: unknown;
}

export interface ApiEnvelope<T> {
  success: boolean;
  data?: T;
  error?: ApiError;
  meta?: ApiMeta;
}

/** A list payload unwrapped with its pagination meta preserved. */
export interface Paginated<T> {
  data: T[];
  meta: {
    page: number;
    per_page: number;
    total: number;
    pages: number;
  };
}
