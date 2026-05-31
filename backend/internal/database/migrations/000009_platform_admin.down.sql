-- Reverts 000009.
ALTER TABLE public.organizations DROP COLUMN IF EXISTS suspended;
ALTER TABLE public.users DROP COLUMN IF EXISTS disabled;
ALTER TABLE public.users DROP COLUMN IF EXISTS super_admin;
