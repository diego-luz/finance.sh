-- Feature: tags (rótulos) — many-to-many com transações + busca global usa o nome.
CREATE TABLE IF NOT EXISTS public.tags (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    organization_id uuid NOT NULL,
    name text NOT NULL,
    color character varying(9) DEFAULT '#6b7280'::character varying,
    CONSTRAINT tags_pkey PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_tags_deleted_at ON public.tags USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_tags_organization_id ON public.tags USING btree (organization_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_org_tag_name ON public.tags USING btree (organization_id, name);

-- Join table transação <-> tag.
CREATE TABLE IF NOT EXISTS public.transaction_tags (
    transaction_id uuid NOT NULL,
    tag_id uuid NOT NULL,
    CONSTRAINT transaction_tags_pkey PRIMARY KEY (transaction_id, tag_id),
    CONSTRAINT fk_transaction_tags_transaction FOREIGN KEY (transaction_id) REFERENCES public.transactions(id) ON DELETE CASCADE,
    CONSTRAINT fk_transaction_tags_tag FOREIGN KEY (tag_id) REFERENCES public.tags(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_transaction_tags_tag_id ON public.transaction_tags USING btree (tag_id);
