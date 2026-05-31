import { useState } from 'react';
import {
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
} from 'recharts';
import { TrendingUp, TrendingDown, Scale, PieChart as PieIcon, BarChart3 } from 'lucide-react';
import { PageHeader } from '@/components/PageHeader';
import { Card, CardHeader, StatCard, EmptyState, Skeleton } from '@/components/ui';
import { PeriodFilter, type PeriodValue } from '@/components/ui';
import { ExportMenu } from '@/components/ExportMenu';
import { Money } from '@/components/Money';
import { useReportSummary } from '@/hooks';
import { useThemeStore } from '@/stores/themeStore';
import { formatCurrency } from '@/lib/currency';
import { resolvePeriodPreset } from '@/lib/date';
import type { ReportCategoryTotal, ReportMonthTotal } from '@/types';

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
      {label && <p className="mb-1 font-medium text-gray-900 dark:text-gray-100">{label}</p>}
      {payload.map((p) => (
        <p key={p.name} className="flex items-center gap-1.5 text-gray-600 dark:text-gray-300">
          <span className="h-2 w-2 rounded-full" style={{ backgroundColor: p.color }} />
          {p.name}: <span className="font-medium tabular-nums">{formatCurrency(p.value)}</span>
        </p>
      ))}
    </div>
  );
}

function CategoryDonut({ data }: { data: ReportCategoryTotal[] }) {
  if (!data.length) {
    return (
      <EmptyState
        icon={PieIcon}
        title="Sem dados de categorias"
        description="Não há movimentações categorizadas no período selecionado."
      />
    );
  }
  return (
    <div className="money-sensitive">
      <ResponsiveContainer width="100%" height={300}>
      <PieChart>
        <Pie
          data={data}
          dataKey="total"
          nameKey="name"
          innerRadius={64}
          outerRadius={100}
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

function MonthBars({ data, isDark }: { data: ReportMonthTotal[]; isDark: boolean }) {
  if (!data.length) {
    return (
      <EmptyState
        icon={BarChart3}
        title="Sem dados mensais"
        description="Não há movimentações no período selecionado."
      />
    );
  }
  const grid = isDark ? '#262932' : '#e5e7eb';
  const axis = isDark ? '#6b7280' : '#9ca3af';
  return (
    <div className="money-sensitive">
      <ResponsiveContainer width="100%" height={300}>
      <BarChart data={data} margin={{ top: 10, right: 6, left: -10, bottom: 0 }}>
        <CartesianGrid strokeDasharray="3 3" stroke={grid} vertical={false} />
        <XAxis dataKey="month" tick={{ fontSize: 12, fill: axis }} axisLine={false} tickLine={false} />
        <YAxis
          tick={{ fontSize: 12, fill: axis }}
          axisLine={false}
          tickLine={false}
          tickFormatter={(v: number) => formatCurrency(v)}
          width={80}
        />
        <Tooltip cursor={{ fill: isDark ? '#ffffff0a' : '#0000000a' }} content={<ChartTooltip />} />
        <Legend
          iconType="circle"
          formatter={(value) => (
            <span className="text-xs text-gray-600 dark:text-gray-300">{value}</span>
          )}
        />
        <Bar dataKey="income" name="Receitas" fill="#10b981" radius={[4, 4, 0, 0]} />
        <Bar dataKey="expense" name="Despesas" fill="#f43f5e" radius={[4, 4, 0, 0]} />
      </BarChart>
    </ResponsiveContainer>
    </div>
  );
}

function CategoryBreakdown({ data }: { data: ReportCategoryTotal[] }) {
  const max = Math.max(1, ...data.map((c) => c.total));
  return (
    <ul className="space-y-3">
      {data.map((c) => (
        <li key={c.category_id}>
          <div className="mb-1 flex items-center justify-between gap-3 text-sm">
            <span className="flex min-w-0 items-center gap-2">
              <span className="h-2.5 w-2.5 shrink-0 rounded-full" style={{ backgroundColor: c.color || '#10b981' }} />
              <span className="truncate text-gray-700 dark:text-gray-300">{c.name}</span>
            </span>
            <Money
              value={c.total}
              className="shrink-0 font-semibold text-gray-900 dark:text-gray-100"
            />
          </div>
          <div className="h-2 w-full overflow-hidden rounded-full bg-gray-100 dark:bg-ink-elevated">
            <div
              className="h-full rounded-full transition-all"
              style={{ width: `${Math.round((c.total / max) * 100)}%`, backgroundColor: c.color || '#10b981' }}
            />
          </div>
        </li>
      ))}
    </ul>
  );
}

export function ReportsPage() {
  // Default to "Este mês" so the page loads with meaningful data.
  const [range, setRange] = useState<PeriodValue>(() => {
    const { from, to } = resolvePeriodPreset('this_month');
    return { from, to };
  });

  const { data, isLoading, isError, isFetching } = useReportSummary(range);
  const isDark = useThemeStore((s) => s.theme === 'dark');

  const byCategory = data?.by_category ?? [];
  const byMonth = data?.by_month ?? [];
  const hasData = byCategory.length > 0 || byMonth.length > 0;

  return (
    <div>
      <PageHeader
        eyebrow="Análises"
        title="Relatórios"
        description="Visão consolidada de receitas, despesas e resultado por período."
        action={<ExportMenu range={range} />}
      />

      <Card className="mb-4 p-4">
        <PeriodFilter value={range} onChange={setRange} />
      </Card>

      {isError ? (
        <Card className="border-red-200 bg-red-50 dark:border-red-500/30 dark:bg-red-500/10">
          <p className="text-sm text-red-600 dark:text-red-400">
            Não foi possível carregar os relatórios. Verifique sua conexão e tente novamente.
          </p>
        </Card>
      ) : (
        <>
          {/* KPI cards */}
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            {isLoading || !data ? (
              Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-28 rounded-xl" />)
            ) : (
              <>
                <StatCard
                  label="Receitas"
                  value={<Money value={data.income} />}
                  icon={TrendingUp}
                  tone="primary"
                  trend="up"
                />
                <StatCard
                  label="Despesas"
                  value={<Money value={data.expense} />}
                  icon={TrendingDown}
                  tone="red"
                  trend="down"
                />
                <StatCard
                  label="Resultado"
                  value={<Money value={data.net} />}
                  icon={Scale}
                  tone={data.net >= 0 ? 'primary' : 'red'}
                  caption={data.net >= 0 ? 'Superávit' : 'Déficit'}
                />
              </>
            )}
          </div>

          {/* Charts */}
          <div className="mt-4 grid grid-cols-1 gap-4 lg:grid-cols-2">
            <Card>
              <CardHeader
                eyebrow="Distribuição"
                title="Por categoria"
                subtitle="Totais agrupados no período"
              />
              {isLoading || !data ? (
                <div className="skeleton h-[300px] w-full" />
              ) : (
                <CategoryDonut data={byCategory} />
              )}
            </Card>

            <Card>
              <CardHeader
                eyebrow="Evolução"
                title="Receitas x Despesas por mês"
                subtitle="Comparativo mensal no período"
              />
              {isLoading || !data ? (
                <div className="skeleton h-[300px] w-full" />
              ) : (
                <MonthBars data={byMonth} isDark={isDark} />
              )}
            </Card>
          </div>

          {/* Category breakdown list */}
          <Card className="mt-4">
            <CardHeader
              eyebrow="Detalhamento"
              title="Categorias"
              subtitle="Valores por categoria, do maior para o menor"
            />
            {isLoading || !data ? (
              <div className="space-y-3">
                {Array.from({ length: 5 }).map((_, i) => (
                  <Skeleton key={i} className="h-10 rounded-lg" />
                ))}
              </div>
            ) : byCategory.length === 0 ? (
              <EmptyState
                icon={PieIcon}
                title="Sem dados no período"
                description="Ajuste o período para visualizar o detalhamento por categoria."
              />
            ) : (
              <CategoryBreakdown data={[...byCategory].sort((a, b) => b.total - a.total)} />
            )}
          </Card>

          {!isLoading && data && !hasData && (
            <p className="mt-4 text-center text-xs text-gray-400">
              Nenhuma movimentação no período selecionado.
              {isFetching && <span className="ml-2 text-primary">atualizando…</span>}
            </p>
          )}
        </>
      )}
    </div>
  );
}
