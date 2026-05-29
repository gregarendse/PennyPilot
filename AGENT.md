# PennyPilot

## Goal
Self-hosted personal finance tracker. Automatic bank sync, spreadsheet-style budget vs actual view, modern responsive UI. Single user. Deployed on homelab via Docker Compose.

## Stack
| Layer | Technology |
|---|---|
| Backend | Go — Chi router, golang-migrate, pgx |
| Frontend | Next.js (App Router), TypeScript, Tailwind, TanStack Table |
| Database | PostgreSQL |
| Deployment | Docker Compose |

## Repo Layout
```
/
├── backend/
│   ├── cmd/server/
│   └── internal/
│       ├── api/
│       ├── config/
│       ├── domain/
│       ├── sync/            # connector interface + per-bank impls
│       │   ├── monzo/
│       │   ├── truelayer/
│       │   └── csv/
│       └── budget/
│   └── migrations/
├── frontend/
│   ├── app/
│   ├── components/
│   └── lib/
├── docker-compose.yml
└── .env.example
```

## Bank Connectivity
| Bank | Method | Notes |
|---|---|---|
| Monzo | Direct OAuth2 — docs.monzo.com | Re-auth required every 90 days by design |
| Barclays (current) | TrueLayer (Open Banking) | |
| Barclaycard | TrueLayer (same connection) | |
| American Express | CSV/OFX import only | Not PSD2-compliant; no aggregator supports it |
| Additional banks | Implement `BankConnector` interface | |

## Core Interfaces

```go
// Every bank implements this
type BankConnector interface {
    Name() string
    AuthURL(state string) string
    Exchange(ctx context.Context, code string) (*Credentials, error)
    FetchTransactions(ctx context.Context, creds *Credentials, accountID string, since time.Time) ([]domain.Transaction, error)
    FetchAccounts(ctx context.Context, creds *Credentials) ([]domain.Account, error)
    RefreshCredentials(ctx context.Context, creds *Credentials) (*Credentials, error)
}
```

## Database Tables
- `accounts` — one row per connected bank account
- `provider_credentials` — OAuth tokens, AES-256 encrypted at rest
- `transactions` — `UNIQUE(account_id, external_id)` enforces deduplication
- `categories` — hierarchical, colour + icon
- `category_rules` — auto-categorisation: contains / starts_with / regex on description or merchant_name
- `budgets` — monthly amount per category, `UNIQUE(category_id, month)`
- `sync_log` — sync run history per account

## Rules
- Amounts always stored as BIGINT pence — never floats
- Syncs must be idempotent — re-fetching a transaction must not create a duplicate
- OAuth tokens encrypted before writing to DB; never logged
- No bank-specific logic outside of `internal/sync/<bank>/`
- Monzo 90-day re-auth: surface a prompt to the user, do not attempt to work around it

## Build Phases
1. **Foundation** — Go backend, DB migrations, Monzo connector, basic REST API
2. **Core UI** — transaction list, category assignment, budget setup, budget vs actual view
3. **More sync** — TrueLayer (Barclays + Barclaycard), CSV importer, background sync scheduler
4. **Polish** — dashboard/charts, transfer detection, alerts, mobile optimisation