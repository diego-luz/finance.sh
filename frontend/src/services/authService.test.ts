import { type AxiosResponse, type InternalAxiosRequestConfig } from 'axios';
import { afterEach, expect, it } from 'vitest';
import { api } from '@/lib/axios';
import { useAuthStore } from '@/stores/authStore';
import { authService } from './authService';

afterEach(() => useAuthStore.getState().logout());

it('changePassword keeps the current session (sends keep_refresh_token)', async () => {
  useAuthStore.setState({ accessToken: 'A', refreshToken: 'R-atual' });
  let corpo: Record<string, unknown> = {};
  api.defaults.adapter = async (config: InternalAxiosRequestConfig) => {
    corpo = JSON.parse(String(config.data));
    return { status: 200, statusText: 'OK', data: { success: true, data: {} }, headers: {}, config } as AxiosResponse;
  };
  await authService.changePassword({ current_password: 'a', new_password: 'b' });
  expect(corpo).toEqual({ current_password: 'a', new_password: 'b', keep_refresh_token: 'R-atual' });
});
