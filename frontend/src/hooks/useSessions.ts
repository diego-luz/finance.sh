import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { sessionService } from '@/services';
import { useToast } from '@/contexts/ToastContext';
import { ApiRequestError } from '@/lib/axios';
import { queryKeys } from './queryKeys';

/** Lists the authenticated user's active sessions. */
export function useSessions() {
  return useQuery({
    queryKey: queryKeys.sessions,
    queryFn: () => sessionService.list(),
  });
}

/** Revokes a single session by id, then refetches the list. */
export function useRevokeSession() {
  const queryClient = useQueryClient();
  const toast = useToast();
  return useMutation({
    mutationFn: (id: string) => sessionService.revoke(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.sessions });
      toast.success('Sessão encerrada.');
    },
    onError: (err: ApiRequestError) =>
      toast.error(err.message || 'Não foi possível encerrar a sessão.'),
  });
}

/** Revokes every other session, keeping the current one alive. */
export function useRevokeOtherSessions() {
  const queryClient = useQueryClient();
  const toast = useToast();
  return useMutation({
    mutationFn: (keepRefreshToken?: string) => sessionService.revokeOthers(keepRefreshToken),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.sessions });
      toast.success('As outras sessões foram encerradas.');
    },
    onError: (err: ApiRequestError) =>
      toast.error(err.message || 'Não foi possível encerrar as outras sessões.'),
  });
}
