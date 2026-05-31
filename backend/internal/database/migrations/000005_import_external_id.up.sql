-- Feature: import OFX/CSV. Adds external_id for dedup (OFX FITID or CSV hash)
-- so re-importing the same statement does not create duplicate transactions.
ALTER TABLE public.transactions ADD COLUMN IF NOT EXISTS external_id text;
CREATE INDEX IF NOT EXISTS idx_transactions_external_id ON public.transactions USING btree (external_id);
