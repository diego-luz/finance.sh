import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { notificationService } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { queryKeys } from './queryKeys';
import { ApiRequestError } from '@/lib/axios';
import { useToast } from '@/contexts/ToastContext';

export function useNotifications() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  return useQuery({
    queryKey: queryKeys.notifications(orgId),
    queryFn: () => notificationService.list(),
    enabled: Boolean(orgId),
    refetchInterval: 60_000,
    refetchOnWindowFocus: true,
  });
}

export function useMarkNotificationRead() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => notificationService.markRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.notifications(orgId) });
    },
  });
}

export function useMarkAllNotificationsRead() {
  const orgId = useAuthStore((s) => s.currentOrgId);
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation({
    mutationFn: () => notificationService.markAllRead(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.notifications(orgId) });
    },
    onError: (err: ApiRequestError) => toast.error(err.message),
  });
}
