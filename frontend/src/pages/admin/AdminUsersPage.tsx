import { useEffect, useState } from 'react';
import { Navigate } from 'react-router-dom';
import {
  Users,
  Search,
  ChevronLeft,
  ChevronRight,
  ShieldCheck,
  Ban,
  RotateCcw,
  BadgeCheck,
  KeyRound,
} from 'lucide-react';
import { PageHeader } from '@/components/PageHeader';
import {
  Badge,
  Button,
  Card,
  EmptyState,
  Input,
  Table,
  THead,
  Th,
  TBody,
  Tr,
  Td,
  SkeletonRows,
} from '@/components/ui';
import {
  useAdminUsers,
  useDisableUser,
  useEnableUser,
  useDebouncedValue,
} from '@/hooks';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';
import { formatDateShort } from '@/lib/date';
import { AdminResetPasswordModal } from '@/components/admin/AdminResetPasswordModal';
import type { AdminUser } from '@/types';

const PER_PAGE = 20;

export function AdminUsersPage() {
  const user = useAuthStore((s) => s.user);
  const toast = useToast();

  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const [resetUser, setResetUser] = useState<AdminUser | null>(null);

  const debouncedSearch = useDebouncedValue(search, 400);
  const { data, isLoading, isFetching, isError } = useAdminUsers({
    search: debouncedSearch || undefined,
    page,
    per_page: PER_PAGE,
  });
  const disable = useDisableUser();
  const enable = useEnableUser();

  useEffect(() => {
    if (isError) toast.error('Não foi possível carregar os usuários.');
  }, [isError, toast]);

  useEffect(() => {
    setPage(1);
  }, [debouncedSearch]);

  // Hooks run before the guard to preserve hook order (mirrors MembersPage).
  if (!user?.super_admin) {
    return <Navigate to="/" replace />;
  }

  const meta = data?.meta;
  const pages = meta?.pages ?? 1;
  const total = meta?.total ?? 0;
  const users = data?.data ?? [];

  const handleToggle = (u: AdminUser) => {
    if (u.disabled) {
      if (window.confirm(`Reativar o usuário ${u.name}?`)) {
        enable.mutate(u.id);
      }
    } else if (
      window.confirm(`Desativar o usuário ${u.name}? Ele perderá o acesso à plataforma.`)
    ) {
      disable.mutate(u.id);
    }
  };

  return (
    <div>
      <PageHeader
        eyebrow="Plataforma"
        title="Usuários"
        description="Gerencie todas as contas de usuário da plataforma."
      />

      <Card className="mb-4 p-4">
        <Input
          placeholder="Buscar por nome ou e-mail…"
          leftIcon={<Search className="h-4 w-4" />}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          aria-label="Buscar usuários"
        />
      </Card>

      <Card className="p-0">
        {isLoading ? (
          <Table>
            <THead>
              <Th>Usuário</Th>
              <Th>Tipo</Th>
              <Th>Verificado</Th>
              <Th>Status</Th>
              <Th className="text-right">Orgs</Th>
              <Th>Criado em</Th>
              <Th className="text-right">Ações</Th>
            </THead>
            <TBody>
              <SkeletonRows rows={8} cols={7} />
            </TBody>
          </Table>
        ) : users.length === 0 ? (
          <div className="p-6">
            <EmptyState
              icon={Users}
              title="Nenhum usuário"
              description="Ajuste a busca para encontrar contas."
            />
          </div>
        ) : (
          <Table>
            <THead>
              <Th>Usuário</Th>
              <Th>Tipo</Th>
              <Th>Verificado</Th>
              <Th>Status</Th>
              <Th className="text-right">Orgs</Th>
              <Th>Criado em</Th>
              <Th className="text-right">Ações</Th>
            </THead>
            <TBody>
              {users.map((u) => (
                <Tr key={u.id}>
                  <Td>
                    <div className="min-w-0">
                      <p className="truncate font-medium text-gray-900 dark:text-gray-100">
                        {u.name}
                      </p>
                      <p className="truncate text-xs text-gray-400">{u.email}</p>
                    </div>
                  </Td>
                  <Td>
                    {u.super_admin ? (
                      <Badge variant="info">
                        <ShieldCheck className="h-3.5 w-3.5" /> Super admin
                      </Badge>
                    ) : (
                      <Badge variant="neutral">Usuário</Badge>
                    )}
                  </Td>
                  <Td>
                    {u.email_verified ? (
                      <span className="inline-flex items-center gap-1 text-xs font-medium text-primary">
                        <BadgeCheck className="h-3.5 w-3.5" /> Sim
                      </span>
                    ) : (
                      <span className="text-xs text-gray-400">Não</span>
                    )}
                  </Td>
                  <Td>
                    {u.disabled ? (
                      <Badge variant="danger" dot>
                        Desativado
                      </Badge>
                    ) : (
                      <Badge variant="success" dot>
                        Ativo
                      </Badge>
                    )}
                  </Td>
                  <Td className="text-right tabular-nums">{u.organizations.length}</Td>
                  <Td className="whitespace-nowrap text-gray-500 dark:text-gray-400">
                    {formatDateShort(u.created_at)}
                  </Td>
                  <Td>
                    <div className="flex items-center justify-end gap-1.5">
                      {!u.disabled && (
                        <button
                          onClick={() => setResetUser(u)}
                          className="rounded-md p-1.5 text-gray-400 transition hover:bg-primary/10 hover:text-primary"
                          aria-label="Resetar senha do usuário"
                          title="Resetar senha"
                        >
                          <KeyRound className="h-4 w-4" />
                        </button>
                      )}
                      {u.super_admin ? (
                        <span
                          className="text-gray-300 dark:text-gray-600"
                          title="Super admins não podem ser desativados pela interface"
                        >
                          —
                        </span>
                      ) : (
                        <button
                          onClick={() => handleToggle(u)}
                          className={
                            u.disabled
                              ? 'rounded-md p-1.5 text-gray-400 transition hover:bg-primary/10 hover:text-primary'
                              : 'rounded-md p-1.5 text-gray-400 transition hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-500/10'
                          }
                          aria-label={u.disabled ? 'Reativar usuário' : 'Desativar usuário'}
                          title={u.disabled ? 'Reativar' : 'Desativar'}
                        >
                          {u.disabled ? (
                            <RotateCcw className="h-4 w-4" />
                          ) : (
                            <Ban className="h-4 w-4" />
                          )}
                        </button>
                      )}
                    </div>
                  </Td>
                </Tr>
              ))}
            </TBody>
          </Table>
        )}

        {users.length > 0 && (
          <div className="flex items-center justify-between border-t border-gray-100 px-4 py-3 dark:border-ink-border">
            <p className="text-xs text-gray-500 dark:text-gray-400">
              {total} usuário{total === 1 ? '' : 's'} · página {page} de {pages}
              {isFetching && <span className="ml-2 text-primary">atualizando…</span>}
            </p>
            <div className="flex items-center gap-1.5">
              <Button
                variant="outline"
                size="sm"
                disabled={page <= 1 || isFetching}
                onClick={() => setPage((p) => p - 1)}
              >
                <ChevronLeft className="h-4 w-4" /> Anterior
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={page >= pages || isFetching}
                onClick={() => setPage((p) => p + 1)}
              >
                Próxima <ChevronRight className="h-4 w-4" />
              </Button>
            </div>
          </div>
        )}
      </Card>

      <AdminResetPasswordModal user={resetUser} onClose={() => setResetUser(null)} />
    </div>
  );
}
