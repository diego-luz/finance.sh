import { useEffect, useState, type ComponentType, type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { NavLink, Link, Outlet } from 'react-router-dom';
import {
  LayoutDashboard,
  ArrowLeftRight,
  Repeat,
  ArrowUpCircle,
  ArrowDownCircle,
  LineChart,
  BarChart3,
  Contact as ContactIcon,
  Landmark,
  Tags,
  Tag as TagIcon,
  Wand2,
  CreditCard,
  Target,
  PiggyBank,
  Users,
  ScrollText,
  Settings,
  ShieldCheck,
  Menu,
  Search,
  X,
  ChevronDown,
  PanelLeftClose,
  PanelLeftOpen,
} from 'lucide-react';
import { Logo } from '@/components/Logo';
import { OrgSwitcher } from '@/components/OrgSwitcher';
import { UserMenu } from '@/components/UserMenu';
import { NotificationBell } from '@/components/NotificationBell';
import { EmailVerificationBanner } from '@/components/EmailVerificationBanner';
import { CommandPalette } from '@/components/CommandPalette';
import { ThemeToggle, PrivacyToggle } from '@/components/ui';
import { LanguageMenu } from '@/components/LanguageMenu';
import { useIsAdmin } from '@/hooks';
import { useAuthStore } from '@/stores/authStore';
import { usePrivacyStore } from '@/stores/privacyStore';
import { useSidebarStore } from '@/stores/sidebarStore';
import { cn } from '@/lib/cn';

// At most this many expandable sections stay open at once; opening one more
// auto-closes the oldest-opened (accordion with a cap).
const MAX_OPEN_SECTIONS = 2;

interface NavItem {
  to: string;
  /** Key under the `nav` namespace, resolved with `t()` at render time. */
  labelKey: string;
  icon: ComponentType<{ className?: string }>;
  end?: boolean;
}

interface NavSection {
  title: string;
  items: NavItem[];
}

// Sidebar menu grouped into sections. Section titles are pt-BR (mirrors the
// existing "Plataforma" header); move them to i18n if other locales need it.
const navSections: NavSection[] = [
  {
    title: 'Visão geral',
    items: [{ to: '/', labelKey: 'dashboard', icon: LayoutDashboard, end: true }],
  },
  {
    title: 'Finanças',
    items: [
      { to: '/transactions', labelKey: 'transactions', icon: ArrowLeftRight },
      { to: '/recorrencias', labelKey: 'recurrences', icon: Repeat },
      { to: '/contas-a-pagar', labelKey: 'accountsPayable', icon: ArrowUpCircle },
      { to: '/contas-a-receber', labelKey: 'accountsReceivable', icon: ArrowDownCircle },
    ],
  },
  {
    title: 'Cadastros',
    items: [
      { to: '/accounts', labelKey: 'accounts', icon: Landmark },
      { to: '/credit-cards', labelKey: 'creditCards', icon: CreditCard },
      { to: '/contatos', labelKey: 'contacts', icon: ContactIcon },
      { to: '/categories', labelKey: 'categories', icon: Tags },
      { to: '/tags', labelKey: 'tags', icon: TagIcon },
      { to: '/categorizacao', labelKey: 'categorization', icon: Wand2 },
    ],
  },
  {
    title: 'Planejamento',
    items: [
      { to: '/budgets', labelKey: 'budgets', icon: PiggyBank },
      { to: '/goals', labelKey: 'goals', icon: Target },
    ],
  },
  {
    title: 'Análise',
    items: [
      { to: '/projecao', labelKey: 'forecast', icon: LineChart },
      { to: '/relatorios', labelKey: 'reports', icon: BarChart3 },
    ],
  },
];

const adminSection: NavSection = {
  title: 'Equipe',
  items: [
    { to: '/team', labelKey: 'team', icon: Users },
    { to: '/auditoria', labelKey: 'audit', icon: ScrollText },
  ],
};

const accountSection: NavSection = {
  title: 'Conta',
  items: [{ to: '/settings', labelKey: 'settings', icon: Settings }],
};

const navLinkClass = ({ isActive }: { isActive: boolean }) =>
  cn(
    'group flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition',
    isActive
      ? 'bg-primary/10 text-primary-700 dark:text-primary-300'
      : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-ink-elevated dark:hover:text-gray-100',
  );

const navIconClass = (isActive: boolean) =>
  cn(
    'h-[18px] w-[18px]',
    isActive ? 'text-primary' : 'text-gray-400 group-hover:text-gray-600 dark:group-hover:text-gray-300',
  );

function SectionTitle({ children }: { children: ReactNode }) {
  return (
    <p className="px-3 pb-2 pt-4 text-[11px] font-semibold uppercase tracking-wider text-gray-400">
      {children}
    </p>
  );
}

function SidebarContent({
  onNavigate,
  collapsed = false,
  onToggleCollapse,
}: {
  onNavigate?: () => void;
  /** Icons-only rail (desktop). */
  collapsed?: boolean;
  /** When provided, renders the collapse/expand toggle (desktop only). */
  onToggleCollapse?: () => void;
}) {
  const { t } = useTranslation();
  const isAdmin = useIsAdmin();
  const isSuperAdmin = useAuthStore((s) => Boolean(s.user?.super_admin));

  const sections: NavSection[] = [
    ...navSections,
    ...(isAdmin ? [adminSection] : []),
    accountSection,
  ];

  // Accordion (expanded mode): at most MAX_OPEN_SECTIONS open; opening another
  // closes the oldest-opened. Default-open the first two sections.
  const [openSections, setOpenSections] = useState<string[]>(() =>
    navSections.slice(0, MAX_OPEN_SECTIONS).map((s) => s.title),
  );
  function toggleSection(title: string) {
    setOpenSections((prev) => {
      if (prev.includes(title)) return prev.filter((x) => x !== title);
      const next = [...prev, title];
      return next.length > MAX_OPEN_SECTIONS ? next.slice(next.length - MAX_OPEN_SECTIONS) : next;
    });
  }

  const renderItem = (item: NavItem, label?: string) => {
    const text = label ?? t(`nav.${item.labelKey}`);
    return (
      <NavLink
        key={item.to}
        to={item.to}
        end={item.end}
        onClick={onNavigate}
        title={collapsed ? text : undefined}
        className={({ isActive }) => cn(navLinkClass({ isActive }), collapsed && 'justify-center px-0')}
      >
        {({ isActive }) => (
          <>
            <item.icon className={navIconClass(isActive)} />
            {!collapsed && text}
          </>
        )}
      </NavLink>
    );
  };

  return (
    <div className="flex h-full flex-col">
      <div className={cn('flex h-16 items-center px-6', collapsed && 'justify-center px-0')}>
        <Logo showWord={!collapsed} />
      </div>

      <nav className="flex-1 space-y-1 overflow-y-auto px-3 py-2">
        {sections.map((section) => {
          const open = collapsed || openSections.includes(section.title);
          return (
            <div key={section.title}>
              {collapsed ? (
                <div className="my-1 border-t border-gray-100 dark:border-ink-border" />
              ) : (
                <button
                  type="button"
                  onClick={() => toggleSection(section.title)}
                  className="flex w-full items-center justify-between px-3 pb-1 pt-4 text-[11px] font-semibold uppercase tracking-wider text-gray-400 transition hover:text-gray-600 dark:hover:text-gray-300"
                >
                  {section.title}
                  <ChevronDown className={cn('h-3.5 w-3.5 transition-transform', !open && '-rotate-90')} />
                </button>
              )}
              {open && section.items.map((item) => renderItem(item))}
            </div>
          );
        })}

        {isSuperAdmin && (
          <div>
            {collapsed ? (
              <div className="my-1 border-t border-gray-100 dark:border-ink-border" />
            ) : (
              <SectionTitle>Plataforma</SectionTitle>
            )}
            {renderItem({ to: '/admin', labelKey: 'admin', icon: ShieldCheck }, 'Administração')}
          </div>
        )}
      </nav>

      {onToggleCollapse && (
        <div className="border-t border-gray-200 px-3 py-2 dark:border-ink-border">
          <button
            type="button"
            onClick={onToggleCollapse}
            title={collapsed ? 'Expandir menu' : 'Recolher menu'}
            aria-label={collapsed ? 'Expandir menu' : 'Recolher menu'}
            className={cn(
              'flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-gray-500 transition hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-ink-elevated dark:hover:text-gray-100',
              collapsed && 'justify-center px-0',
            )}
          >
            {collapsed ? (
              <PanelLeftOpen className="h-[18px] w-[18px]" />
            ) : (
              <PanelLeftClose className="h-[18px] w-[18px]" />
            )}
            {!collapsed && 'Recolher menu'}
          </button>
        </div>
      )}
    </div>
  );
}

export function DashboardLayout() {
  const { t } = useTranslation();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [paletteOpen, setPaletteOpen] = useState(false);
  const privacyHidden = usePrivacyStore((s) => s.hidden);
  const collapsed = useSidebarStore((s) => s.collapsed);
  const toggleSidebar = useSidebarStore((s) => s.toggle);

  // Keep the global `values-hidden` class on <html> in sync with the privacy
  // store. Toggling the class blurs every `.money` / `.money-sensitive` element
  // instantly (CSS only), so no React re-render of money displays is needed and
  // charts are covered too.
  useEffect(() => {
    document.documentElement.classList.toggle('values-hidden', privacyHidden);
  }, [privacyHidden]);

  // Global ⌘K / Ctrl+K opens the command palette (Esc is handled inside it).
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && (e.key === 'k' || e.key === 'K')) {
        e.preventDefault();
        setPaletteOpen(true);
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, []);

  return (
    <div className="flex min-h-screen bg-gray-100 dark:bg-ink-base">
      {/* Desktop sidebar (collapsible to an icons-only rail) */}
      <aside
        className={cn(
          'hidden shrink-0 border-r border-gray-200 bg-white transition-[width] duration-200 dark:border-ink-border dark:bg-ink-surface lg:block',
          collapsed ? 'w-[76px]' : 'w-64',
        )}
      >
        <div className="sticky top-0 h-screen">
          <SidebarContent collapsed={collapsed} onToggleCollapse={toggleSidebar} />
        </div>
      </aside>

      {/* Mobile drawer */}
      {mobileOpen && (
        <div className="fixed inset-0 z-50 lg:hidden">
          <div
            className="absolute inset-0 bg-black/50 backdrop-blur-sm"
            onClick={() => setMobileOpen(false)}
            aria-hidden
          />
          <aside className="relative h-full w-64 animate-slide-in-left border-r border-gray-200 bg-white dark:border-ink-border dark:bg-ink-surface">
            <button
              onClick={() => setMobileOpen(false)}
              className="absolute right-3 top-4 rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 dark:hover:bg-ink-elevated"
              aria-label={t('common.closeMenu')}
            >
              <X className="h-5 w-5" />
            </button>
            <SidebarContent onNavigate={() => setMobileOpen(false)} />
          </aside>
        </div>
      )}

      {/* Main column */}
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-20 flex h-16 items-center gap-3 border-b border-gray-200 bg-white/80 px-4 backdrop-blur dark:border-ink-border dark:bg-ink-base/80 sm:px-6">
          <button
            onClick={() => setMobileOpen(true)}
            className="rounded-lg p-1.5 text-gray-500 hover:bg-gray-100 dark:hover:bg-ink-elevated lg:hidden"
            aria-label={t('common.openMenu')}
          >
            <Menu className="h-5 w-5" />
          </button>

          <OrgSwitcher />

          {/* Global search trigger (⌘K) */}
          <button
            onClick={() => setPaletteOpen(true)}
            className="ml-2 hidden h-9 items-center gap-2 rounded-lg border border-gray-200 bg-gray-50 px-3 text-sm text-gray-400 transition hover:border-gray-300 hover:text-gray-600 dark:border-ink-border dark:bg-ink-elevated dark:hover:border-gray-600 dark:hover:text-gray-300 sm:flex"
            aria-label={t('common.search')}
          >
            <Search className="h-4 w-4" />
            <span>{t('common.searchPlaceholder')}</span>
            <kbd className="ml-2 rounded border border-gray-200 bg-white px-1.5 py-0.5 text-[10px] font-medium text-gray-400 dark:border-ink-border dark:bg-ink-surface">
              ⌘K
            </kbd>
          </button>

          <div className="ml-auto flex items-center gap-1.5">
            <button
              onClick={() => setPaletteOpen(true)}
              className="rounded-lg p-1.5 text-gray-500 hover:bg-gray-100 dark:hover:bg-ink-elevated sm:hidden"
              aria-label={t('common.search')}
            >
              <Search className="h-5 w-5" />
            </button>
            <NotificationBell />
            <LanguageMenu />
            <PrivacyToggle />
            <ThemeToggle />
            <div className="mx-1 h-6 w-px bg-gray-200 dark:bg-ink-border" />
            <UserMenu />
          </div>
        </header>

        <main className="flex-1 px-4 py-6 sm:px-6 lg:px-8">
          <div className="mx-auto max-w-7xl">
            <EmailVerificationBanner />
            <Outlet />
          </div>
        </main>

        <footer className="border-t border-gray-200 px-4 py-4 dark:border-ink-border sm:px-6 lg:px-8">
          <div className="mx-auto flex max-w-7xl flex-col items-center justify-between gap-2 text-xs text-gray-400 sm:flex-row">
            <span>{t('footer.rights', { year: new Date().getFullYear() })}</span>
            <div className="flex items-center gap-4">
              <Link to="/termos" className="hover:text-primary">
                {t('footer.terms')}
              </Link>
              <Link to="/privacidade" className="hover:text-primary">
                {t('footer.privacy')}
              </Link>
            </div>
          </div>
        </footer>
      </div>

      <CommandPalette open={paletteOpen} onClose={() => setPaletteOpen(false)} />
    </div>
  );
}
