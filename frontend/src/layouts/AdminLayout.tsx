import { useState, type ComponentType } from 'react';
import { NavLink, Link, Outlet } from 'react-router-dom';
import {
  ShieldCheck,
  LayoutDashboard,
  Building2,
  Users,
  ArrowLeft,
  Menu,
  X,
} from 'lucide-react';
import { Avatar, ThemeToggle } from '@/components/ui';
import { useAuthStore } from '@/stores/authStore';
import { useLogout } from '@/hooks';
import { cn } from '@/lib/cn';

interface AdminNavItem {
  to: string;
  label: string;
  icon: ComponentType<{ className?: string }>;
  end?: boolean;
}

function AdminSidebar({ onNavigate }: { onNavigate?: () => void }) {
  const navItems: AdminNavItem[] = [
    { to: '/admin', label: 'Painel', icon: LayoutDashboard, end: true },
    { to: '/admin/organizacoes', label: 'Organizações', icon: Building2 },
    { to: '/admin/usuarios', label: 'Usuários', icon: Users },
  ];

  return (
    <div className="flex h-full flex-col text-gray-300">
      <div className="flex h-16 items-center gap-2.5 px-6">
        <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-white shadow-glow">
          <ShieldCheck className="h-5 w-5" />
        </span>
        <div className="leading-tight">
          <p className="font-heading text-base font-extrabold tracking-tight text-white">finance.sh</p>
          <p className="text-[10px] font-semibold uppercase tracking-wider text-primary-300">
            Administração
          </p>
        </div>
      </div>

      <nav className="flex-1 space-y-1 px-3 py-2">
        <p className="px-3 pb-2 pt-3 text-[11px] font-semibold uppercase tracking-wider text-gray-500">
          Plataforma
        </p>
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.end}
            onClick={onNavigate}
            className={({ isActive }) =>
              cn(
                'group flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition',
                isActive
                  ? 'bg-primary/15 text-primary-200'
                  : 'text-gray-400 hover:bg-white/5 hover:text-gray-100',
              )
            }
          >
            {({ isActive }) => (
              <>
                <item.icon
                  className={cn(
                    'h-[18px] w-[18px]',
                    isActive ? 'text-primary-300' : 'text-gray-500 group-hover:text-gray-300',
                  )}
                />
                <span className="flex-1">{item.label}</span>
              </>
            )}
          </NavLink>
        ))}
      </nav>

      <div className="px-3 pb-4">
        <Link
          to="/"
          onClick={onNavigate}
          className="flex items-center gap-3 rounded-lg border border-white/10 px-3 py-2.5 text-sm font-medium text-gray-300 transition hover:bg-white/5 hover:text-white"
        >
          <ArrowLeft className="h-[18px] w-[18px] text-gray-500" />
          Voltar ao app
        </Link>
      </div>
    </div>
  );
}

export function AdminLayout() {
  const [mobileOpen, setMobileOpen] = useState(false);
  const user = useAuthStore((s) => s.user);
  const logout = useLogout();

  return (
    // The sidebar + header are permanently dark (the visual differentiation from
    // the tenant app), while the content area follows the user's theme so shared
    // components (.card, tables, Modal) render correctly in light or dark mode.
    <div className="flex min-h-screen bg-gray-100 dark:bg-ink-base">
      {/* Desktop sidebar (always dark, distinct from tenant app) */}
      <aside className="hidden w-64 shrink-0 border-r border-white/10 bg-gray-900 lg:block">
        <div className="sticky top-0 h-screen">
          <AdminSidebar />
        </div>
      </aside>

      {/* Mobile drawer */}
      {mobileOpen && (
        <div className="fixed inset-0 z-50 lg:hidden">
          <div
            className="absolute inset-0 bg-black/60 backdrop-blur-sm"
            onClick={() => setMobileOpen(false)}
            aria-hidden
          />
          <aside className="relative h-full w-64 animate-slide-in-left border-r border-white/10 bg-gray-900">
            <button
              onClick={() => setMobileOpen(false)}
              className="absolute right-3 top-4 rounded-lg p-1.5 text-gray-400 hover:bg-white/10"
              aria-label="Fechar menu"
            >
              <X className="h-5 w-5" />
            </button>
            <AdminSidebar onNavigate={() => setMobileOpen(false)} />
          </aside>
        </div>
      )}

      {/* Main column */}
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-20 flex h-16 items-center gap-3 border-b border-white/10 bg-gray-900/80 px-4 backdrop-blur sm:px-6">
          <button
            onClick={() => setMobileOpen(true)}
            className="rounded-lg p-1.5 text-gray-400 hover:bg-white/10 lg:hidden"
            aria-label="Abrir menu"
          >
            <Menu className="h-5 w-5" />
          </button>

          <span className="inline-flex items-center gap-1.5 rounded-full bg-primary/15 px-2.5 py-1 text-xs font-semibold text-primary-200">
            <ShieldCheck className="h-3.5 w-3.5" />
            Modo administrador
          </span>

          <div className="ml-auto flex items-center gap-2">
            <ThemeToggle />
            <div className="mx-1 h-6 w-px bg-white/10" />
            {user && (
              <div className="flex items-center gap-2.5">
                <Avatar name={user.name} src={user.avatar_url} size="sm" />
                <div className="hidden min-w-0 leading-tight sm:block">
                  <p className="truncate text-sm font-medium text-gray-100">{user.name}</p>
                  <p className="truncate text-xs text-gray-500">{user.email}</p>
                </div>
              </div>
            )}
            <button
              onClick={() => logout.mutate()}
              className="ml-1 rounded-lg px-3 py-1.5 text-sm font-medium text-gray-300 transition hover:bg-white/5 hover:text-white"
            >
              Sair
            </button>
          </div>
        </header>

        <main className="flex-1 px-4 py-6 sm:px-6 lg:px-8">
          <div className="mx-auto max-w-7xl">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  );
}
