# Personal Finance Budget Tracker — Architecture & Build Plan

## What We Learned from Previous Attempts

| Repo | Stack | What broke |
|---|---|---|
| `emolument` | Angular + Firebase | No data layer — UI without bank sync is just a manual spreadsheet |
| `actual-budget-cli` | TypeScript CLI | Smart: used Actual's API as middleware, but still tied to Actual's model and sync |
| `PennyPilot` | (empty) | Never started |
| `monzo-task-review` | Go (private) | Direct Monzo API exploration — this is the right instinct |

**Root cause of failure**: Every attempt either skipped bank sync entirely or tried to bolt it on. Bank sync is the hard part and must be the **foundation**, not an afterthought.

---

## The Bank Connectivity Problem (UK)

This is the thorniest part. Here's the honest reality for each provider:

### Monzo ✅ Easy — direct public API
Monzo has a first-class OAuth 2.0 API. You can get read access to accounts, balances, and transactions with a personal client. Free, well-documented, reliable.
- Docs: https://docs.monzo.com
- Auth: OAuth2 PKCE flow
- Refresh: 90-day re-auth required (Monzo security policy — you can automate the reminder)

### Barclays (current account + credit card) ✅ Possible via Open Banking
Barclays is PSD2-compliant and participates in UK Open Banking. The cleanest approach is **TrueLayer**, which is FCA-regulated and supports Barclays (current account) and Barclaycard.

> ⚠️ GoCardless Bank Account Data (formerly Nordigen) — was the obvious free choice but they've **sunset the free tier** for new registrations as of late 2024/2025. Firefly III still references it but new signups are closed.

TrueLayer has a free sandbox and charges for production use. Reasonable for personal use if you self-host, but requires registration.

**Alternative**: Plaid also covers Barclays and many UK banks, similar pricing model.

### American Express ⚠️ Hard — not UK Open Banking compliant
This is the ugly truth: **American Express does not participate in UK Open Banking / PSD2**. They are not an ASPSP (Account Servicing Payment Service Provider) under UK regs. Neither TrueLayer, GoCardless, nor any aggregator can pull Amex UK transactions via Open Banking.

Options for Amex:
1. **CSV import** (always available — Amex lets you export OFX/CSV). Build a smart CSV importer.
2. **Plaid** — Plaid supports Amex in the US; UK coverage is limited/inconsistent.
3. **Screen scraping** — fragile, against ToS, not recommended.
4. **Amex Developer API** — exists but is B2B focused, not for personal account data.

**Recommended pragmatic approach**: Start with CSV import for Amex with a clean one-click upload UI. Most Amex users check their statement monthly anyway. This is better than a broken sync.

---

## Recommended Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      Frontend (Next.js)                  │
│  • Spreadsheet budget view     • Transaction list       │
│  • Dashboard / charts          • Category editor        │
│  • Bank connection setup       • CSV import             │
└────────────────────────┬────────────────────────────────┘
                         │ REST / GraphQL
┌────────────────────────▼────────────────────────────────┐
│                    Backend API (Go)                      │
│  • Transaction ingestion & deduplication                │
│  • Category rule engine                                  │
│  • Budget management                                     │
│  • Sync scheduler (cron)                                │
└──────┬──────────────────┬────────────────────┬──────────┘
       │                  │                    │
┌──────▼──────┐  ┌────────▼───────┐  ┌────────▼──────────┐
│  Monzo API  │  │  TrueLayer API │  │  CSV Import       │
│  (direct)   │  │  (Barclays,    │  │  (Amex, any bank) │
│  OAuth2     │  │   Barclaycard) │  │  OFX / CSV / QIF  │
└──────┬──────┘  └────────┬───────┘  └────────┬──────────┘
       │                  │                    │
┌──────▼──────────────────▼────────────────────▼──────────┐
│                   PostgreSQL Database                    │
│  accounts │ transactions │ categories │ budgets         │
└─────────────────────────────────────────────────────────┘
```

**Deployment**: Docker Compose on your homelab. Fits your existing infrastructure.

---

## Stack Decisions

### Backend: Go
You already have `BountyBeacon` and `monzo-task-review` in Go. It's the right fit:
- Excellent HTTP/OAuth libraries
- Single binary, easy Docker
- Great for cron-style sync jobs
- Chi or Fiber for REST API

### Frontend: Next.js (React + TypeScript)
You know TypeScript well. Next.js gives you:
- SSR for fast first load (good for mobile)
- API routes if you want to colocate some backend logic
- Excellent ecosystem for tables (TanStack Table for the spreadsheet view)
- Easy to deploy as a static export or Node server

### Database: PostgreSQL
Self-hosted, your homelab already likely runs it. Schema is simple.

### Bank Sync:
| Provider | Method | Cost | Notes |
|---|---|---|---|
| Monzo | Direct API | Free | First-class API, best option |
| Barclays current | TrueLayer | Freemium | Register for free dev tier |
| Barclaycard | TrueLayer | Freemium | Same TrueLayer connection |
| American Express | CSV import | Free | Build a slick upload UX |
| Future banks | TrueLayer / plug-in | Varies | Modular connector design |

---

## Database Schema

```sql
-- Accounts (synced banks + manual accounts)
CREATE TABLE accounts (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    type        TEXT NOT NULL,  -- 'current', 'credit', 'savings'
    provider    TEXT NOT NULL,  -- 'monzo', 'barclays', 'barclaycard', 'amex', 'manual'
    currency    TEXT DEFAULT 'GBP',
    is_active   BOOLEAN DEFAULT true,
    created_at  TIMESTAMPTZ DEFAULT now()
);

-- Provider credentials (encrypted at rest)
CREATE TABLE provider_credentials (
    id              UUID PRIMARY KEY,
    account_id      UUID REFERENCES accounts(id),
    provider        TEXT NOT NULL,
    access_token    TEXT,  -- encrypted
    refresh_token   TEXT,  -- encrypted
    expires_at      TIMESTAMPTZ,
    provider_data   JSONB  -- provider-specific metadata
);

-- Raw transactions from all sources
CREATE TABLE transactions (
    id              UUID PRIMARY KEY,
    account_id      UUID REFERENCES accounts(id),
    external_id     TEXT,  -- provider's transaction ID (for dedup)
    date            DATE NOT NULL,
    amount          BIGINT NOT NULL,  -- pence, always positive
    direction       TEXT NOT NULL,    -- 'debit' | 'credit'
    description     TEXT NOT NULL,
    merchant_name   TEXT,
    merchant_logo   TEXT,
    category_id     UUID REFERENCES categories(id),
    notes           TEXT,
    is_transfer     BOOLEAN DEFAULT false,
    raw_data        JSONB,  -- original provider payload
    created_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE(account_id, external_id)
);

-- Hierarchical categories
CREATE TABLE categories (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    parent_id   UUID REFERENCES categories(id),
    color       TEXT,
    icon        TEXT
);

-- Category assignment rules
CREATE TABLE category_rules (
    id          UUID PRIMARY KEY,
    category_id UUID REFERENCES categories(id),
    field       TEXT NOT NULL,   -- 'description', 'merchant_name'
    operator    TEXT NOT NULL,   -- 'contains', 'starts_with', 'regex'
    value       TEXT NOT NULL,
    priority    INT DEFAULT 0
);

-- Monthly budgets per category
CREATE TABLE budgets (
    id          UUID PRIMARY KEY,
    category_id UUID REFERENCES categories(id),
    month       DATE NOT NULL,   -- first day of month
    amount      BIGINT NOT NULL, -- pence
    UNIQUE(category_id, month)
);
```

---

## Build Phases

### Phase 1 — Foundation (Start Here)
1. **Go backend scaffold**: project layout, DB migrations (golang-migrate), config loading
2. **Monzo connector**: OAuth2 flow, transaction fetch, dedup logic
3. **Basic REST API**: accounts, transactions endpoints
4. **Next.js frontend scaffold**: project setup, Tailwind, basic routing

### Phase 2 — Core UI
1. **Transaction list**: sortable/filterable table (TanStack Table)
2. **Category assignment**: manual + rule-based auto-categorisation
3. **Budget setup UI**: set monthly amounts per category
4. **Budget vs Actual view**: the "spreadsheet" — monthly grid of budget/spent/remaining

### Phase 3 — More Bank Sync
1. **TrueLayer integration**: Barclays current account + Barclaycard
2. **CSV import**: Amex and any other bank — smart parser, column mapper
3. **Sync scheduler**: background cron in Go, configurable intervals

### Phase 4 — Polish & Features
1. **Dashboard**: spending trends, category breakdowns, monthly summary
2. **Transfer detection**: flag transactions between own accounts
3. **Notifications**: over-budget alerts (webhook / email)
4. **Mobile-optimised**: responsive layouts that feel native on phone

---

## Key Design Principles (Learning from What Failed Before)

1. **Ingest first, display second.** Don't build the UI until you have real data flowing in. The previous attempts built beautiful UIs around empty/mock data and stalled.

2. **Deduplication is non-negotiable.** Every sync will potentially re-fetch recent transactions. Build the `UNIQUE(account_id, external_id)` constraint from day one.

3. **Amount in pence (integers), never floats.** Float arithmetic on money causes bugs. Store everything as `BIGINT` pence.

4. **Don't fight Amex.** Accept CSV import is the answer. Build it well — drag-and-drop, auto-detect columns, preview before import. Most people check Amex monthly anyway.

5. **90-day Monzo re-auth is a feature, not a bug.** Build a clean re-auth flow with a reminder notification, not a hack to avoid it.

6. **Extensible connector model.** Each bank is a Go interface `BankConnector` with `Connect()`, `FetchTransactions()`, `RefreshToken()`. Adding a new bank is adding a new struct that implements the interface.

---

## Project Name

**Ledger** — or continue with PennyPilot if you like it. I'd go for something clean.

---

## Suggested Repo Structure

```
/
├── backend/           # Go API + sync engine
│   ├── cmd/
│   │   └── server/   # main entrypoint
│   ├── internal/
│   │   ├── api/      # HTTP handlers
│   │   ├── sync/     # bank connectors
│   │   │   ├── monzo/
│   │   │   ├── truelayer/
│   │   │   └── csv/
│   │   ├── db/       # queries (sqlc)
│   │   └── budget/   # budget calculation logic
│   └── migrations/
├── frontend/          # Next.js app
│   ├── app/          # App Router pages
│   ├── components/
│   └── lib/
├── docker-compose.yml
└── .env.example
```

---

## Immediate Next Step

Register for a **Monzo developer account** (monzo.com/developers) and a **TrueLayer sandbox** account (truelayer.com) to get your client IDs. These are free and instant. Then we scaffold the Go backend and get the first real transaction flowing in.

That's when it becomes real — and stays real.