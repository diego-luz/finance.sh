import { api, unwrap } from '@/lib/axios';
import type { ApiEnvelope, Dashboard } from '@/types';

export const dashboardService = {
  get: () => unwrap<Dashboard>(api.get<ApiEnvelope<Dashboard>>('/dashboard')),
};
