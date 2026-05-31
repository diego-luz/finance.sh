--
-- PostgreSQL database dump
--


-- Dumped from database version 16.14
-- Dumped by pg_dump version 16.14

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: accounts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.accounts (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    organization_id uuid NOT NULL,
    name text NOT NULL,
    type character varying(20) DEFAULT 'bank'::character varying,
    initial_balance bigint DEFAULT 0,
    color character varying(9) DEFAULT '#10b981'::character varying,
    icon text DEFAULT 'wallet'::text,
    archived boolean DEFAULT false
);


--
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audit_logs (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    organization_id uuid,
    user_id uuid,
    action text NOT NULL,
    entity text,
    entity_id text,
    ip text,
    metadata jsonb
);


--
-- Name: budgets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.budgets (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    organization_id uuid NOT NULL,
    category_id uuid NOT NULL,
    amount bigint NOT NULL,
    month bigint NOT NULL,
    year bigint NOT NULL
);


--
-- Name: categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.categories (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    organization_id uuid NOT NULL,
    name text NOT NULL,
    kind character varying(10) DEFAULT 'expense'::character varying,
    color character varying(9) DEFAULT '#6366f1'::character varying,
    icon text DEFAULT 'tag'::text
);


--
-- Name: credit_cards; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.credit_cards (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    organization_id uuid NOT NULL,
    name text NOT NULL,
    "limit" bigint DEFAULT 0,
    closing_day bigint DEFAULT 1,
    due_day bigint DEFAULT 10,
    color character varying(9) DEFAULT '#0f1115'::character varying
);


--
-- Name: email_verifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.email_verifications (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id uuid NOT NULL,
    token_hash text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    used boolean DEFAULT false
);


--
-- Name: goals; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.goals (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    organization_id uuid NOT NULL,
    name text NOT NULL,
    target_amount bigint NOT NULL,
    current_amount bigint DEFAULT 0,
    deadline timestamp with time zone,
    color character varying(9) DEFAULT '#10b981'::character varying
);


--
-- Name: invitations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invitations (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    organization_id uuid NOT NULL,
    email text NOT NULL,
    role character varying(20) DEFAULT 'member'::character varying,
    token text NOT NULL,
    accepted boolean DEFAULT false
);


--
-- Name: memberships; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.memberships (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id uuid NOT NULL,
    organization_id uuid NOT NULL,
    role character varying(20) DEFAULT 'member'::character varying
);


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    organization_id uuid NOT NULL,
    user_id uuid,
    type character varying(30),
    title text NOT NULL,
    message text,
    read boolean DEFAULT false
);


--
-- Name: organizations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.organizations (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    slug text NOT NULL,
    owner_id uuid NOT NULL,
    currency character varying(3) DEFAULT 'BRL'::character varying
);


--
-- Name: password_resets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.password_resets (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id uuid NOT NULL,
    token_hash text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    used boolean DEFAULT false
);


--
-- Name: recovery_codes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.recovery_codes (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id uuid NOT NULL,
    code_hash text NOT NULL,
    used boolean DEFAULT false
);


--
-- Name: refresh_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.refresh_tokens (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id uuid NOT NULL,
    token_hash text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    revoked boolean DEFAULT false,
    user_agent text,
    ip text
);


--
-- Name: subscriptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.subscriptions (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    organization_id uuid NOT NULL,
    plan character varying(20) DEFAULT 'free'::character varying,
    status character varying(20) DEFAULT 'trialing'::character varying,
    trial_ends_at timestamp with time zone,
    current_period_end timestamp with time zone,
    max_users bigint DEFAULT 1,
    max_transactions bigint DEFAULT 500,
    provider character varying(20),
    external_id text
);


--
-- Name: transactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.transactions (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    organization_id uuid NOT NULL,
    account_id uuid NOT NULL,
    category_id uuid,
    credit_card_id uuid,
    transfer_account_id uuid,
    type character varying(10) NOT NULL,
    amount bigint NOT NULL,
    description text NOT NULL,
    date timestamp with time zone NOT NULL,
    paid boolean NOT NULL,
    recurring boolean DEFAULT false,
    notes text
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    email text NOT NULL,
    password_hash text NOT NULL,
    email_verified boolean DEFAULT false,
    avatar_url text,
    terms_accepted_at timestamp with time zone,
    terms_version character varying(20),
    two_factor_enabled boolean DEFAULT false,
    two_factor_secret text
);


--
-- Name: accounts accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (id);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: budgets budgets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.budgets
    ADD CONSTRAINT budgets_pkey PRIMARY KEY (id);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- Name: credit_cards credit_cards_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.credit_cards
    ADD CONSTRAINT credit_cards_pkey PRIMARY KEY (id);


--
-- Name: email_verifications email_verifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_verifications
    ADD CONSTRAINT email_verifications_pkey PRIMARY KEY (id);


--
-- Name: goals goals_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.goals
    ADD CONSTRAINT goals_pkey PRIMARY KEY (id);


--
-- Name: invitations invitations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invitations
    ADD CONSTRAINT invitations_pkey PRIMARY KEY (id);


--
-- Name: memberships memberships_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.memberships
    ADD CONSTRAINT memberships_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: organizations organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_pkey PRIMARY KEY (id);


--
-- Name: password_resets password_resets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.password_resets
    ADD CONSTRAINT password_resets_pkey PRIMARY KEY (id);


--
-- Name: recovery_codes recovery_codes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recovery_codes
    ADD CONSTRAINT recovery_codes_pkey PRIMARY KEY (id);


--
-- Name: refresh_tokens refresh_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_pkey PRIMARY KEY (id);


--
-- Name: subscriptions subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_pkey PRIMARY KEY (id);


--
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_accounts_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_accounts_deleted_at ON public.accounts USING btree (deleted_at);


--
-- Name: idx_accounts_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_accounts_organization_id ON public.accounts USING btree (organization_id);


--
-- Name: idx_audit_logs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_deleted_at ON public.audit_logs USING btree (deleted_at);


--
-- Name: idx_audit_logs_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_organization_id ON public.audit_logs USING btree (organization_id);


--
-- Name: idx_audit_logs_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_user_id ON public.audit_logs USING btree (user_id);


--
-- Name: idx_budgets_category_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_budgets_category_id ON public.budgets USING btree (category_id);


--
-- Name: idx_budgets_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_budgets_deleted_at ON public.budgets USING btree (deleted_at);


--
-- Name: idx_budgets_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_budgets_organization_id ON public.budgets USING btree (organization_id);


--
-- Name: idx_categories_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_categories_deleted_at ON public.categories USING btree (deleted_at);


--
-- Name: idx_categories_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_categories_organization_id ON public.categories USING btree (organization_id);


--
-- Name: idx_credit_cards_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_credit_cards_deleted_at ON public.credit_cards USING btree (deleted_at);


--
-- Name: idx_credit_cards_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_credit_cards_organization_id ON public.credit_cards USING btree (organization_id);


--
-- Name: idx_email_verifications_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_email_verifications_deleted_at ON public.email_verifications USING btree (deleted_at);


--
-- Name: idx_email_verifications_token_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_email_verifications_token_hash ON public.email_verifications USING btree (token_hash);


--
-- Name: idx_email_verifications_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_email_verifications_user_id ON public.email_verifications USING btree (user_id);


--
-- Name: idx_goals_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_goals_deleted_at ON public.goals USING btree (deleted_at);


--
-- Name: idx_goals_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_goals_organization_id ON public.goals USING btree (organization_id);


--
-- Name: idx_invitations_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_invitations_deleted_at ON public.invitations USING btree (deleted_at);


--
-- Name: idx_invitations_email; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_invitations_email ON public.invitations USING btree (email);


--
-- Name: idx_invitations_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_invitations_organization_id ON public.invitations USING btree (organization_id);


--
-- Name: idx_invitations_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_invitations_token ON public.invitations USING btree (token);


--
-- Name: idx_memberships_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_memberships_deleted_at ON public.memberships USING btree (deleted_at);


--
-- Name: idx_notifications_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_deleted_at ON public.notifications USING btree (deleted_at);


--
-- Name: idx_notifications_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_organization_id ON public.notifications USING btree (organization_id);


--
-- Name: idx_notifications_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_user_id ON public.notifications USING btree (user_id);


--
-- Name: idx_organizations_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_organizations_deleted_at ON public.organizations USING btree (deleted_at);


--
-- Name: idx_organizations_owner_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_organizations_owner_id ON public.organizations USING btree (owner_id);


--
-- Name: idx_organizations_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_organizations_slug ON public.organizations USING btree (slug);


--
-- Name: idx_password_resets_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_password_resets_deleted_at ON public.password_resets USING btree (deleted_at);


--
-- Name: idx_password_resets_token_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_password_resets_token_hash ON public.password_resets USING btree (token_hash);


--
-- Name: idx_password_resets_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_password_resets_user_id ON public.password_resets USING btree (user_id);


--
-- Name: idx_recovery_codes_code_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_recovery_codes_code_hash ON public.recovery_codes USING btree (code_hash);


--
-- Name: idx_recovery_codes_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_recovery_codes_deleted_at ON public.recovery_codes USING btree (deleted_at);


--
-- Name: idx_recovery_codes_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_recovery_codes_user_id ON public.recovery_codes USING btree (user_id);


--
-- Name: idx_refresh_tokens_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_refresh_tokens_deleted_at ON public.refresh_tokens USING btree (deleted_at);


--
-- Name: idx_refresh_tokens_token_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_refresh_tokens_token_hash ON public.refresh_tokens USING btree (token_hash);


--
-- Name: idx_refresh_tokens_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_refresh_tokens_user_id ON public.refresh_tokens USING btree (user_id);


--
-- Name: idx_subscriptions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_subscriptions_deleted_at ON public.subscriptions USING btree (deleted_at);


--
-- Name: idx_subscriptions_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_subscriptions_organization_id ON public.subscriptions USING btree (organization_id);


--
-- Name: idx_transactions_account_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transactions_account_id ON public.transactions USING btree (account_id);


--
-- Name: idx_transactions_category_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transactions_category_id ON public.transactions USING btree (category_id);


--
-- Name: idx_transactions_credit_card_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transactions_credit_card_id ON public.transactions USING btree (credit_card_id);


--
-- Name: idx_transactions_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transactions_date ON public.transactions USING btree (date);


--
-- Name: idx_transactions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transactions_deleted_at ON public.transactions USING btree (deleted_at);


--
-- Name: idx_transactions_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transactions_organization_id ON public.transactions USING btree (organization_id);


--
-- Name: idx_transactions_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transactions_type ON public.transactions USING btree (type);


--
-- Name: idx_user_org; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_user_org ON public.memberships USING btree (user_id, organization_id);


--
-- Name: idx_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_email ON public.users USING btree (email);


--
-- PostgreSQL database dump complete
--


