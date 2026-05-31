-- Feature: captura de leads pela landing pública. Super-admin gerencia em
-- /admin/leads e pode provisionar org+dono a partir de um lead.
CREATE TABLE IF NOT EXISTS public.leads (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    email text NOT NULL,
    phone character varying(30),
    organization_name text,
    profile_type character varying(20),
    message text,
    status character varying(20) NOT NULL DEFAULT 'new',
    notes text,
    converted_org_id uuid,
    ip text,
    user_agent text,
    CONSTRAINT leads_pkey PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_leads_deleted_at ON public.leads USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_leads_email ON public.leads USING btree (email);
CREATE INDEX IF NOT EXISTS idx_leads_status ON public.leads USING btree (status);
