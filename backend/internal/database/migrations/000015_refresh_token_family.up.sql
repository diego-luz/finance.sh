-- Rotação de refresh token com detecção de reuso: cada login abre uma
-- "família" (family_id) que as rotações herdam, com um prazo absoluto
-- (session_expires_at). Reapresentar um token já trocado revoga a família.
ALTER TABLE public.refresh_tokens ADD COLUMN IF NOT EXISTS family_id uuid;
ALTER TABLE public.refresh_tokens ADD COLUMN IF NOT EXISTS session_expires_at timestamp with time zone;
ALTER TABLE public.refresh_tokens ADD COLUMN IF NOT EXISTS revoked_at timestamp with time zone;
-- tokens de antes: cada um é a própria família, com 30 dias desde a criação
UPDATE public.refresh_tokens SET family_id = id WHERE family_id IS NULL;
UPDATE public.refresh_tokens
   SET session_expires_at = COALESCE(created_at, now()) + interval '30 days'
 WHERE session_expires_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_family_id ON public.refresh_tokens (family_id);
