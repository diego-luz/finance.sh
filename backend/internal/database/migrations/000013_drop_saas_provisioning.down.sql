-- Rollback: restore the organizations.suspended flag and the dropped tables
-- (minimal shapes; the original column/index detail is not fully reproduced).
ALTER TABLE public.organizations ADD COLUMN IF NOT EXISTS suspended boolean NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS public.subscriptions (
    id uuid PRIMARY KEY,
    created_at timestamptz,
    updated_at timestamptz,
    deleted_at timestamptz,
    organization_id uuid NOT NULL,
    plan varchar(20) DEFAULT 'community',
    status varchar(20) DEFAULT 'active',
    trial_ends_at timestamptz,
    current_period_end timestamptz,
    max_users integer DEFAULT 0,
    max_transactions integer DEFAULT 0,
    provider varchar(20),
    external_id text
);

CREATE TABLE IF NOT EXISTS public.leads (
    id uuid PRIMARY KEY,
    created_at timestamptz,
    updated_at timestamptz,
    deleted_at timestamptz,
    name text NOT NULL,
    email text NOT NULL,
    phone varchar(30),
    organization_name text,
    profile_type varchar(20),
    message text,
    status varchar(20) DEFAULT 'new',
    notes text,
    converted_org_id uuid,
    ip text,
    user_agent text
);
