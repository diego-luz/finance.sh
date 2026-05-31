import { Navigate, useLocation } from 'react-router-dom';
import type { ReactNode } from 'react';
import { useAuthStore } from '@/stores/authStore';
import { useSetupStatus } from '@/hooks/useSetupStatus';
import { Spinner } from '@/components/ui';

/**
 * Routes that remain reachable while the platform is uninitialized. Anything
 * else is bounced to `/setup`. Public legal pages stay accessible so support
 * links in the marketing site keep working pre-bootstrap.
 */
const SETUP_ALLOWLIST = new Set<string>([
  '/setup',
  '/termos',
  '/privacidade',
]);

/**
 * Reads the public `GET /setup/status` probe and enforces the first-run wizard.
 *
 * Rules:
 *   - `needs_setup === true` and current path NOT in the allow-list → /setup
 *   - `needs_setup === false` and current path IS /setup            → /login
 *
 * Render-blocks until the probe resolves so we never flash a protected route
 * just to redirect to /setup a tick later. Fail-open: if the probe errors,
 * proceeds with the requested route so the app stays usable.
 */
export function SetupGate({ children }: { children: ReactNode }) {
  const { data, isLoading, isError } = useSetupStatus();
  const location = useLocation();

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-100 dark:bg-ink-base">
        <Spinner />
      </div>
    );
  }

  // Fail-open: a backend that can't answer the probe shouldn't trap the user
  // on a blank screen. Render whatever was matched and let downstream
  // protected routes apply their own auth checks.
  if (isError || !data) {
    return <>{children}</>;
  }

  const onSetup = location.pathname === '/setup';

  if (data.needs_setup) {
    if (!SETUP_ALLOWLIST.has(location.pathname)) {
      return <Navigate to="/setup" replace />;
    }
    return <>{children}</>;
  }

  // Initialized: don't let anyone re-run the wizard.
  if (onSetup) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

/**
 * Redirects to /login when there is no access token. Also enforces the
 * forced-password-change flow: whenever the hydrated user carries
 * `must_change_password = true`, every protected route except
 * `/change-password` itself bounces to that page.
 */
export function ProtectedRoute({ children }: { children: ReactNode }) {
  const accessToken = useAuthStore((s) => s.accessToken);
  const user = useAuthStore((s) => s.user);
  const location = useLocation();

  if (!accessToken) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }

  // Forced password change takes precedence over the requested route. We
  // gate on `user` being already hydrated so the AuthProvider splash window
  // (when /me is in-flight) doesn't bounce around before the flag is known.
  if (
    user?.must_change_password === true &&
    location.pathname !== '/change-password'
  ) {
    return <Navigate to="/change-password" replace />;
  }

  return <>{children}</>;
}

/** Redirects already-authenticated users away from public auth pages. */
export function PublicOnlyRoute({ children }: { children: ReactNode }) {
  const accessToken = useAuthStore((s) => s.accessToken);
  const user = useAuthStore((s) => s.user);

  if (accessToken) {
    // If the session is mid-forced-password-change, route them to the
    // change-password page instead of the dashboard.
    if (user?.must_change_password === true) {
      return <Navigate to="/change-password" replace />;
    }
    return <Navigate to="/" replace />;
  }
  return <>{children}</>;
}
