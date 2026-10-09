import axios, { AxiosError, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { api } from './axios';
import { useAuthStore } from '@/stores/authStore';

/** An unsigned JWT whose payload only carries `exp` (the client never verifies). */
function jwt(expInSeconds: number, tag = ''): string {
  const b64 = (o: object) => btoa(JSON.stringify(o)).replace(/=+$/, '');
  return `${b64({ alg: 'none' })}.${b64({ exp: Math.floor(Date.now() / 1000) + expInSeconds, tag })}.x`;
}

type Reply = { status: number; data?: unknown };
type Handler = (config: InternalAxiosRequestConfig) => Reply;

/** Fake server: every request (api and the bare refresh call) goes through it. */
function serve(handler: Handler) {
  const calls: InternalAxiosRequestConfig[] = [];
  const adapter = async (config: InternalAxiosRequestConfig): Promise<AxiosResponse> => {
    calls.push(config);
    const { status, data } = handler(config);
    const response = { status, statusText: String(status), data, headers: {}, config } as AxiosResponse;
    if (status >= 400) {
      throw new AxiosError(`status ${status}`, String(status), config, null, response);
    }
    return response;
  };
  api.defaults.adapter = adapter;
  axios.defaults.adapter = adapter;
  return calls;
}

const ok = (data: unknown): Reply => ({ status: 200, data: { success: true, data } });
const fail = (status: number, code: string): Reply => ({
  status,
  data: { success: false, error: { code, message: code } },
});
const isRefresh = (c: InternalAxiosRequestConfig) => String(c.url).endsWith('/auth/refresh');
const bearer = (c: InternalAxiosRequestConfig) => String(c.headers?.Authorization ?? '');

function login(access: string, refresh: string) {
  useAuthStore.setState({
    accessToken: access,
    refreshToken: refresh,
    user: { id: 'u1', name: 'U', email: 'u@x', must_change_password: false } as never,
  });
}

beforeEach(() => {
  localStorage.clear();
  vi.spyOn(window.location, 'assign').mockImplementation(() => {});
  vi.spyOn(window.location, 'reload').mockImplementation(() => {});
});
afterEach(() => {
  vi.restoreAllMocks();
  useAuthStore.getState().logout();
});

describe('refresh on 401', () => {
  it('refreshes once and retries with the new token', async () => {
    const velho = jwt(-60, 'velho');
    const novo = jwt(900, 'novo');
    login(velho, 'R0');
    const calls = serve((c) =>
      isRefresh(c)
        ? ok({ access_token: novo, refresh_token: 'R1' })
        : bearer(c).endsWith(novo)
          ? ok({ fine: true })
          : fail(401, 'unauthorized'),
    );
    const res = await api.get('/accounts');
    expect(res.data.data).toEqual({ fine: true });
    expect(calls.filter(isRefresh)).toHaveLength(1);
    expect(useAuthStore.getState().refreshToken).toBe('R1');
  });

  it('a 429 on refresh keeps the session (no logout)', async () => {
    login(jwt(-60), 'R0');
    serve((c) => (isRefresh(c) ? fail(429, 'rate_limited') : fail(401, 'unauthorized')));
    await expect(api.get('/accounts')).rejects.toBeTruthy();
    expect(useAuthStore.getState().refreshToken).toBe('R0');
    expect(window.location.assign).not.toHaveBeenCalled();
  });

  it('a network error on refresh keeps the session', async () => {
    login(jwt(-60), 'R0');
    serve((c) => {
      if (isRefresh(c)) throw new AxiosError('Network Error', 'ERR_NETWORK', c);
      return fail(401, 'unauthorized');
    });
    await expect(api.get('/accounts')).rejects.toBeTruthy();
    expect(useAuthStore.getState().refreshToken).toBe('R0');
  });

  it('a rejected refresh token ends the session', async () => {
    login(jwt(-60), 'R0');
    serve((c) => (isRefresh(c) ? fail(401, 'invalid_token') : fail(401, 'unauthorized')));
    await expect(api.get('/accounts')).rejects.toBeTruthy();
    expect(useAuthStore.getState().accessToken).toBeNull();
    expect(window.location.assign).toHaveBeenCalledWith('/login');
  });

  it('a wrong current password does not trigger a refresh', async () => {
    login(jwt(900), 'R0');
    const calls = serve(() => fail(401, 'wrong_password'));
    await expect(api.post('/me/change-password', {})).rejects.toMatchObject({ code: 'wrong_password' });
    expect(calls.filter(isRefresh)).toHaveLength(0);
    expect(calls).toHaveLength(1); // not resent
    expect(useAuthStore.getState().refreshToken).toBe('R0');
  });

  it('uses tokens another tab already rotated instead of refreshing again', async () => {
    const velho = jwt(-60, 'velho');
    const daOutraAba = jwt(900, 'outra');
    login(velho, 'R0');
    // the other tab wrote R1 to the shared storage; this tab still holds R0
    localStorage.setItem(
      'finance-sh-auth',
      JSON.stringify({ state: { accessToken: daOutraAba, refreshToken: 'R1' }, version: 0 }),
    );
    const calls = serve((c) => (bearer(c).endsWith(daOutraAba) ? ok({ fine: true }) : fail(401, 'unauthorized')));
    const res = await api.get('/accounts');
    expect(res.data.data).toEqual({ fine: true });
    expect(calls.filter(isRefresh)).toHaveLength(0); // R0 never presented again
    expect(useAuthStore.getState().refreshToken).toBe('R1');
  });

  it('concurrent 401s share a single refresh', async () => {
    const novo = jwt(900, 'novo');
    login(jwt(-60), 'R0');
    const calls = serve((c) =>
      isRefresh(c)
        ? ok({ access_token: novo, refresh_token: 'R1' })
        : bearer(c).endsWith(novo)
          ? ok({})
          : fail(401, 'unauthorized'),
    );
    await Promise.all([api.get('/a'), api.get('/b'), api.get('/c')]);
    expect(calls.filter(isRefresh)).toHaveLength(1);
  });
});

describe('account states', () => {
  it('account_disabled logs out', async () => {
    login(jwt(900), 'R0');
    serve(() => fail(403, 'account_disabled'));
    await expect(api.get('/accounts')).rejects.toBeTruthy();
    expect(useAuthStore.getState().accessToken).toBeNull();
  });

  it('must_change_password flags the user so the router sends them to /change-password', async () => {
    login(jwt(900), 'R0');
    serve(() => fail(403, 'must_change_password'));
    await expect(api.get('/accounts')).rejects.toBeTruthy();
    expect(useAuthStore.getState().user?.must_change_password).toBe(true);
    expect(useAuthStore.getState().accessToken).not.toBeNull();
  });
});
