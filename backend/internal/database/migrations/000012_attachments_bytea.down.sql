ALTER TABLE public.attachments DROP COLUMN IF EXISTS data;
ALTER TABLE public.attachments DROP COLUMN IF EXISTS size_bytes;
ALTER TABLE public.attachments ALTER COLUMN object_key SET NOT NULL;
