-- Integridade referencial. Até aqui o schema não tinha foreign keys (exceto na
-- join table transaction_tags): apagar uma conta deixava seus lançamentos
-- apontando para um UUID inexistente, e nada no banco impedia isso.
--
-- POLÍTICA DE ON DELETE
--   organization_id  -> CASCADE    a organização é a raiz do agregado; quando a
--                                  linha é finalmente removida (purga de
--                                  retenção), os dados dela vão junto.
--   user_id (sessão) -> CASCADE    refresh tokens, resets, códigos 2FA: artefatos
--                                  do usuário, não sobrevivem a ele.
--   user_id (registro) -> SET NULL notificações e auditoria preservam o registro
--                                  e perdem apenas o vínculo.
--   account_id       -> NO ACTION  um lançamento SEM conta é um razão quebrado.
--   category/contact/card -> SET NULL  são opcionais no lançamento (colunas
--                                  anuláveis): perder o vínculo não corrompe o valor.
--
-- Por que NO ACTION e não RESTRICT em account_id: RESTRICT é checado
-- imediatamente, então apagar uma organização falharia — o CASCADE removeria
-- accounts antes de transactions e o RESTRICT dispararia no meio do caminho.
-- NO ACTION adia a checagem para o fim do statement, dando a mesma proteção
-- contra apagar uma conta com lançamentos, sem quebrar o cascade da organização.

-- ---------------------------------------------------------------------------
-- 1. Reparo: referências opcionais que já apontam para linhas inexistentes.
--    Só colunas anuláveis são reparáveis desta forma (vira NULL, que é o mesmo
--    estado que a API já apresenta hoje: "sem categoria", "sem contato").
-- ---------------------------------------------------------------------------
UPDATE public.transactions t SET category_id = NULL
 WHERE category_id IS NOT NULL
   AND NOT EXISTS (SELECT 1 FROM public.categories c WHERE c.id = t.category_id);

UPDATE public.transactions t SET contact_id = NULL
 WHERE contact_id IS NOT NULL
   AND NOT EXISTS (SELECT 1 FROM public.contacts c WHERE c.id = t.contact_id);

UPDATE public.transactions t SET credit_card_id = NULL
 WHERE credit_card_id IS NOT NULL
   AND NOT EXISTS (SELECT 1 FROM public.credit_cards c WHERE c.id = t.credit_card_id);

UPDATE public.recurrence_rules r SET category_id = NULL
 WHERE category_id IS NOT NULL
   AND NOT EXISTS (SELECT 1 FROM public.categories c WHERE c.id = r.category_id);

UPDATE public.recurrence_rules r SET contact_id = NULL
 WHERE contact_id IS NOT NULL
   AND NOT EXISTS (SELECT 1 FROM public.contacts c WHERE c.id = r.contact_id);

UPDATE public.notifications n SET user_id = NULL
 WHERE user_id IS NOT NULL
   AND NOT EXISTS (SELECT 1 FROM public.users u WHERE u.id = n.user_id);

UPDATE public.audit_logs a SET user_id = NULL
 WHERE user_id IS NOT NULL
   AND NOT EXISTS (SELECT 1 FROM public.users u WHERE u.id = a.user_id);

UPDATE public.audit_logs a SET organization_id = NULL
 WHERE organization_id IS NOT NULL
   AND NOT EXISTS (SELECT 1 FROM public.organizations o WHERE o.id = a.organization_id);

UPDATE public.attachments a SET transaction_id = NULL
 WHERE transaction_id IS NOT NULL
   AND NOT EXISTS (SELECT 1 FROM public.transactions t WHERE t.id = a.transaction_id);

-- ---------------------------------------------------------------------------
-- 2. Constraints sobre colunas ANULÁVEIS — já reparadas acima, entram validadas.
-- ---------------------------------------------------------------------------
ALTER TABLE public.transactions
  ADD CONSTRAINT fk_transactions_category
  FOREIGN KEY (category_id) REFERENCES public.categories(id) ON DELETE SET NULL;

ALTER TABLE public.transactions
  ADD CONSTRAINT fk_transactions_contact
  FOREIGN KEY (contact_id) REFERENCES public.contacts(id) ON DELETE SET NULL;

ALTER TABLE public.transactions
  ADD CONSTRAINT fk_transactions_credit_card
  FOREIGN KEY (credit_card_id) REFERENCES public.credit_cards(id) ON DELETE SET NULL;

ALTER TABLE public.recurrence_rules
  ADD CONSTRAINT fk_recurrence_rules_category
  FOREIGN KEY (category_id) REFERENCES public.categories(id) ON DELETE SET NULL;

ALTER TABLE public.recurrence_rules
  ADD CONSTRAINT fk_recurrence_rules_contact
  FOREIGN KEY (contact_id) REFERENCES public.contacts(id) ON DELETE SET NULL;

ALTER TABLE public.notifications
  ADD CONSTRAINT fk_notifications_user
  FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE public.audit_logs
  ADD CONSTRAINT fk_audit_logs_user
  FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE public.audit_logs
  ADD CONSTRAINT fk_audit_logs_organization
  FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE SET NULL;

ALTER TABLE public.attachments
  ADD CONSTRAINT fk_attachments_transaction
  FOREIGN KEY (transaction_id) REFERENCES public.transactions(id) ON DELETE CASCADE;

-- ---------------------------------------------------------------------------
-- 3. Constraints sobre colunas NOT NULL — entram como NOT VALID.
--
--    NOT VALID pula APENAS a verificação das linhas já existentes; inserções,
--    updates e as ações de ON DELETE passam a ser enforçadas normalmente. Isso
--    é deliberado: uma instalação antiga pode já ter órfãos reais (o usuário
--    apagou uma conta com lançamentos e, 90 dias depois, a purga de retenção
--    removeu a linha da conta de vez). Validar aqui faria o app não subir.
--
--    Depois de limpar os órfãos legados, o operador valida com:
--      ALTER TABLE public.transactions VALIDATE CONSTRAINT fk_transactions_account;
--    (uma por constraint; instalações novas nascem sem órfão nenhum).
-- ---------------------------------------------------------------------------
ALTER TABLE public.transactions
  ADD CONSTRAINT fk_transactions_account
  FOREIGN KEY (account_id) REFERENCES public.accounts(id) ON DELETE NO ACTION NOT VALID;

ALTER TABLE public.transactions
  ADD CONSTRAINT fk_transactions_transfer_account
  FOREIGN KEY (transfer_account_id) REFERENCES public.accounts(id) ON DELETE NO ACTION NOT VALID;

ALTER TABLE public.recurrence_rules
  ADD CONSTRAINT fk_recurrence_rules_account
  FOREIGN KEY (account_id) REFERENCES public.accounts(id) ON DELETE NO ACTION NOT VALID;

ALTER TABLE public.budgets
  ADD CONSTRAINT fk_budgets_category
  FOREIGN KEY (category_id) REFERENCES public.categories(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.category_rules
  ADD CONSTRAINT fk_category_rules_category
  FOREIGN KEY (category_id) REFERENCES public.categories(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.organizations
  ADD CONSTRAINT fk_organizations_owner
  FOREIGN KEY (owner_id) REFERENCES public.users(id) ON DELETE NO ACTION NOT VALID;

ALTER TABLE public.memberships
  ADD CONSTRAINT fk_memberships_user
  FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.refresh_tokens
  ADD CONSTRAINT fk_refresh_tokens_user
  FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.password_resets
  ADD CONSTRAINT fk_password_resets_user
  FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.recovery_codes
  ADD CONSTRAINT fk_recovery_codes_user
  FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.email_verifications
  ADD CONSTRAINT fk_email_verifications_user
  FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE NOT VALID;

-- ---------------------------------------------------------------------------
-- 4. organization_id em todas as tabelas org-scoped -> CASCADE.
-- ---------------------------------------------------------------------------
ALTER TABLE public.accounts
  ADD CONSTRAINT fk_accounts_organization
  FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.categories
  ADD CONSTRAINT fk_categories_organization
  FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.contacts
  ADD CONSTRAINT fk_contacts_organization
  FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.credit_cards
  ADD CONSTRAINT fk_credit_cards_organization
  FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.transactions
  ADD CONSTRAINT fk_transactions_organization
  FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.budgets
  ADD CONSTRAINT fk_budgets_organization
  FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.goals
  ADD CONSTRAINT fk_goals_organization
  FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.tags
  ADD CONSTRAINT fk_tags_organization
  FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.category_rules
  ADD CONSTRAINT fk_category_rules_organization
  FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.recurrence_rules
  ADD CONSTRAINT fk_recurrence_rules_organization
  FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.attachments
  ADD CONSTRAINT fk_attachments_organization
  FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.memberships
  ADD CONSTRAINT fk_memberships_organization
  FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.invitations
  ADD CONSTRAINT fk_invitations_organization
  FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.notifications
  ADD CONSTRAINT fk_notifications_organization
  FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE NOT VALID;

-- ---------------------------------------------------------------------------
-- 5. Índice que faltava: a listagem de lançamentos filtra por organização e
--    ordena por data. Sem ele o plano é seq scan + sort, que degrada conforme
--    o histórico cresce.
-- ---------------------------------------------------------------------------
CREATE INDEX IF NOT EXISTS idx_transactions_org_date
  ON public.transactions USING btree (organization_id, date DESC)
  WHERE deleted_at IS NULL;
