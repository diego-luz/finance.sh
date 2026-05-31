-- Feature: contas a pagar/receber + projeção de fluxo de caixa.
-- Adds the `contacts` table (fornecedores/clientes) and the bill-related
-- columns on transactions (contact_id, due_date = vencimento, paid_at).
-- All additive and nullable, so existing rows/flows are unaffected.

CREATE TABLE IF NOT EXISTS public.contacts (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    organization_id uuid NOT NULL,
    name text NOT NULL,
    type character varying(10) DEFAULT 'both'::character varying,
    document character varying(20),
    email text,
    phone character varying(30),
    notes text,
    CONSTRAINT contacts_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_contacts_deleted_at ON public.contacts USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_contacts_organization_id ON public.contacts USING btree (organization_id);

ALTER TABLE public.transactions
    ADD COLUMN IF NOT EXISTS contact_id uuid,
    ADD COLUMN IF NOT EXISTS due_date timestamp with time zone,
    ADD COLUMN IF NOT EXISTS paid_at timestamp with time zone;

CREATE INDEX IF NOT EXISTS idx_transactions_contact_id ON public.transactions USING btree (contact_id);
CREATE INDEX IF NOT EXISTS idx_transactions_due_date ON public.transactions USING btree (due_date);
