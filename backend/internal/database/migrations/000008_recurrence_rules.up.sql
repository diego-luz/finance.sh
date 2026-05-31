-- Feature: motor de recorrência. Regras (template + agenda) que geram
-- transações automaticamente via worker.
CREATE TABLE IF NOT EXISTS public.recurrence_rules (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    organization_id uuid NOT NULL,
    type character varying(10) NOT NULL,
    amount bigint NOT NULL,
    description text NOT NULL,
    account_id uuid NOT NULL,
    category_id uuid,
    contact_id uuid,
    paid boolean NOT NULL DEFAULT false,
    frequency character varying(10) NOT NULL,
    "interval" bigint NOT NULL DEFAULT 1,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone,
    max_occurrences bigint NOT NULL DEFAULT 0,
    occurrences_count bigint NOT NULL DEFAULT 0,
    next_run_date timestamp with time zone NOT NULL,
    last_generated_at timestamp with time zone,
    active boolean NOT NULL DEFAULT true,
    CONSTRAINT recurrence_rules_pkey PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_recurrence_rules_deleted_at ON public.recurrence_rules USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_recurrence_rules_organization_id ON public.recurrence_rules USING btree (organization_id);
CREATE INDEX IF NOT EXISTS idx_recurrence_rules_account_id ON public.recurrence_rules USING btree (account_id);
CREATE INDEX IF NOT EXISTS idx_recurrence_rules_category_id ON public.recurrence_rules USING btree (category_id);
CREATE INDEX IF NOT EXISTS idx_recurrence_rules_contact_id ON public.recurrence_rules USING btree (contact_id);
CREATE INDEX IF NOT EXISTS idx_recurrence_rules_next_run_date ON public.recurrence_rules USING btree (next_run_date);
