-- Feature: senha inicial no provisionamento + obrigação de troca no 1º login.
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS must_change_password boolean NOT NULL DEFAULT false;
