import { api, unwrap, unwrapPaginated } from '@/lib/axios';
import type {
  AdminListFilters,
  AdminOrg,
  AdminStats,
  AdminUser,
  ApiEnvelope,
} from '@/types';

/** Builds query params, dropping empty/undefined values. */
function listParams(filters: AdminListFilters) {
  const params: Record<string, string | number> = {};
  if (filters.search) params.search = filters.search;
  if (filters.page) params.page = filters.page;
  if (filters.per_page) params.per_page = filters.per_page;
  return params;
}

export const adminService = {
  // Stats --------------------------------------------------------------------
  stats: () => unwrap<AdminStats>(api.get<ApiEnvelope<AdminStats>>('/admin/stats')),

  // Organizations (read-only) ------------------------------------------------
  listOrganizations: (filters: AdminListFilters) =>
    unwrapPaginated<AdminOrg>(
      api.get<ApiEnvelope<AdminOrg[]>>('/admin/organizations', {
        params: listParams(filters),
      }),
    ),

  // Users --------------------------------------------------------------------
  listUsers: (filters: AdminListFilters) =>
    unwrapPaginated<AdminUser>(
      api.get<ApiEnvelope<AdminUser[]>>('/admin/users', {
        params: listParams(filters),
      }),
    ),

  disableUser: (id: string) =>
    unwrap<AdminUser>(api.post<ApiEnvelope<AdminUser>>(`/admin/users/${id}/disable`, {})),

  enableUser: (id: string) =>
    unwrap<AdminUser>(api.post<ApiEnvelope<AdminUser>>(`/admin/users/${id}/enable`, {})),

  /** Super-admin: set a new password for a user (forces change on next login). */
  resetUserPassword: (userId: string, new_password: string) =>
    unwrap<{ message: string }>(
      api.post<ApiEnvelope<{ message: string }>>(`/admin/users/${userId}/reset-password`, {
        new_password,
      }),
    ),
};
