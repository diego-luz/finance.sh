-- Reverts 000017.
ALTER TABLE public.users DROP COLUMN IF EXISTS totp_last_step;
