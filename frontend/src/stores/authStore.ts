import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { AuthResponse, Organization, User } from '@/types';

const STORAGE_KEY = 'finance-sh-auth';

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
      name: STORAGE_KEY,
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


/**
 * Tokens as last written to localStorage by ANY tab. The in-memory store is
 * only read from storage on page load, so a tab that slept while another one
 * rotated the refresh token would otherwise present the rotated one again —
 * which the backend treats as a stolen token and revokes the whole session.
 */
function readPersistedTokens(): { accessToken: string | null; refreshToken: string | null } | null {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    const state = JSON.parse(raw)?.state;
    return { accessToken: state?.accessToken ?? null, refreshToken: state?.refreshToken ?? null };
  } catch {
    return null;
  }
}

// Another tab logged in, refreshed or logged out: follow it.
if (typeof window !== 'undefined') {
  window.addEventListener('storage', (e) => {
    if (e.key === STORAGE_KEY) void useAuthStore.persist.rehydrate();
  });
}

/** Non-hook accessors for use outside React (e.g. axios interceptors). */
export const authStore = {
  readPersistedTokens,
  /** Flag the session as needing a password change (server said so). */
  requirePasswordChange: () => {
    const user = useAuthStore.getState().user;
    if (user && !user.must_change_password) {
      useAuthStore.getState().setUser({ ...user, must_change_password: true });
    }
  },
  getAccessToken: () => useAuthStore.getState().accessToken,
  getRefreshToken: () => useAuthStore.getState().refreshToken,
  getCurrentOrgId: () => useAuthStore.getState().currentOrgId,
  setTokens: (a: string, r: string) => useAuthStore.getState().setTokens(a, r),
  logout: () => useAuthStore.getState().logout(),
  isAuthenticated: () => Boolean(useAuthStore.getState().accessToken),
};
