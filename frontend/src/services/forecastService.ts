import { api, unwrap } from '@/lib/axios';
import type { ApiEnvelope, Forecast } from '@/types';

export const forecastService = {
  get: (months: number) =>
    unwrap<Forecast>(
      api.get<ApiEnvelope<Forecast>>('/cashflow/forecast', { params: { months } }),
    ),
};
