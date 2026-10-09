-- 2FA: o último intervalo TOTP (30 s) aceito de cada conta. Um código só vale
-- se for de um intervalo mais novo, então não pode ser reusado.
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS totp_last_step bigint NOT NULL DEFAULT 0;
