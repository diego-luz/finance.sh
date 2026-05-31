-- Feature: camada de super-admin de plataforma + suspensão de org + desativar usuário.
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS super_admin boolean NOT NULL DEFAULT false;
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS disabled boolean NOT NULL DEFAULT false;
ALTER TABLE public.organizations ADD COLUMN IF NOT EXISTS suspended boolean NOT NULL DEFAULT false;
