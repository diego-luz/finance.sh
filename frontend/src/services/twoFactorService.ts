import { api, unwrap } from '@/lib/axios';
import type {
  ApiEnvelope,
  TwoFactorEnableResponse,
  TwoFactorSetupResponse,
} from '@/types';

/** Authenticated 2FA management (settings flows). */
export const twoFactorService = {
  /** Begins enrollment: returns a TOTP secret + otpauth:// URL for the QR. */
  setup: () =>
    unwrap<TwoFactorSetupResponse>(
      api.post<ApiEnvelope<TwoFactorSetupResponse>>('/me/2fa/setup'),
    ),

  /** Confirms enrollment with a current code; returns recovery codes. */
  enable: (code: string) =>
    unwrap<TwoFactorEnableResponse>(
      api.post<ApiEnvelope<TwoFactorEnableResponse>>('/me/2fa/enable', { code }),
    ),

  /** Disables 2FA after verifying a current code. */
  disable: (code: string) =>
    api.post<ApiEnvelope<unknown>>('/me/2fa/disable', { code }),
};
