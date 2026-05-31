-- Fully open-source, no plan/edition tiers and no client-provisioning:
-- drop the subscriptions and leads tables and the organizations.suspended flag.
DROP TABLE IF EXISTS public.subscriptions;
DROP TABLE IF EXISTS public.leads;
ALTER TABLE public.organizations DROP COLUMN IF EXISTS suspended;
