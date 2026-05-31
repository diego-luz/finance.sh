-- Feature: anexos de comprovante. Attachment metadata table. The blob itself is
-- added later as a BYTEA column (migration 000012); object_key is unused.

CREATE TABLE IF NOT EXISTS public.attachments (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    organization_id uuid NOT NULL,
    transaction_id uuid,
    file_name text NOT NULL,
    content_type text,
    size bigint,
    object_key text NOT NULL,
    uploaded_by uuid,
    CONSTRAINT attachments_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_attachments_deleted_at ON public.attachments USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_attachments_organization_id ON public.attachments USING btree (organization_id);
CREATE INDEX IF NOT EXISTS idx_attachments_transaction_id ON public.attachments USING btree (transaction_id);
