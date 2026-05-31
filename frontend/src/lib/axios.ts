import axios, {
  AxiosError,
  type AxiosInstance,
  type AxiosResponse,
  type InternalAxiosRequestConfig,
} from 'axios';
import { authStore } from '@/stores/authStore';
import type { ApiEnvelope, ApiError } from '@/types';

const BASE_URL =
  import.meta.env.VITE_API_URL ?? 'http://localhost:8080/api/v1';

/**
 * A thrown API error. Carries the backend error code/message/fields plus the
 * HTTP status so callers (forms, toasts) can react appropriately.
 */
export class ApiRequestError extends Error {
  code: string;
  fields?: Record<string, string>;
  status?: number;

  constructor(error: ApiError, status?: number) {
    super(error.message || 'Erro inesperado');
    this.name = 'ApiRequestError';
    this.code = error.code;
    this.fields = error.fields;
    this.status = status;
  }
}

export const api: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  headers: { 'Content-Type': 'application/json' },
  timeout: 30_000,
});

// ---------------------------------------------------------------------------
// Request interceptor: inject Bearer + tenant header.
// ---------------------------------------------------------------------------
api.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = authStore.getAccessToken();
  const orgId = authStore.getCurrentOrgId();
  if (token) {
    config.headers.set('Authorization', `Bearer ${token}`);
  }
  if (orgId) {
    config.headers.set('X-Organization-ID', orgId);
  }
  return config;
});

// ---------------------------------------------------------------------------
// Refresh-token queue. While a refresh is in flight, concurrent 401s wait.
// ---------------------------------------------------------------------------
let isRefreshing = false;
let refreshWaiters: Array<(token: string | null) => void> = [];

function notifyWaiters(token: string | null) {
  refreshWaiters.forEach((cb) => cb(token));
  refreshWaiters = [];
}

interface RetriableConfig extends InternalAxiosRequestConfig {
  _retried?: boolean;
}

async function performRefresh(): Promise<string | null> {
  const refreshToken = authStore.getRefreshToken();
  if (!refreshToken) return null;
  try {
    // Use a bare axios call to avoid recursive interceptors.
    const res = await axios.post<ApiEnvelope<{ access_token: string; refresh_token: string }>>(
      `${BASE_URL}/auth/refresh`,
      { refresh_token: refreshToken },
      { headers: { 'Content-Type': 'application/json' } },
    );
    if (res.data?.success && res.data.data) {
      const { access_token, refresh_token } = res.data.data;
      authStore.setTokens(access_token, refresh_token);
      return access_token;
    }
    return null;
  } catch {
    return null;
  }
}

function forceLogout() {
  authStore.logout();
  // Redirect to login, preserving nothing sensitive.
  if (window.location.pathname !== '/login') {
    window.location.assign('/login');
  }
}

// ---------------------------------------------------------------------------
// Response interceptor: unwrap envelope; handle 401 with one-time refresh.
// ---------------------------------------------------------------------------
api.interceptors.response.use(
  (response: AxiosResponse<ApiEnvelope<unknown>>) => {
    const envelope = response.data;
    // Some endpoints (e.g. logout 204) may return no body.
    if (envelope == null) {
      return response;
    }
    if (envelope.success === false && envelope.error) {
      return Promise.reject(new ApiRequestError(envelope.error, response.status));
    }
    // Attach meta to the response for list endpoints, then return the
    // raw axios response so service layer can read `.data` (the envelope).
    return response;
  },
  async (error: AxiosError<ApiEnvelope<unknown>>) => {
    const original = error.config as RetriableConfig | undefined;
    const status = error.response?.status;

    // Attempt a single refresh on 401 (skip the refresh endpoint itself).
    const isAuthEndpoint =
      original?.url?.includes('/auth/refresh') ||
      original?.url?.includes('/auth/login') ||
      original?.url?.includes('/auth/register');

    if (status === 401 && original && !original._retried && !isAuthEndpoint) {
      original._retried = true;

      if (isRefreshing) {
        // Queue until the in-flight refresh resolves.
        return new Promise((resolve, reject) => {
          refreshWaiters.push((token) => {
            if (token) {
              original.headers.set('Authorization', `Bearer ${token}`);
              resolve(api(original));
            } else {
              reject(error);
            }
          });
        });
      }

      isRefreshing = true;
      const newToken = await performRefresh();
      isRefreshing = false;
      notifyWaiters(newToken);

      if (newToken) {
        original.headers.set('Authorization', `Bearer ${newToken}`);
        return api(original);
      }
      forceLogout();
    }

    // Normalize the error to ApiRequestError when the envelope is present.
    const envelope = error.response?.data;
    if (envelope && typeof envelope === 'object' && 'error' in envelope && envelope.error) {
      return Promise.reject(new ApiRequestError(envelope.error, status));
    }

    return Promise.reject(
      new ApiRequestError(
        {
          code: error.code ?? 'network_error',
          message:
            error.message === 'Network Error'
              ? 'Não foi possível conectar ao servidor.'
              : (error.message ?? 'Erro inesperado'),
        },
        status,
      ),
    );
  },
);

/**
 * Helper that unwraps the success envelope and returns `data`.
 * Throws ApiRequestError on failure (already normalized by the interceptor).
 */
export async function unwrap<T>(
  promise: Promise<AxiosResponse<ApiEnvelope<T>>>,
): Promise<T> {
  const res = await promise;
  const envelope = res.data;
  if (!envelope || envelope.success === false) {
    throw new ApiRequestError(
      envelope?.error ?? { code: 'unknown', message: 'Erro inesperado' },
      res.status,
    );
  }
  return envelope.data as T;
}

/** Like {@link unwrap} but also returns pagination meta. */
export async function unwrapPaginated<T>(
  promise: Promise<AxiosResponse<ApiEnvelope<T[]>>>,
): Promise<{ data: T[]; meta: { page: number; per_page: number; total: number; pages: number } }> {
  const res = await promise;
  const envelope = res.data;
  if (!envelope || envelope.success === false) {
    throw new ApiRequestError(
      envelope?.error ?? { code: 'unknown', message: 'Erro inesperado' },
      res.status,
    );
  }
  const meta = envelope.meta ?? {};
  return {
    data: envelope.data ?? [],
    meta: {
      page: Number(meta.page ?? 1),
      per_page: Number(meta.per_page ?? (envelope.data?.length ?? 0)),
      total: Number(meta.total ?? (envelope.data?.length ?? 0)),
      pages: Number(meta.pages ?? 1),
    },
  };
}
