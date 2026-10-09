-- Reverts 000016. Os tokens continuam como hash: convites pendentes criados
-- depois da 000016 deixam de poder ser aceitos e precisam ser refeitos.
ALTER TABLE public.invitations DROP COLUMN IF EXISTS expires_at;
