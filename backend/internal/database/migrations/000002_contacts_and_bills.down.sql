-- Reverts 000002: removes bill columns from transactions and drops contacts.
DROP INDEX IF EXISTS public.idx_transactions_due_date;
DROP INDEX IF EXISTS public.idx_transactions_contact_id;

ALTER TABLE public.transactions
    DROP COLUMN IF EXISTS paid_at,
    DROP COLUMN IF EXISTS due_date,
    DROP COLUMN IF EXISTS contact_id;

DROP TABLE IF EXISTS public.contacts;
