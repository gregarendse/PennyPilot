CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider TEXT NOT NULL,
    external_id TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'GBP',
    balance_pence BIGINT NOT NULL DEFAULT 0,
    last_synced_at TIMESTAMPTZ,
    reconnect_after TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, external_id)
);

CREATE TABLE provider_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider TEXT NOT NULL,
    account_id UUID REFERENCES accounts(id) ON DELETE CASCADE,
    encrypted_payload BYTEA NOT NULL,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, account_id)
);

CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    color TEXT NOT NULL DEFAULT '#64748b',
    icon TEXT NOT NULL DEFAULT 'tag',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (parent_id, name)
);

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    external_id TEXT NOT NULL,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    amount_pence BIGINT NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'GBP',
    description TEXT NOT NULL,
    merchant_name TEXT,
    occurred_at TIMESTAMPTZ NOT NULL,
    pending BOOLEAN NOT NULL DEFAULT false,
    notes TEXT,
    imported_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (account_id, external_id)
);

CREATE TABLE category_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    field TEXT NOT NULL CHECK (field IN ('description', 'merchant_name')),
    operator TEXT NOT NULL CHECK (operator IN ('contains', 'starts_with', 'regex')),
    pattern TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 100,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    month DATE NOT NULL,
    amount_pence BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (category_id, month)
);

CREATE TABLE sync_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
    provider TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('started', 'succeeded', 'failed')),
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ,
    transactions_seen INTEGER NOT NULL DEFAULT 0,
    transactions_imported INTEGER NOT NULL DEFAULT 0,
    error_message TEXT
);

CREATE INDEX transactions_account_occurred_at_idx ON transactions (account_id, occurred_at DESC);
CREATE INDEX transactions_category_occurred_at_idx ON transactions (category_id, occurred_at DESC);
CREATE INDEX category_rules_priority_idx ON category_rules (enabled, priority);
CREATE INDEX sync_log_account_started_at_idx ON sync_log (account_id, started_at DESC);
