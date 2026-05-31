import { QueryClientProvider } from '@tanstack/react-query';
import { RouterProvider } from 'react-router-dom';
import { queryClient } from '@/lib/queryClient';
import { ToastProvider } from '@/contexts/ToastContext';
import { AuthProvider } from '@/contexts/AuthProvider';
import { CookieConsentBanner } from '@/components/CookieConsentBanner';
import { InstallPrompt } from '@/components/InstallPrompt';
import { router } from '@/routes';

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <AuthProvider>
          <RouterProvider router={router} />
          <CookieConsentBanner />
          <InstallPrompt />
        </AuthProvider>
      </ToastProvider>
    </QueryClientProvider>
  );
}
