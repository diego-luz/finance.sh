import { useMemo, useState } from 'react';
import {
  ResponsiveContainer,
  ComposedChart,
  Area,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
  Legend,
} from 'recharts';
import { Wallet, TrendingUp, TrendingDown, AlertTriangle, LineChart } from 'lucide-react';
import { PageHeader } from '@/components/PageHeader';
import { Card, CardHeader, StatCard, Select, EmptyState } from '@/components/ui';
import { Money } from '@/components/Money';
import { useForecast } from '@/hooks';
import { formatCurrency } from '@/lib/currency';
import { useThemeStore } from '@/stores/themeStore';
import type { ForecastMonth } from '@/types';

const monthsOptions = [
  { value: '3', label: '3 meses' },
  { value: '6', label: '6 meses' },
  { value: '12', label: '12 meses' },
];

/** "2026-05" -> "mai/26" */
function monthLabel(month: string): string {
  const m = /^(\d{4})-(\d{2})/.exec(month);
  if (!m) return month;
  const names = [
    'jan',
    'fev',
    'mar',
    'abr',
    'mai',
    'jun',
    'jul',
    'ago',
    'set',
    'out',
    'nov',
    'dez',
  ];
  const idx = Number(m[2]) - 1;
  return `${names[idx] ?? m[2]}/${m[1].slice(2)}`;
}

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

function ForecastChart({ months }: { months: ForecastMonth[] }) {
  const isDark = useThemeStore((s) => s.theme === 'dark');
  const grid = isDark ? '#262932' : '#e5e7eb';
  const axis = isDark ? '#6b7280' : '#9ca3af';

  const data = months.map((m) => ({ ...m, label: monthLabel(m.month) }));

  return (
    <div className="money-sensitive">
      <ResponsiveContainer width="100%" height={320}>
      <ComposedChart data={data} margin={{ top: 10, right: 6, left: -10, bottom: 0 }}>
        <defs>
          <linearGradient id="balanceGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#10b981" stopOpacity={0.35} />
            <stop offset="100%" stopColor="#10b981" stopOpacity={0} />
          </linearGradient>
        </defs>
        <CartesianGrid strokeDasharray="3 3" stroke={grid} vertical={false} />
        <XAxis dataKey="label" tick={{ fontSize: 12, fill: axis }} axisLine={false} tickLine={false} />
        <YAxis
          tick={{ fontSize: 12, fill: axis }}
          axisLine={false}
          tickLine={false}
          tickFormatter={(v: number) => formatCurrency(v).replace(/\s/g, ' ')}
          width={80}
        />
        <Tooltip content={<ChartTooltip />} cursor={{ fill: isDark ? '#ffffff10' : '#00000008' }} />
        <Legend
          iconType="circle"
          formatter={(value) => (
            <span className="text-xs text-gray-600 dark:text-gray-300">{value}</span>
          )}
        />
        <Bar dataKey="inflow" name="Entradas" fill="#10b981" radius={[4, 4, 0, 0]} barSize={14} />
        <Bar dataKey="outflow" name="Saídas" fill="#f43f5e" radius={[4, 4, 0, 0]} barSize={14} />
        <Area
          type="monotone"
          dataKey="projected_balance"
          name="Saldo projetado"
          stroke="#0ea5e9"
          strokeWidth={2.5}
          fill="url(#balanceGrad)"
        />
      </ComposedChart>
    </ResponsiveContainer>
    </div>
  );
}

export function ForecastPage() {
  const [months, setMonths] = useState(3);
  const { data, isLoading, isError } = useForecast(months);

  const hasNegative = useMemo(
    () => (data?.months ?? []).some((m) => m.projected_balance < 0),
    [data],
  );
  const alerts = data?.alerts ?? [];

  return (
    <div>
      <PageHeader
        eyebrow="Planejamento"
        title="Projeção de caixa"
        description="Estime a evolução do seu saldo com base nas contas a pagar e receber."
        action={
          <Select
            className="w-auto"
            options={monthsOptions}
            value={String(months)}
            onChange={(e) => setMonths(Number(e.target.value))}
          />
        }
      />

      {/* Alert banner */}
      {(hasNegative || alerts.length > 0) && (
        <div className="mb-4 flex items-start gap-3 rounded-xl border border-red-200 bg-red-50 px-4 py-3 dark:border-red-500/30 dark:bg-red-500/10">
          <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0 text-red-500" />
          <div className="text-sm text-red-600 dark:text-red-400">
            <p className="font-semibold">Atenção: saldo negativo projetado</p>
            {alerts.length > 0 ? (
              <ul className="mt-1 list-disc space-y-0.5 pl-4">
                {alerts.map((a) => (
                  <li key={a.month}>
                    <span className="font-medium">{monthLabel(a.month)}:</span> {a.message}
                  </li>
                ))}
              </ul>
            ) : (
              <p className="mt-0.5">
                Há meses com saldo projetado abaixo de zero no período selecionado.
              </p>
            )}
          </div>
        </div>
      )}

      {/* KPI cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard
          label="Saldo atual"
          value={<Money value={data?.current_balance ?? 0} />}
          icon={Wallet}
          tone="primary"
        />
        <StatCard
          label="Saldo projetado (fim)"
          value={<Money value={data?.end_balance ?? 0} />}
          icon={(data?.end_balance ?? 0) >= 0 ? TrendingUp : TrendingDown}
          tone={(data?.end_balance ?? 0) >= 0 ? 'sky' : 'red'}
        />
        <StatCard
          label="Menor saldo no período"
          value={<Money value={data?.lowest?.balance ?? 0} />}
          icon={TrendingDown}
          tone={(data?.lowest?.balance ?? 0) >= 0 ? 'amber' : 'red'}
          caption={data?.lowest?.month ? `em ${monthLabel(data.lowest.month)}` : undefined}
        />
      </div>

      {isError && (
        <Card className="mt-4 border-red-200 bg-red-50 dark:border-red-500/30 dark:bg-red-500/10">
          <p className="text-sm text-red-600 dark:text-red-400">
            Não foi possível carregar a projeção. Verifique sua conexão e tente novamente.
          </p>
        </Card>
      )}

      {/* Chart */}
      <Card className="mt-4">
        <CardHeader
          eyebrow="Fluxo projetado"
          title="Evolução do saldo"
          subtitle="Entradas e saídas estimadas por mês"
        />
        {isLoading || !data ? (
          <div className="skeleton h-[320px] w-full" />
        ) : data.months.length === 0 ? (
          <EmptyState
            icon={LineChart}
            title="Sem dados de projeção"
            description="Cadastre contas a pagar e receber com vencimento para gerar a projeção."
          />
        ) : (
          <ForecastChart months={data.months} />
        )}
      </Card>
    </div>
  );
}
