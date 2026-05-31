import { useState } from 'react';
import { Navigate } from 'react-router-dom';
import { UserPlus, Users, Trash2, Clock, ShieldAlert, X } from 'lucide-react';
import { PageHeader } from '@/components/PageHeader';
import {
  Button,
  Card,
  CardHeader,
  Avatar,
  Badge,
  EmptyState,
  Select,
  Table,
  THead,
  Th,
  TBody,
  Tr,
  Td,
  SkeletonRows,
} from '@/components/ui';
import { InviteMemberModal } from '@/components/forms/InviteMemberModal';
import {
  useMembers,
  useInvitations,
  useUpdateMemberRole,
  useRemoveMember,
  useRevokeInvitation,
  useIsAdmin,
  useIsOwner,
} from '@/hooks';
import { useAuthStore } from '@/stores/authStore';
import { formatDateShort } from '@/lib/date';
import type { Member, OrgRole } from '@/types';

const roleLabels: Record<string, string> = {
  owner: 'Proprietário',
  admin: 'Administrador',
  member: 'Membro',
  viewer: 'Visualizador',
};

const roleSelectOptions: { value: OrgRole; label: string }[] = [
  { value: 'admin', label: 'Administrador' },
  { value: 'member', label: 'Membro' },
  { value: 'viewer', label: 'Visualizador' },
];

function MemberRow({
  member,
  isSelf,
  canManage,
  onRoleChange,
  onRemove,
}: {
  member: Member;
  isSelf: boolean;
  canManage: boolean;
  onRoleChange: (role: OrgRole) => void;
  onRemove: () => void;
}) {
  const isOwner = member.role === 'owner';
  return (
    <Tr>
      <Td>
        <div className="flex items-center gap-3">
          <Avatar name={member.user.name} src={member.user.avatar_url} size="md" />
          <div className="min-w-0">
            <p className="flex items-center gap-2 truncate font-medium text-gray-900 dark:text-gray-100">
              {member.user.name}
              {isSelf && <Badge variant="neutral">Você</Badge>}
            </p>
            <p className="truncate text-xs text-gray-400">{member.user.email}</p>
          </div>
        </div>
      </Td>
      <Td>
        {canManage && !isOwner && !isSelf ? (
          <Select
            className="w-44"
            options={roleSelectOptions}
            value={member.role}
            onChange={(e) => onRoleChange(e.target.value as OrgRole)}
          />
        ) : (
          <Badge variant={isOwner ? 'success' : 'neutral'}>
            {roleLabels[member.role] ?? member.role}
          </Badge>
        )}
      </Td>
      <Td className="text-right">
        {canManage && !isOwner && !isSelf ? (
          <button
            onClick={onRemove}
            className="rounded-md p-1.5 text-gray-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-500/10"
            aria-label="Remover membro"
          >
            <Trash2 className="h-4 w-4" />
          </button>
        ) : (
          <span className="text-gray-300 dark:text-gray-600">—</span>
        )}
      </Td>
    </Tr>
  );
}

export function MembersPage() {
  const canManage = useIsAdmin();
  const isOwner = useIsOwner();
  const currentUser = useAuthStore((s) => s.user);

  const { data: members, isLoading } = useMembers();
  const { data: invitations, isLoading: invLoading } = useInvitations();
  const updateRole = useUpdateMemberRole();
  const removeMember = useRemoveMember();
  const revokeInvitation = useRevokeInvitation();
  const [inviteOpen, setInviteOpen] = useState(false);

  // Route guard: only owners/admins may access the team management page.
  if (!canManage) {
    return <Navigate to="/" replace />;
  }

  const pending = (invitations ?? []).filter((i) => !i.accepted);

  const handleRemove = (m: Member) => {
    if (window.confirm(`Remover ${m.user.name} da organização?`)) {
      removeMember.mutate(m.id);
    }
  };

  return (
    <div>
      <PageHeader
        eyebrow="Organização"
        title="Equipe"
        description="Gerencie quem tem acesso à organização e suas permissões."
        action={
          <Button onClick={() => setInviteOpen(true)}>
            <UserPlus className="h-4 w-4" /> Convidar
          </Button>
        }
      />

      <Card className="p-0">
        <div className="px-5 pt-5">
          <CardHeader
            title="Membros"
            subtitle="Pessoas com acesso a esta organização."
            action={<Users className="h-5 w-5 text-gray-400" />}
          />
        </div>
        {isLoading ? (
          <Table>
            <THead>
              <Th>Membro</Th>
              <Th>Função</Th>
              <Th className="w-16" />
            </THead>
            <TBody>
              <SkeletonRows rows={4} cols={3} />
            </TBody>
          </Table>
        ) : !members || members.length === 0 ? (
          <div className="p-6">
            <EmptyState icon={Users} title="Nenhum membro" description="Convide pessoas para colaborar." />
          </div>
        ) : (
          <Table>
            <THead>
              <Th>Membro</Th>
              <Th>Função</Th>
              <Th className="w-16 text-right">Ações</Th>
            </THead>
            <TBody>
              {members.map((m) => (
                <MemberRow
                  key={m.id}
                  member={m}
                  isSelf={m.user.id === currentUser?.id}
                  canManage={canManage}
                  onRoleChange={(role) => updateRole.mutate({ id: m.id, payload: { role } })}
                  onRemove={() => handleRemove(m)}
                />
              ))}
            </TBody>
          </Table>
        )}
      </Card>

      {/* Pending invitations */}
      <Card className="mt-4">
        <CardHeader
          title="Convites pendentes"
          subtitle="Convites enviados que ainda não foram aceitos."
          action={<Clock className="h-5 w-5 text-gray-400" />}
        />
        {invLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 2 }).map((_, i) => (
              <div key={i} className="skeleton h-14 w-full rounded-xl" />
            ))}
          </div>
        ) : pending.length === 0 ? (
          <p className="py-6 text-center text-sm text-gray-400">Nenhum convite pendente.</p>
        ) : (
          <ul className="space-y-2">
            {pending.map((inv) => (
              <li
                key={inv.id}
                className="flex items-center justify-between gap-3 rounded-xl border border-gray-100 p-3 dark:border-ink-border"
              >
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium text-gray-900 dark:text-gray-100">
                    {inv.email}
                  </p>
                  <p className="text-xs text-gray-400">
                    {roleLabels[inv.role] ?? inv.role} · enviado {formatDateShort(inv.created_at)}
                  </p>
                </div>
                <div className="flex shrink-0 items-center gap-2">
                  <Badge variant="warning" dot>
                    Pendente
                  </Badge>
                  <button
                    onClick={() => {
                      if (window.confirm(`Revogar o convite de ${inv.email}?`)) {
                        revokeInvitation.mutate(inv.id);
                      }
                    }}
                    className="rounded-md p-1.5 text-gray-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-500/10"
                    aria-label="Revogar convite"
                  >
                    <X className="h-4 w-4" />
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </Card>

      {!isOwner && (
        <p className="mt-4 flex items-center gap-2 text-xs text-gray-400">
          <ShieldAlert className="h-3.5 w-3.5" />
          Apenas o proprietário pode gerenciar o plano e remover administradores.
        </p>
      )}

      <InviteMemberModal open={inviteOpen} onClose={() => setInviteOpen(false)} />
    </div>
  );
}
