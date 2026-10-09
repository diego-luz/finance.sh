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
/**
 * Outcome of a refresh: a new access token, a definitive "the session is
 * over" (the server rejected the refresh token), or a transient failure
 * (429, offline, 5xx) that must NOT log the user out.
 */
type RefreshResult = { token: string } | { ended: true } | { transient: true };

let refreshing: Promise<RefreshResult> | null = null;

interface RetriableConfig extends InternalAxiosRequestConfig {
  _retried?: boolean;
}

/** Seconds since epoch at which a JWT expires (0 when unreadable). */
function jwtExp(token: string): number {
  try {
    const payload = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/');
    return Number(JSON.parse(atob(payload)).exp) || 0;
  } catch {
    return 0;
  }
}

async function refreshOnce(failedToken: string | null): Promise<RefreshResult> {
  // Another tab may already have rotated the tokens: use theirs instead of
  // presenting the old refresh token again (the backend revokes the whole
  // session when a rotated refresh token comes back).
  const stored = authStore.readPersistedTokens();
  if (
    stored?.accessToken &&
    stored.refreshToken &&
    stored.accessToken !== failedToken &&
    jwtExp(stored.accessToken) > Date.now() / 1000 + 10
  ) {
    authStore.setTokens(stored.accessToken, stored.refreshToken);
    return { token: stored.accessToken };
  }
  const refreshToken = stored?.refreshToken ?? authStore.getRefreshToken();
  if (!refreshToken) return { ended: true };
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
      return { token: access_token };
    }
    return { ended: true };
  } catch (e) {
    const status = (e as AxiosError).response?.status;
    // Only a rejected refresh token ends the session.
    return status === 400 || status === 401 ? { ended: true } : { transient: true };
  }
}

/**
 * One refresh at a time per tab, and across tabs when the browser has Web
 * Locks: the second tab waits, then finds the first tab's tokens in storage.
 */
function performRefresh(failedToken: string | null): Promise<RefreshResult> {
  if (!refreshing) {
    const run = () => refreshOnce(failedToken);
    const locked =
      typeof navigator !== 'undefined' && navigator.locks?.request
        ? (navigator.locks.request('finance-sh-refresh', run) as Promise<unknown>).then(
            (r) => r as RefreshResult,
          )
        : run();
    refreshing = locked.finally(() => {
      refreshing = null;
    });
  }
  return refreshing;
}

function forceLogout() {
  authStore.logout();
  // Leave the page so nothing financial stays in memory (react-query cache).
  if (window.location.pathname !== '/login') {
    window.location.assign('/login');
  } else {
    window.location.reload();
  }
}

/** The backend error code of a failed response, when it sent one. */
function errorCode(error: AxiosError<ApiEnvelope<unknown>>): string | undefined {
  const data = error.response?.data;
  return data && typeof data === 'object' && 'error' in data ? data.error?.code : undefined;
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

    // Account states the server enforces on every request.
    const code = errorCode(error);
    if (status === 403 && code === 'account_disabled') {
      forceLogout();
    } else if (status === 403 && code === 'must_change_password') {
      // ProtectedRoute sends the user to /change-password once flagged.
      authStore.requirePasswordChange();
    }

    // A 401 only means "token expired" when the code says so: a wrong
    // current password (change-password, delete account) is also a 401 and
    // must not trigger a refresh plus a resend of the same wrong password.
    const tokenProblem = code === undefined || code === 'unauthorized' || code === 'invalid_token';
    if (status === 401 && tokenProblem && original && !original._retried && !isAuthEndpoint) {
      original._retried = true;
      const failedToken =
        String(original.headers.get('Authorization') ?? '').replace(/^Bearer\s+/, '') || null;
      const result = await performRefresh(failedToken);
      if ('token' in result) {
        original.headers.set('Authorization', `Bearer ${result.token}`);
        return api(original);
      }
      if ('ended' in result) forceLogout();
      // transient: keep the session; the request fails and can be retried
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
