CREATE TABLE accounts (
    id TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    external_id TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'GBP',
    balance_pence BIGINT NOT NULL DEFAULT 0,
    last_synced_at TEXT,
    reconnect_after TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE (provider, external_id)
);

CREATE TABLE provider_credentials (
    id TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    account_id TEXT REFERENCES accounts(id) ON DELETE CASCADE,
    encrypted_payload BLOB NOT NULL,
    expires_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE (provider, account_id)
);

CREATE TABLE categories (
    id TEXT PRIMARY KEY,
    parent_id TEXT REFERENCES categories(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    color TEXT NOT NULL DEFAULT '#64748b',
    icon TEXT NOT NULL DEFAULT 'tag',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE (parent_id, name)
);

CREATE TABLE transactions (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    external_id TEXT NOT NULL,
    category_id TEXT REFERENCES categories(id) ON DELETE SET NULL,
    amount_pence BIGINT NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'GBP',
    description TEXT NOT NULL,
    merchant_name TEXT,
    occurred_at TEXT NOT NULL,
    pending BOOLEAN NOT NULL DEFAULT 0,
    notes TEXT,
    imported_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE (account_id, external_id)
);

CREATE TABLE category_rules (
    id TEXT PRIMARY KEY,
    category_id TEXT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    field TEXT NOT NULL CHECK (field IN ('description', 'merchant_name')),
    operator TEXT NOT NULL CHECK (operator IN ('contains', 'starts_with', 'regex')),
    pattern TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 100,
    enabled BOOLEAN NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE budgets (
    id TEXT PRIMARY KEY,
    category_id TEXT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    month TEXT NOT NULL,
    amount_pence BIGINT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE (category_id, month)
);

CREATE TABLE sync_log (
    id TEXT PRIMARY KEY,
    account_id TEXT REFERENCES accounts(id) ON DELETE SET NULL,
    provider TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('started', 'succeeded', 'failed')),
    started_at TEXT NOT NULL DEFAULT (datetime('now')),
    finished_at TEXT,
    transactions_seen INTEGER NOT NULL DEFAULT 0,
    transactions_imported INTEGER NOT NULL DEFAULT 0,
    error_message TEXT
);

CREATE INDEX transactions_account_occurred_at_idx ON transactions (account_id, occurred_at);
CREATE INDEX transactions_category_occurred_at_idx ON transactions (category_id, occurred_at);
CREATE INDEX category_rules_priority_idx ON category_rules (enabled, priority);
CREATE INDEX sync_log_account_started_at_idx ON sync_log (account_id, started_at);

-- Triggers to update updated_at
CREATE TRIGGER update_accounts_updated_at AFTER UPDATE ON accounts
BEGIN
    UPDATE accounts SET updated_at = datetime('now') WHERE id = OLD.id;
END;

CREATE TRIGGER update_provider_credentials_updated_at AFTER UPDATE ON provider_credentials
BEGIN
    UPDATE provider_credentials SET updated_at = datetime('now') WHERE id = OLD.id;
END;

CREATE TRIGGER update_categories_updated_at AFTER UPDATE ON categories
BEGIN
    UPDATE categories SET updated_at = datetime('now') WHERE id = OLD.id;
END;

CREATE TRIGGER update_transactions_updated_at AFTER UPDATE ON transactions
BEGIN
    UPDATE transactions SET updated_at = datetime('now') WHERE id = OLD.id;
END;

CREATE TRIGGER update_budgets_updated_at AFTER UPDATE ON budgets
BEGIN
    UPDATE budgets SET updated_at = datetime('now') WHERE id = OLD.id;
END;
