-- Reverts 000005.
DROP INDEX IF EXISTS public.idx_transactions_external_id;
ALTER TABLE public.transactions DROP COLUMN IF EXISTS external_id;
