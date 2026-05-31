import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { AuthResponse, Organization, User } from '@/types';

interface AuthState {
  user: User | null;
  organizations: Organization[];
  currentOrgId: string | null;
  accessToken: string | null;
  refreshToken: string | null;

  /** Set full auth from a login/register/refresh response. */
  setAuth: (res: AuthResponse) => void;
  /** Update just tokens (used by the refresh flow). */
  setTokens: (accessToken: string, refreshToken: string) => void;
  /** Replace the list of organizations (from /me). */
  setOrganizations: (orgs: Organization[]) => void;
  /** Update the user profile. */
  setUser: (user: User) => void;
  /** Switch the active organization. */
  setOrg: (orgId: string) => void;
  /** Merge updated fields (e.g. name/currency) into an organization. */
  setOrgDetails: (org: Pick<Organization, 'id'> & Partial<Organization>) => void;
  /** Clear everything. */
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      organizations: [],
      currentOrgId: null,
      accessToken: null,
      refreshToken: null,

      setAuth: (res) =>
        set((state) => {
          const orgs = res.organization
            ? mergeOrg(state.organizations, res.organization)
            : state.organizations;
          return {
            user: res.user,
            organizations: orgs,
            currentOrgId: res.organization?.id ?? state.currentOrgId,
            accessToken: res.access_token,
            refreshToken: res.refresh_token,
          };
        }),

      setTokens: (accessToken, refreshToken) =>
        set({ accessToken, refreshToken }),

      setOrganizations: (orgs) =>
        set((state) => ({
          organizations: orgs,
          currentOrgId:
            state.currentOrgId && orgs.some((o) => o.id === state.currentOrgId)
              ? state.currentOrgId
              : (orgs[0]?.id ?? null),
        })),

      setUser: (user) => set({ user }),

      setOrg: (orgId) => {
        if (get().organizations.some((o) => o.id === orgId)) {
          set({ currentOrgId: orgId });
        }
      },

      setOrgDetails: (org) =>
        set((state) => ({
          organizations: state.organizations.map((o) =>
            o.id === org.id ? { ...o, ...org } : o,
          ),
        })),

      logout: () =>
        set({
          user: null,
          organizations: [],
          currentOrgId: null,
          accessToken: null,
          refreshToken: null,
        }),
    }),
    {
      name: 'finance-sh-auth',
      partialize: (s) => ({
        user: s.user,
        organizations: s.organizations,
        currentOrgId: s.currentOrgId,
        accessToken: s.accessToken,
        refreshToken: s.refreshToken,
      }),
    },
  ),
);

function mergeOrg(orgs: Organization[], org: Organization): Organization[] {
  const idx = orgs.findIndex((o) => o.id === org.id);
  if (idx === -1) return [...orgs, org];
  const next = [...orgs];
  next[idx] = org;
  return next;
}

/** Non-hook accessors for use outside React (e.g. axios interceptors). */
export const authStore = {
  getAccessToken: () => useAuthStore.getState().accessToken,
  getRefreshToken: () => useAuthStore.getState().refreshToken,
  getCurrentOrgId: () => useAuthStore.getState().currentOrgId,
  setTokens: (a: string, r: string) => useAuthStore.getState().setTokens(a, r),
  logout: () => useAuthStore.getState().logout(),
  isAuthenticated: () => Boolean(useAuthStore.getState().accessToken),
};
