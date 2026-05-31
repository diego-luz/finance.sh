import { api, unwrap } from '@/lib/axios';
import type {
  ApiEnvelope,
  CreditCard,
  CreditCardPayload,
  Invoice,
  InvoiceDetail,
  InvoicePayPayload,
} from '@/types';

export const creditCardService = {
  list: () => unwrap<CreditCard[]>(api.get<ApiEnvelope<CreditCard[]>>('/credit-cards')),

  create: (payload: CreditCardPayload) =>
    unwrap<CreditCard>(api.post<ApiEnvelope<CreditCard>>('/credit-cards', payload)),

  update: (id: string, payload: CreditCardPayload) =>
    unwrap<CreditCard>(api.put<ApiEnvelope<CreditCard>>(`/credit-cards/${id}`, payload)),

  remove: (id: string) => api.delete(`/credit-cards/${id}`),

  listInvoices: (cardId: string, limit = 12) =>
    unwrap<Invoice[]>(
      api.get<ApiEnvelope<Invoice[]>>(`/credit-cards/${cardId}/invoices`, {
        params: { limit },
      }),
    ),

  getInvoice: (cardId: string, reference: string) =>
    unwrap<InvoiceDetail>(
      api.get<ApiEnvelope<InvoiceDetail>>(`/credit-cards/${cardId}/invoices/${reference}`),
    ),

  payInvoice: (cardId: string, reference: string, payload: InvoicePayPayload) =>
    unwrap<Invoice>(
      api.post<ApiEnvelope<Invoice>>(
        `/credit-cards/${cardId}/invoices/${reference}/pay`,
        payload,
      ),
    ),
};
