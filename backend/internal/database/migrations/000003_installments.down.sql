-- Reverts 000003: removes installment columns from transactions.
DROP INDEX IF EXISTS public.idx_transactions_installment_group_id;

ALTER TABLE public.transactions
    DROP COLUMN IF EXISTS installment_total,
    DROP COLUMN IF EXISTS installment_number,
    DROP COLUMN IF EXISTS installment_group_id;
