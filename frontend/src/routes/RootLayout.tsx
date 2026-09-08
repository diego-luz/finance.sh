import { Outlet } from 'react-router-dom';
import { Spinner } from '@/components/ui';
import { SetupGate } from './ProtectedRoute';

/**
 * Route-level components used by the router config. They live here rather than
 * in routes/index.tsx so that module exports only the `router` constant and this
 * one exports only components — the split React Fast Refresh needs to hot-reload
 * either file.
 */

/** Centered spinner fallback used while a lazy route chunk loads. */
export function RouteFallback() {
  return (
    <div className="flex min-h-[60vh] items-center justify-center">
      <Spinner />
    </div>
  );
}

/**
 * Root layout: the SetupGate runs on EVERY route so the first-run wizard takes
 * priority over any deep-linked URL. The gate either renders <Outlet /> (the
 * matched route) or redirects to /setup or /login as required.
 */
export function RootLayout() {
  return (
    <SetupGate>
      <Outlet />
    </SetupGate>
  );
}
