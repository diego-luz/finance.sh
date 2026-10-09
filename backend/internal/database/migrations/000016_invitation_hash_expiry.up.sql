-- Convites: o token passa a ser guardado só como hash (SHA-256 em hex, como
-- refresh e reset) e ganha validade. Convites pendentes de antes valem 7 dias
-- a partir de agora.
ALTER TABLE public.invitations ADD COLUMN IF NOT EXISTS expires_at timestamp with time zone;
UPDATE public.invitations
   SET token = encode(sha256(convert_to(token, 'UTF8')), 'hex'),
       expires_at = now() + interval '7 days'
 WHERE expires_at IS NULL;
