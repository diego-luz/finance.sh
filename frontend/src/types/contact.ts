export type ContactType = 'customer' | 'supplier' | 'both';

export interface Contact {
  id: string;
  name: string;
  type: ContactType;
  document?: string;
  email?: string;
  phone?: string;
  notes?: string;
}

export interface ContactPayload {
  name: string;
  type: ContactType;
  document?: string;
  email?: string;
  phone?: string;
  notes?: string;
}
