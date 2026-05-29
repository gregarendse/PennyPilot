# PennyPilot — Project Prompt

## What this is
A self-hosted personal finance budget tracker. The goal is automatic bank sync across all accounts, a spreadsheet-style budget vs actual view, and a modern responsive UI. Inspired by Actual Budget but built from scratch with a better model.

## Stack
- **Backend**: Go (Chi router, golang-migrate, pgx)
- **Frontend**: Next.js (React, TypeScript, TanStack Table, Tailwind)
- **Database**: PostgreSQL — amounts stored as BIGINT pence, never floats
- **Deployment**: Docker Compose, self-hosted homelab

## Repo structure
```
/
├── backend/
│   ├── cmd/server/          # main entrypoint
│   ├── internal/
│   │   ├── api/             # HTTP handlers
│   │   ├── config/          # env config loader
│   │   ├── domain/          # shared types (Account, Transaction, Budget)
│   │   ├── sync/            # BankConnector interface + per-bank implementations
│   │   │   ├── monzo/
│   │   │   ├── truelayer/
│   │   │   └── csv/
│   │   └── budget/          # budget calculation logic
│   └── migrations/
├── frontend/
│   ├── app/                 # Next.js App Router
│   ├── components/
│   └── lib/
├── docker-compose.yml
└── .env.example
```

## Bank connectivity
| Provider | Method | Status |
|---|---|---|
| Monzo | Direct OAuth2 API (docs.monzo.com) | ✅ Connector written |
| Barclays current account | TrueLayer (FCA-regulated Open Banking) | 🔜 Phase 3 |
| Barclaycard | TrueLayer (same connection as Barclays) | 🔜 Phase 3 |
| American Express | CSV/OFX import — Amex does not participate in UK Open Banking | 🔜 Phase 3 |
| Future banks | Implement `BankConnector` interface | Plug-in |

**Monzo note**: Requires re-auth every 90 days by design. Handle with a clean re-auth prompt, not a workaround.

**Amex note**: No PSD2 compliance, no aggregator supports it. CSV import is the correct solution — build it well.

## Core connector interface (Go)
```go
type BankConnector interface {
    Name() string
    AuthURL(state string) string
    Exchange(ctx context.Context, code string) (*Credentials, error)
    FetchTransactions(ctx context.Context, creds *Credentials, accountID string, since time.Time) ([]domain.Transaction, error)
    FetchAccounts(ctx context.Context, creds *Credentials) ([]domain.Account, error)
    RefreshCredentials(ctx context.Context, creds *Credentials) (*Credentials, error)
}
```
Adding a new bank = a new struct implementing this interface.

## Key database tables
- `accounts` — one row per connected bank account
- `provider_credentials` — OAuth tokens, AES-encrypted at rest
- `transactions` — all transactions; `UNIQUE(account_id, external_id)` prevents duplicates
- `categories` — hierarchical, with colour + icon
- `category_rules` — rule-based auto-categorisation (contains / starts_with / regex)
- `budgets` — monthly amount per category (`UNIQUE(category_id, month)`)
- `sync_log` — history of sync runs per account

## Build phases
1. **Foundation** ✅ — Go backend scaffold, DB migrations, Monzo connector, basic REST API
2. **Core UI** — Transaction list (TanStack Table), category assignment, budget setup, budget vs actual spreadsheet view
3. **More sync** — TrueLayer (Barclays + Barclaycard), CSV importer, background sync scheduler
4. **Polish** — Dashboard/charts, transfer detection, over-budget alerts, mobile optimisation

## Non-negotiable design rules
- Amounts always in pence (BIGINT), never floats
- Deduplication via `UNIQUE(account_id, external_id)` — syncs are idempotent
- Ingest first, UI second — don't build views around mock data
- Every bank connector is a plug-in; no bank-specific logic in the core API
- Encrypted tokens at rest (AES-256)

## Current state
Phase 1 is scaffolded:
- Go backend compiles and runs
- DB migrations applied (PostgreSQL via Docker Compose)
- Monzo OAuth2 flow and transaction fetcher implemented
- REST endpoints: `GET /api/accounts`, `GET /api/transactions`, `GET /api/categories`, `POST /api/accounts/{id}/sync`
- Auth routes: `GET /auth/monzo`, `GET /auth/monzo/callback`
- Credential storage in DB not yet wired (tokens returned from callback but not persisted)

## Immediate next tasks
1. Persist Monzo credentials (encrypted) to `provider_credentials` after OAuth callback
2. Store account record in `accounts` table on first connect
3. Implement `POST /api/accounts/{id}/sync` — load credentials from DB, call Monzo, upsert transactions
4. Frontend scaffold — Next.js project, connect to API, render transaction list