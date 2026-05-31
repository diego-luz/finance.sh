-- Reverts 000010.
ALTER TABLE public.users DROP COLUMN IF EXISTS must_change_password;
