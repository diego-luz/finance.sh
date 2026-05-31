import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { accountPrivacyService, downloadBlob } from '@/services';
import { useAuthStore } from '@/stores/authStore';
import { useToast } from '@/contexts/ToastContext';
import { ApiRequestError } from '@/lib/axios';

/** Exports the authenticated user's data as a downloaded JSON file (LGPD). */
export function useExportMyData() {
  const toast = useToast();
  return useMutation<Blob, ApiRequestError, void>({
    mutationFn: () => accountPrivacyService.exportData(),
    onSuccess: (blob) => {
      downloadBlob(blob, 'finance-sh-meus-dados.json');
      toast.success('Exportação concluída. O download foi iniciado.');
    },
    onError: (err: ApiRequestError) =>
      toast.error(err.message || 'Não foi possível exportar seus dados.'),
  });
}

/** Permanently deletes / anonymizes the account, then logs out (LGPD). */
export function useDeleteMyAccount() {
  const doLogout = useAuthStore((s) => s.logout);
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation<unknown, ApiRequestError, string>({
    mutationFn: (password: string) => accountPrivacyService.deleteAccount(password),
    onSuccess: () => {
      doLogout();
      queryClient.clear();
      navigate('/login', { replace: true });
      toast.success('Sua conta foi excluída. Sentiremos sua falta.');
    },
    onError: (err: ApiRequestError) =>
      toast.error(err.message || 'Não foi possível excluir a conta.'),
  });
}
