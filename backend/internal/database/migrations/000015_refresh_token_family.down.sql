-- Reverts 000015.
DROP INDEX IF EXISTS public.idx_refresh_tokens_family_id;
ALTER TABLE public.refresh_tokens DROP COLUMN IF EXISTS revoked_at;
ALTER TABLE public.refresh_tokens DROP COLUMN IF EXISTS session_expires_at;
ALTER TABLE public.refresh_tokens DROP COLUMN IF EXISTS family_id;
