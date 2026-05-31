import {
  MonitorSmartphone,
  Smartphone,
  Monitor,
  Globe,
  LogOut,
  ShieldX,
} from 'lucide-react';
import { Card, CardHeader, Button, Skeleton, EmptyState } from '@/components/ui';
import { useSessions, useRevokeSession, useRevokeOtherSessions } from '@/hooks';
import { useAuthStore } from '@/stores/authStore';
import { formatDate } from '@/lib/date';
import type { Session } from '@/types';

/** Best-effort device label parsed from a user-agent string. */
function parseDevice(ua: string): { label: string; mobile: boolean } {
  const s = (ua || '').toLowerCase();
  const mobile = /mobile|android|iphone|ipad|ipod/.test(s);

  let os = 'Dispositivo';
  if (/windows/.test(s)) os = 'Windows';
  else if (/iphone|ipad|ipod/.test(s)) os = 'iOS';
  else if (/android/.test(s)) os = 'Android';
  else if (/mac os|macintosh/.test(s)) os = 'macOS';
  else if (/linux/.test(s)) os = 'Linux';

  let browser = '';
  if (/edg\//.test(s)) browser = 'Edge';
  else if (/chrome|crios/.test(s)) browser = 'Chrome';
  else if (/firefox|fxios/.test(s)) browser = 'Firefox';
  else if (/safari/.test(s)) browser = 'Safari';

  const label = browser ? `${browser} · ${os}` : os;
  return { label, mobile };
}

function SessionRow({ session }: { session: Session }) {
  const revoke = useRevokeSession();
  const device = parseDevice(session.user_agent);
  const Icon = device.mobile ? Smartphone : Monitor;

  return (
    <div className="flex items-center gap-3 rounded-lg border border-gray-100 px-3.5 py-3 dark:border-ink-border">
      <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
        <Icon className="h-[18px] w-[18px]" />
      </span>
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium text-gray-900 dark:text-gray-100">
          {device.label}
        </p>
        <p className="mt-0.5 flex flex-wrap items-center gap-x-2 gap-y-0.5 text-xs text-gray-400">
          <span className="inline-flex items-center gap-1">
            <Globe className="h-3 w-3" /> {session.ip || 'IP desconhecido'}
          </span>
          <span aria-hidden>·</span>
          <span>Criada em {formatDate(session.created_at)}</span>
          {session.expires_at && (
            <>
              <span aria-hidden>·</span>
              <span>Expira em {formatDate(session.expires_at)}</span>
            </>
          )}
        </p>
      </div>
      <Button
        variant="ghost"
        size="sm"
        loading={revoke.isPending && revoke.variables === session.id}
        onClick={() => revoke.mutate(session.id)}
      >
        <LogOut className="h-4 w-4" /> Encerrar
      </Button>
    </div>
  );
}

export function SessionsCard() {
  const { data: sessions, isLoading, isError } = useSessions();
  const revokeOthers = useRevokeOtherSessions();
  const refreshToken = useAuthStore((s) => s.refreshToken);

  const list = sessions ?? [];

  return (
    <Card className="mt-4">
      <CardHeader
        eyebrow="Segurança"
        title="Sessões ativas"
        subtitle="Dispositivos e navegadores conectados à sua conta."
        action={
          list.length > 1 ? (
            <Button
              variant="outline"
              size="sm"
              loading={revokeOthers.isPending}
              onClick={() => revokeOthers.mutate(refreshToken ?? undefined)}
            >
              <ShieldX className="h-4 w-4" /> Encerrar outras sessões
            </Button>
          ) : undefined
        }
      />

      {isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-16 rounded-lg" />
          ))}
        </div>
      ) : isError ? (
        <p className="text-sm text-gray-400">Não foi possível carregar as sessões.</p>
      ) : list.length === 0 ? (
        <EmptyState
          icon={MonitorSmartphone}
          title="Nenhuma sessão ativa"
          description="Não encontramos sessões ativas para a sua conta."
        />
      ) : (
        <div className="space-y-2">
          {list.map((s) => (
            <SessionRow key={s.id} session={s} />
          ))}
        </div>
      )}
    </Card>
  );
}
