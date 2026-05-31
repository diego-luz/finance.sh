-- Store attachment blobs directly in Postgres BYTEA. TOAST handles large
-- columns transparently; the target is receipts in the 200KB-2MB range, 10MB cap.
ALTER TABLE public.attachments
    ADD COLUMN IF NOT EXISTS data BYTEA;

-- object_key is unused; make it nullable so INSERTs carrying only the BYTEA blob
-- are accepted.
ALTER TABLE public.attachments
    ALTER COLUMN object_key DROP NOT NULL;

-- Byte size is stored in size_bytes (mirrored) for forward compatibility.
ALTER TABLE public.attachments
    ADD COLUMN IF NOT EXISTS size_bytes BIGINT;
