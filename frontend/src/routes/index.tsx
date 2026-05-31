import { Suspense, lazy, type ReactNode } from 'react';
import { createBrowserRouter, Navigate, Outlet } from 'react-router-dom';
import { DashboardLayout } from '@/layouts/DashboardLayout';
import { AdminLayout } from '@/layouts/AdminLayout';
import { ProtectedRoute, PublicOnlyRoute, SetupGate } from './ProtectedRoute';
import { Spinner } from '@/components/ui';

// Lazily-loaded pages. Each becomes its own chunk, trimming the initial bundle.
const LoginPage = lazy(() => import('@/pages/LoginPage').then((m) => ({ default: m.LoginPage })));
const SetupPage = lazy(() => import('@/pages/SetupPage').then((m) => ({ default: m.SetupPage })));
const RegisterPage = lazy(() =>
  import('@/pages/RegisterPage').then((m) => ({ default: m.RegisterPage })),
);
const ForgotPasswordPage = lazy(() =>
  import('@/pages/ForgotPasswordPage').then((m) => ({ default: m.ForgotPasswordPage })),
);
const ResetPasswordPage = lazy(() =>
  import('@/pages/ResetPasswordPage').then((m) => ({ default: m.ResetPasswordPage })),
);
const ChangePasswordPage = lazy(() =>
  import('@/pages/ChangePasswordPage').then((m) => ({ default: m.ChangePasswordPage })),
);
const VerifyEmailPage = lazy(() =>
  import('@/pages/VerifyEmailPage').then((m) => ({ default: m.VerifyEmailPage })),
);
const TermsPage = lazy(() => import('@/pages/TermsPage').then((m) => ({ default: m.TermsPage })));
const PrivacyPolicyPage = lazy(() =>
  import('@/pages/PrivacyPolicyPage').then((m) => ({ default: m.PrivacyPolicyPage })),
);
const DashboardPage = lazy(() =>
  import('@/pages/DashboardPage').then((m) => ({ default: m.DashboardPage })),
);
const TransactionsPage = lazy(() =>
  import('@/pages/TransactionsPage').then((m) => ({ default: m.TransactionsPage })),
);
const RecurrencesPage = lazy(() =>
  import('@/pages/RecurrencesPage').then((m) => ({ default: m.RecurrencesPage })),
);
const PayablesPage = lazy(() =>
  import('@/pages/PayablesPage').then((m) => ({ default: m.PayablesPage })),
);
const ReceivablesPage = lazy(() =>
  import('@/pages/ReceivablesPage').then((m) => ({ default: m.ReceivablesPage })),
);
const ContactsPage = lazy(() =>
  import('@/pages/ContactsPage').then((m) => ({ default: m.ContactsPage })),
);
const ForecastPage = lazy(() =>
  import('@/pages/ForecastPage').then((m) => ({ default: m.ForecastPage })),
);
const ReportsPage = lazy(() =>
  import('@/pages/ReportsPage').then((m) => ({ default: m.ReportsPage })),
);
const AccountsPage = lazy(() =>
  import('@/pages/AccountsPage').then((m) => ({ default: m.AccountsPage })),
);
const CreditCardsPage = lazy(() =>
  import('@/pages/CreditCardsPage').then((m) => ({ default: m.CreditCardsPage })),
);
const CategoriesPage = lazy(() =>
  import('@/pages/CategoriesPage').then((m) => ({ default: m.CategoriesPage })),
);
const TagsPage = lazy(() => import('@/pages/TagsPage').then((m) => ({ default: m.TagsPage })));
const CategorizationPage = lazy(() =>
  import('@/pages/CategorizationPage').then((m) => ({ default: m.CategorizationPage })),
);
const BudgetsPage = lazy(() =>
  import('@/pages/BudgetsPage').then((m) => ({ default: m.BudgetsPage })),
);
const GoalsPage = lazy(() => import('@/pages/GoalsPage').then((m) => ({ default: m.GoalsPage })));
const MembersPage = lazy(() =>
  import('@/pages/MembersPage').then((m) => ({ default: m.MembersPage })),
);
const AuditPage = lazy(() =>
  import('@/pages/AuditPage').then((m) => ({ default: m.AuditPage })),
);
const SettingsPage = lazy(() =>
  import('@/pages/SettingsPage').then((m) => ({ default: m.SettingsPage })),
);
const NotFoundPage = lazy(() =>
  import('@/pages/NotFoundPage').then((m) => ({ default: m.NotFoundPage })),
);

// Platform super-admin back-office (gated inside each page on user.super_admin).
const AdminDashboardPage = lazy(() =>
  import('@/pages/admin/AdminDashboardPage').then((m) => ({ default: m.AdminDashboardPage })),
);
const AdminOrganizationsPage = lazy(() =>
  import('@/pages/admin/AdminOrganizationsPage').then((m) => ({
    default: m.AdminOrganizationsPage,
  })),
);
const AdminUsersPage = lazy(() =>
  import('@/pages/admin/AdminUsersPage').then((m) => ({ default: m.AdminUsersPage })),
);

/** Centered spinner fallback used while a lazy route chunk loads. */
function RouteFallback() {
  return (
    <div className="flex min-h-[60vh] items-center justify-center">
      <Spinner />
    </div>
  );
}

/** Wraps a lazily-loaded element in a Suspense boundary. */
function withSuspense(node: ReactNode): ReactNode {
  return <Suspense fallback={<RouteFallback />}>{node}</Suspense>;
}

/**
 * Root layout: the SetupGate runs on EVERY route so the first-run wizard takes
 * priority over any deep-linked URL. The gate either renders <Outlet /> (the
 * matched route) or redirects to /setup or /login as required.
 */
function RootLayout() {
  return (
    <SetupGate>
      <Outlet />
    </SetupGate>
  );
}

export const router = createBrowserRouter([
  {
    element: <RootLayout />,
    children: [
      {
        // First-run bootstrap wizard. Public, but conditional: the SetupGate
        // sends users away when the platform is already initialized.
        path: '/setup',
        element: withSuspense(<SetupPage />),
      },
      {
        path: '/login',
        element: (
          <PublicOnlyRoute>{withSuspense(<LoginPage />)}</PublicOnlyRoute>
        ),
      },
      {
        path: '/register',
        element: (
          <PublicOnlyRoute>{withSuspense(<RegisterPage />)}</PublicOnlyRoute>
        ),
      },
      {
        path: '/forgot-password',
        element: (
          <PublicOnlyRoute>{withSuspense(<ForgotPasswordPage />)}</PublicOnlyRoute>
        ),
      },
      {
        path: '/reset-password',
        element: (
          <PublicOnlyRoute>{withSuspense(<ResetPasswordPage />)}</PublicOnlyRoute>
        ),
      },
      {
        // Email verification: accessible whether or not a session exists.
        path: '/verify-email',
        element: withSuspense(<VerifyEmailPage />),
      },
      {
        // Forced (or self-service) password change for the authenticated user.
        // Uses its own clean AuthLayout — no sidebar — so the forced-redirect
        // can't be visually circumvented.
        path: '/change-password',
        element: (
          <ProtectedRoute>{withSuspense(<ChangePasswordPage />)}</ProtectedRoute>
        ),
      },
      {
        path: '/termos',
        element: withSuspense(<TermsPage />),
      },
      {
        path: '/privacidade',
        element: withSuspense(<PrivacyPolicyPage />),
      },
      {
        path: '/',
        element: (
          <ProtectedRoute>
            <DashboardLayout />
          </ProtectedRoute>
        ),
        children: [
          { index: true, element: withSuspense(<DashboardPage />) },
          { path: 'transactions', element: withSuspense(<TransactionsPage />) },
          { path: 'recorrencias', element: withSuspense(<RecurrencesPage />) },
          { path: 'contas-a-pagar', element: withSuspense(<PayablesPage />) },
          { path: 'contas-a-receber', element: withSuspense(<ReceivablesPage />) },
          { path: 'contatos', element: withSuspense(<ContactsPage />) },
          { path: 'projecao', element: withSuspense(<ForecastPage />) },
          { path: 'relatorios', element: withSuspense(<ReportsPage />) },
          { path: 'accounts', element: withSuspense(<AccountsPage />) },
          { path: 'credit-cards', element: withSuspense(<CreditCardsPage />) },
          { path: 'categories', element: withSuspense(<CategoriesPage />) },
          { path: 'tags', element: withSuspense(<TagsPage />) },
          { path: 'categorizacao', element: withSuspense(<CategorizationPage />) },
          { path: 'budgets', element: withSuspense(<BudgetsPage />) },
          { path: 'goals', element: withSuspense(<GoalsPage />) },
          { path: 'team', element: withSuspense(<MembersPage />) },
          { path: 'auditoria', element: withSuspense(<AuditPage />) },
          { path: 'settings', element: withSuspense(<SettingsPage />) },
        ],
      },
      {
        // Platform super-admin back-office. Authenticated subtree with its own
        // (dark) layout; each page additionally gates on user.super_admin.
        path: '/admin',
        element: (
          <ProtectedRoute>
            <AdminLayout />
          </ProtectedRoute>
        ),
        children: [
          { index: true, element: withSuspense(<AdminDashboardPage />) },
          { path: 'organizacoes', element: withSuspense(<AdminOrganizationsPage />) },
          { path: 'usuarios', element: withSuspense(<AdminUsersPage />) },
        ],
      },
      { path: '/404', element: withSuspense(<NotFoundPage />) },
      { path: '*', element: <Navigate to="/404" replace /> },
    ],
  },
]);
