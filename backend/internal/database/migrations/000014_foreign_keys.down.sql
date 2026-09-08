-- Reverte 000014: remove as foreign keys e o índice composto. Os UPDATEs de
-- reparo do up (referências opcionais penduradas -> NULL) não são revertidos:
-- eles apontavam para linhas que não existem mais, então não há valor a restaurar.
DROP INDEX IF EXISTS public.idx_transactions_org_date;

ALTER TABLE public.notifications   DROP CONSTRAINT IF EXISTS fk_notifications_organization;
ALTER TABLE public.invitations     DROP CONSTRAINT IF EXISTS fk_invitations_organization;
ALTER TABLE public.memberships     DROP CONSTRAINT IF EXISTS fk_memberships_organization;
ALTER TABLE public.attachments     DROP CONSTRAINT IF EXISTS fk_attachments_organization;
ALTER TABLE public.recurrence_rules DROP CONSTRAINT IF EXISTS fk_recurrence_rules_organization;
ALTER TABLE public.category_rules  DROP CONSTRAINT IF EXISTS fk_category_rules_organization;
ALTER TABLE public.tags            DROP CONSTRAINT IF EXISTS fk_tags_organization;
ALTER TABLE public.goals           DROP CONSTRAINT IF EXISTS fk_goals_organization;
ALTER TABLE public.budgets         DROP CONSTRAINT IF EXISTS fk_budgets_organization;
ALTER TABLE public.transactions    DROP CONSTRAINT IF EXISTS fk_transactions_organization;
ALTER TABLE public.credit_cards    DROP CONSTRAINT IF EXISTS fk_credit_cards_organization;
ALTER TABLE public.contacts        DROP CONSTRAINT IF EXISTS fk_contacts_organization;
ALTER TABLE public.categories      DROP CONSTRAINT IF EXISTS fk_categories_organization;
ALTER TABLE public.accounts        DROP CONSTRAINT IF EXISTS fk_accounts_organization;

ALTER TABLE public.email_verifications DROP CONSTRAINT IF EXISTS fk_email_verifications_user;
ALTER TABLE public.recovery_codes  DROP CONSTRAINT IF EXISTS fk_recovery_codes_user;
ALTER TABLE public.password_resets DROP CONSTRAINT IF EXISTS fk_password_resets_user;
ALTER TABLE public.refresh_tokens  DROP CONSTRAINT IF EXISTS fk_refresh_tokens_user;
ALTER TABLE public.memberships     DROP CONSTRAINT IF EXISTS fk_memberships_user;
ALTER TABLE public.organizations   DROP CONSTRAINT IF EXISTS fk_organizations_owner;
ALTER TABLE public.category_rules  DROP CONSTRAINT IF EXISTS fk_category_rules_category;
ALTER TABLE public.budgets         DROP CONSTRAINT IF EXISTS fk_budgets_category;
ALTER TABLE public.recurrence_rules DROP CONSTRAINT IF EXISTS fk_recurrence_rules_account;
ALTER TABLE public.transactions    DROP CONSTRAINT IF EXISTS fk_transactions_transfer_account;
ALTER TABLE public.transactions    DROP CONSTRAINT IF EXISTS fk_transactions_account;

ALTER TABLE public.attachments     DROP CONSTRAINT IF EXISTS fk_attachments_transaction;
ALTER TABLE public.audit_logs      DROP CONSTRAINT IF EXISTS fk_audit_logs_organization;
ALTER TABLE public.audit_logs      DROP CONSTRAINT IF EXISTS fk_audit_logs_user;
ALTER TABLE public.notifications   DROP CONSTRAINT IF EXISTS fk_notifications_user;
ALTER TABLE public.recurrence_rules DROP CONSTRAINT IF EXISTS fk_recurrence_rules_contact;
ALTER TABLE public.recurrence_rules DROP CONSTRAINT IF EXISTS fk_recurrence_rules_category;
ALTER TABLE public.transactions    DROP CONSTRAINT IF EXISTS fk_transactions_credit_card;
ALTER TABLE public.transactions    DROP CONSTRAINT IF EXISTS fk_transactions_contact;
ALTER TABLE public.transactions    DROP CONSTRAINT IF EXISTS fk_transactions_category;
