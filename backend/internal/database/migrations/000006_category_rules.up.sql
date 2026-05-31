-- Feature: categorização automática. Keyword/regex rules that map a transaction
-- description to a category (applied on import and as a suggestion in the form).

CREATE TABLE IF NOT EXISTS public.category_rules (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    organization_id uuid NOT NULL,
    pattern text NOT NULL,
    match_type character varying(10) DEFAULT 'contains'::character varying,
    category_id uuid NOT NULL,
    priority bigint DEFAULT 0,
    active boolean DEFAULT true,
    CONSTRAINT category_rules_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_category_rules_deleted_at ON public.category_rules USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_category_rules_organization_id ON public.category_rules USING btree (organization_id);
CREATE INDEX IF NOT EXISTS idx_category_rules_category_id ON public.category_rules USING btree (category_id);
