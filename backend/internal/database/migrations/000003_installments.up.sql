-- Feature: parcelamento. Adds installment grouping columns to transactions so a
-- single purchase can be split into N linked monthly transactions.
-- All additive/nullable so existing rows/flows are unaffected.

ALTER TABLE public.transactions
    ADD COLUMN IF NOT EXISTS installment_group_id uuid,
    ADD COLUMN IF NOT EXISTS installment_number bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS installment_total bigint DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_transactions_installment_group_id
    ON public.transactions USING btree (installment_group_id);
