import {
  ResponsiveContainer,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
  PieChart,
  Pie,
  Cell,
  Legend,
} from 'recharts';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import {
  Wallet,
  TrendingUp,
  TrendingDown,
  Scale,
  CalendarClock,
  PieChart as PieIcon,
  Target,
  CreditCard as CreditCardIcon,
  AlertTriangle,
  ArrowRight,
} from 'lucide-react';
import { PageHeader } from '@/components/PageHeader';
import { Card, CardHeader, StatCard, Badge, SkeletonCard, EmptyState } from '@/components/ui';
import { Money } from '@/components/Money';
import { useDashboard, useGoals, useCreditCards, useForecast } from '@/hooks';
import { formatCurrency } from '@/lib/currency';
import { formatDateShort } from '@/lib/date';
import { resolveIcon, accountTypeMeta } from '@/lib/icons';
import { cn } from '@/lib/cn';
import { useThemeStore } from '@/stores/themeStore';
import type { CashFlowPoint, TopCategory } from '@/types';

function ChartTooltip({
  active,
  payload,
  label,
}: {
  active?: boolean;
  payload?: Array<{ name: string; value: number; color: string }>;
  label?: string;
}) {
  if (!active || !payload?.length) return null;
  return (
    <div className="rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs shadow-card dark:border-ink-border dark:bg-ink-elevated">
      <p className="mb-1 font-medium text-gray-900 dark:text-gray-100">{label}</p>
      {payload.map((p) => (
        <p key={p.name} className="flex items-center gap-1.5 text-gray-600 dark:text-gray-300">
          <span className="h-2 w-2 rounded-full" style={{ backgroundColor: p.color }} />
          {p.name}: <span className="font-medium tabular-nums">{formatCurrency(p.value)}</span>
        </p>
      ))}
    </div>
  );
}

function CashFlowChart({ data }: { data: CashFlowPoint[] }) {
  const { t } = useTranslation();
  const isDark = useThemeStore((s) => s.theme === 'dark');
  const grid = isDark ? '#262932' : '#e5e7eb';
  const axis = isDark ? '#6b7280' : '#9ca3af';

  return (
    <div className="money-sensitive">
      <ResponsiveContainer width="100%" height={280}>
      <AreaChart data={data} margin={{ top: 10, right: 6, left: -10, bottom: 0 }}>
        <defs>
          <linearGradient id="incomeGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#10b981" stopOpacity={0.35} />
            <stop offset="100%" stopColor="#10b981" stopOpacity={0} />
          </linearGradient>
          <linearGradient id="expenseGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#f43f5e" stopOpacity={0.3} />
            <stop offset="100%" stopColor="#f43f5e" stopOpacity={0} />
          </linearGradient>
        </defs>
        <CartesianGrid strokeDasharray="3 3" stroke={grid} vertical={false} />
        <XAxis dataKey="month" tick={{ fontSize: 12, fill: axis }} axisLine={false} tickLine={false} />
        <YAxis
          tick={{ fontSize: 12, fill: axis }}
          axisLine={false}
          tickLine={false}
          tickFormatter={(v: number) => formatCurrency(v).replace(/\s/g, ' ')}
          width={80}
        />
        <Tooltip content={<ChartTooltip />} />
        <Area
          type="monotone"
          dataKey="income"
          name={t('dashboard.cashFlow.income')}
          stroke="#10b981"
          strokeWidth={2}
          fill="url(#incomeGrad)"
        />
        <Area
          type="monotone"
          dataKey="expense"
          name={t('dashboard.cashFlow.expense')}
          stroke="#f43f5e"
          strokeWidth={2}
          fill="url(#expenseGrad)"
        />
      </AreaChart>
    </ResponsiveContainer>
    </div>
  );
}

function TopCategoriesChart({ data }: { data: TopCategory[] }) {
  const { t } = useTranslation();
  if (!data.length) {
    return (
      <EmptyState
        icon={PieIcon}
        title={t('dashboard.topCategories.emptyTitle')}
        description={t('dashboard.topCategories.emptyDescription')}
      />
    );
  }
  return (
    <div className="money-sensitive">
      <ResponsiveContainer width="100%" height={280}>
      <PieChart>
        <Pie
          data={data}
          dataKey="total"
          nameKey="name"
          innerRadius={60}
          outerRadius={95}
          paddingAngle={2}
          stroke="none"
        >
          {data.map((c) => (
            <Cell key={c.category_id} fill={c.color || '#10b981'} />
          ))}
        </Pie>
        <Tooltip content={<ChartTooltip />} />
        <Legend
          iconType="circle"
          formatter={(value) => (
            <span className="text-xs text-gray-600 dark:text-gray-300">{value}</span>
          )}
        />
      </PieChart>
    </ResponsiveContainer>
    </div>
  );
}

function GoalsWidget() {
  const { t } = useTranslation();
  const { data: goals, isLoading } = useGoals();
  const top = (goals ?? []).slice(0, 3);

  return (
    <Card>
      <CardHeader
        title={t('dashboard.goals.title')}
        subtitle={t('dashboard.goals.subtitle')}
        action={
          <Link
            to="/goals"
            className="inline-flex items-center gap-1 text-xs font-medium text-primary hover:text-primary-600"
          >
            {t('dashboard.goals.seeAll')} <ArrowRight className="h-3.5 w-3.5" />
          </Link>
        }
      />
      {isLoading ? (
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="skeleton h-12 w-full rounded-lg" />
          ))}
        </div>
      ) : top.length === 0 ? (
        <EmptyState
          icon={Target}
          title={t('dashboard.goals.emptyTitle')}
          description={t('dashboard.goals.emptyDescription')}
        />
      ) : (
        <ul className="space-y-4">
          {top.map((goal) => {
            const pct = Math.round(Math.max(0, Math.min(1, goal.progress)) * 100);
            const color = goal.color || '#0ea5e9';
            return (
              <li key={goal.id}>
                <div className="mb-1.5 flex items-center justify-between gap-3 text-sm">
                  <span className="truncate font-medium text-gray-900 dark:text-gray-100">
                    {goal.name}
                  </span>
                  <span className="shrink-0 tabular-nums text-gray-500 dark:text-gray-400">
                    <Money value={goal.current_amount} />
                    <span className="text-gray-400">
                      {' '}/ <Money value={goal.target_amount} />
                    </span>
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="h-2 flex-1 overflow-hidden rounded-full bg-gray-100 dark:bg-ink-elevated">
                    <div
                      className="h-full rounded-full transition-all"
                      style={{ width: `${pct}%`, backgroundColor: color }}
                    />
                  </div>
                  <span className="w-9 shrink-0 text-right text-xs font-medium tabular-nums text-gray-500 dark:text-gray-400">
                    {pct}%
                  </span>
                </div>
              </li>
            );
          })}
        </ul>
      )}
    </Card>
  );
}

function CreditCardsWidget() {
  const { t } = useTranslation();
  const { data: cards, isLoading } = useCreditCards();
  const top = (cards ?? []).slice(0, 3);

  return (
    <Card>
      <CardHeader
        title={t('dashboard.creditCards.title')}
        subtitle={t('dashboard.creditCards.subtitle')}
        action={
          <Link
            to="/credit-cards"
            className="inline-flex items-center gap-1 text-xs font-medium text-primary hover:text-primary-600"
          >
            {t('dashboard.creditCards.seeAll')} <ArrowRight className="h-3.5 w-3.5" />
          </Link>
        }
      />
      {isLoading ? (
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="skeleton h-12 w-full rounded-lg" />
          ))}
        </div>
      ) : top.length === 0 ? (
        <EmptyState
          icon={CreditCardIcon}
          title={t('dashboard.creditCards.emptyTitle')}
          description={t('dashboard.creditCards.emptyDescription')}
        />
      ) : (
        <ul className="space-y-4">
          {top.map((card) => {
            const pct = card.limit > 0 ? Math.min(100, Math.round((card.used / card.limit) * 100)) : 0;
            const over = card.used > card.limit;
            const color = card.color || '#6366f1';
            return (
              <li key={card.id}>
                <div className="mb-1.5 flex items-center justify-between gap-3 text-sm">
                  <span className="truncate font-medium text-gray-900 dark:text-gray-100">
                    {card.name}
                  </span>
                  <span className="shrink-0 tabular-nums text-gray-500 dark:text-gray-400">
                    <Money value={card.available} /> {t('dashboard.creditCards.available')}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="h-2 flex-1 overflow-hidden rounded-full bg-gray-100 dark:bg-ink-elevated">
                    <div
                      className="h-full rounded-full transition-all"
                      style={{ width: `${pct}%`, backgroundColor: over ? '#ef4444' : color }}
                    />
                  </div>
                  <span
                    className={cn(
                      'w-9 shrink-0 text-right text-xs font-medium tabular-nums',
                      over ? 'text-red-500' : 'text-gray-500 dark:text-gray-400',
                    )}
                  >
                    {pct}%
                  </span>
                </div>
              </li>
            );
          })}
        </ul>
      )}
    </Card>
  );
}

function ForecastWidget() {
  const { t } = useTranslation();
  const { data, isLoading } = useForecast(3);
  const hasNegative = (data?.months ?? []).some((m) => m.projected_balance < 0);

  return (
    <Card>
      <CardHeader
        title={t('dashboard.forecast.title')}
        subtitle={t('dashboard.forecast.subtitle')}
        action={
          <Link
            to="/projecao"
            className="inline-flex items-center gap-1 text-xs font-medium text-primary hover:text-primary-600"
          >
            {t('dashboard.forecast.seeForecast')} <ArrowRight className="h-3.5 w-3.5" />
          </Link>
        }
      />
      {isLoading || !data ? (
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="skeleton h-10 w-full rounded-lg" />
          ))}
        </div>
      ) : (
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <div className="rounded-xl border border-gray-100 p-3 dark:border-ink-border">
              <p className="text-xs text-gray-500 dark:text-gray-400">{t('dashboard.forecast.currentBalance')}</p>
              <p className="mt-0.5 font-heading text-lg font-bold tabular-nums text-gray-900 dark:text-gray-50">
                <Money value={data.current_balance} />
              </p>
            </div>
            <div className="rounded-xl border border-gray-100 p-3 dark:border-ink-border">
              <p className="text-xs text-gray-500 dark:text-gray-400">{t('dashboard.forecast.projectedBalance')}</p>
              <p
                className={cn(
                  'mt-0.5 font-heading text-lg font-bold tabular-nums',
                  data.end_balance >= 0 ? 'text-gray-900 dark:text-gray-50' : 'text-red-500',
                )}
              >
                <Money value={data.end_balance} />
              </p>
            </div>
          </div>
          {hasNegative ? (
            <div className="flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-xs text-red-600 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-400">
              <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
              <span>{t('dashboard.forecast.negativeWarning')}</span>
            </div>
          ) : (
            <p className="text-xs text-gray-400">{t('dashboard.forecast.positiveNote')}</p>
          )}
        </div>
      )}
    </Card>
  );
}

export function DashboardPage() {
  const { t } = useTranslation();
  const { data, isLoading, isError } = useDashboard();

  return (
    <div>
      <PageHeader
        eyebrow={t('dashboard.eyebrow')}
        title={t('dashboard.title')}
        description={t('dashboard.description')}
      />

      {/* KPI cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {isLoading || !data ? (
          Array.from({ length: 4 }).map((_, i) => <SkeletonCard key={i} />)
        ) : (
          <>
            <StatCard
              label={t('dashboard.kpi.currentBalance')}
              value={<Money value={data.balance} />}
              icon={Wallet}
              tone="primary"
              caption={t('dashboard.kpi.balanceSum')}
            />
            <StatCard
              label={t('dashboard.kpi.monthIncome')}
              value={<Money value={data.month_income} />}
              icon={TrendingUp}
              tone="sky"
              trend="up"
            />
            <StatCard
              label={t('dashboard.kpi.monthExpense')}
              value={<Money value={data.month_expense} />}
              icon={TrendingDown}
              tone="red"
              trend="down"
            />
            <StatCard
              label={t('dashboard.kpi.monthResult')}
              value={<Money value={data.month_net} />}
              icon={Scale}
              tone={data.month_net >= 0 ? 'primary' : 'red'}
              caption={data.month_net >= 0 ? t('dashboard.kpi.surplus') : t('dashboard.kpi.deficit')}
            />
          </>
        )}
      </div>

      {isError && (
        <Card className="mt-6 border-red-200 bg-red-50 dark:border-red-500/30 dark:bg-red-500/10">
          <p className="text-sm text-red-600 dark:text-red-400">
            {t('dashboard.loadError')}
          </p>
        </Card>
      )}

      {/* Charts */}
      <div className="mt-6 grid grid-cols-1 gap-4 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader
            eyebrow={t('dashboard.cashFlow.eyebrow')}
            title={t('dashboard.cashFlow.title')}
            subtitle={t('dashboard.cashFlow.subtitle')}
          />
          {isLoading || !data ? (
            <div className="skeleton h-[280px] w-full" />
          ) : (
            <CashFlowChart data={data.cash_flow} />
          )}
        </Card>

        <Card>
          <CardHeader
            eyebrow={t('dashboard.topCategories.eyebrow')}
            title={t('dashboard.topCategories.title')}
            subtitle={t('dashboard.topCategories.subtitle')}
          />
          {isLoading || !data ? (
            <div className="skeleton h-[280px] w-full" />
          ) : (
            <TopCategoriesChart data={data.top_categories} />
          )}
        </Card>
      </div>

      {/* Lower section: upcoming bills + accounts summary */}
      <div className="mt-6 grid grid-cols-1 gap-4 lg:grid-cols-3">
        <Card className="lg:col-span-1">
          <CardHeader
            title={t('dashboard.upcomingBills.title')}
            subtitle={t('dashboard.upcomingBills.subtitle')}
            action={<CalendarClock className="h-5 w-5 text-gray-400" />}
          />
          {isLoading || !data ? (
            <div className="space-y-3">
              {Array.from({ length: 4 }).map((_, i) => (
                <div key={i} className="skeleton h-12 w-full rounded-lg" />
              ))}
            </div>
          ) : data.upcoming_bills.length === 0 ? (
            <EmptyState
              icon={CalendarClock}
              title={t('dashboard.upcomingBills.emptyTitle')}
              description={t('dashboard.upcomingBills.emptyDescription')}
            />
          ) : (
            <ul className="space-y-1">
              {data.upcoming_bills.map((bill) => (
                <li
                  key={bill.id}
                  className="flex items-center justify-between gap-3 rounded-lg px-2 py-2.5 transition hover:bg-gray-50 dark:hover:bg-ink-elevated"
                >
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium text-gray-900 dark:text-gray-100">
                      {bill.description}
                    </p>
                    <p className="text-xs text-gray-400">{t('dashboard.upcomingBills.due', { date: formatDateShort(bill.date) })}</p>
                  </div>
                  <Money
                    value={bill.amount}
                    className="shrink-0 text-sm font-semibold text-red-500"
                  />
                </li>
              ))}
            </ul>
          )}
        </Card>

        <Card className="lg:col-span-2">
          <CardHeader
            title={t('dashboard.accountsSummary.title')}
            subtitle={t('dashboard.accountsSummary.subtitle')}
          />
          {isLoading || !data ? (
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              {Array.from({ length: 4 }).map((_, i) => (
                <div key={i} className="skeleton h-20 w-full rounded-xl" />
              ))}
            </div>
          ) : data.accounts_summary.length === 0 ? (
            <EmptyState
              icon={Wallet}
              title={t('dashboard.accountsSummary.emptyTitle')}
              description={t('dashboard.accountsSummary.emptyDescription')}
            />
          ) : (
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              {data.accounts_summary.map((acc) => {
                const Icon = resolveIcon(acc.icon);
                return (
                  <div
                    key={acc.id}
                    className="flex items-center gap-3 rounded-xl border border-gray-100 p-3 dark:border-ink-border"
                  >
                    <span
                      className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg"
                      style={{ backgroundColor: `${acc.color}1a`, color: acc.color }}
                    >
                      <Icon className="h-5 w-5" />
                    </span>
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium text-gray-900 dark:text-gray-100">
                        {acc.name}
                      </p>
                      <Badge variant="neutral">{accountTypeMeta[acc.type]?.label ?? acc.type}</Badge>
                    </div>
                    <Money
                      value={acc.balance}
                      className="shrink-0 text-sm font-semibold text-gray-900 dark:text-gray-100"
                    />
                  </div>
                );
              })}
            </div>
          )}
        </Card>
      </div>

      {/* Forecast widget */}
      <div className="mt-6">
        <ForecastWidget />
      </div>

      {/* Goals + credit cards widgets */}
      <div className="mt-6 grid grid-cols-1 gap-4 lg:grid-cols-2">
        <GoalsWidget />
        <CreditCardsWidget />
      </div>
    </div>
  );
}
