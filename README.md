# PennyPilot — Personal Finance Tracker

A self-hosted budget tracker with automatic bank sync for Monzo, Barclays, and Barclaycard.

## Architecture

```
Next.js frontend  →  Go REST API  →  PostgreSQL
                                  ↳  Monzo API (direct OAuth2)
                                  ↳  TrueLayer (Barclays, Barclaycard)
                                  ↳  CSV import (Amex, others)
```

## Getting Started

### 1. Register API credentials (do this first)

**Monzo Developer Account**
1. Go to https://developers.monzo.com
2. Sign in with your Monzo account
3. Create a new OAuth client
4. Set redirect URI: `http://localhost:8080/auth/monzo/callback`
5. Copy `client_id` and `client_secret`

**TrueLayer Sandbox** (for Barclays — use sandbox until ready for real connection)
1. Go to https://console.truelayer.com
2. Create account + new app
3. Set redirect URI: `http://localhost:8080/auth/truelayer/callback`
4. Copy credentials

### 2. Configure environment

```bash
cp .env.example .env
# Edit .env with your credentials
# Generate encryption key:
openssl rand -hex 32
```

### 3. Run

```bash
docker compose up db -d          # Start Postgres only
cd backend && go run ./cmd/server  # Run backend locally (auto-migrates DB)
```

### Self-Hosted Homelab

For automated deployment to a Kubernetes-based homelab (e.g., [gregarendse/homelab](https://github.com/gregarendse/homelab)), see the [Deployment Guide](DEPLOYMENT.md).

Or run everything in Docker:
```bash
docker compose up --build
```

### 4. Connect your Monzo account

Visit http://localhost:8080/auth/monzo in your browser. You'll be redirected to Monzo to authorise access, then redirected back. That's it — transactions start flowing.

### 5. Verify it works

```bash
curl http://localhost:8080/api/accounts
curl http://localhost:8080/api/transactions
curl http://localhost:8080/api/categories
```


## Project layout

```
/
├── backend/                 # Go REST API and bank connector boundary
│   ├── cmd/server/          # API entrypoint
│   ├── internal/            # API, config, domain, store, budget, and sync packages
│   └── migrations/          # PostgreSQL schema migrations
├── frontend/                # Next.js App Router UI scaffold
│   ├── app/                 # Routes, layout, and global styles
│   ├── components/          # Reusable shell and presentation components
│   └── lib/                 # API client helpers and shared frontend types
├── docker-compose.yml       # Local Postgres, backend, and frontend services
└── .env.example             # Required local configuration
```

The current implementation is intentionally a bootstrap scaffold: API routes return empty placeholder payloads, connector packages expose the target interfaces, and the frontend renders a static product shell while persistence and provider integrations are filled in.

## Development

```bash
# Backend only (fastest iteration)
cd backend
go run ./cmd/server

# With live reload
go install github.com/air-verse/air@latest
air

# Run tests
go test ./...
```

## Adding a new bank connector

Every bank connector implements the `sync.BankConnector` interface in `backend/internal/sync/connector.go`. To add a new bank:

1. Create `backend/internal/sync/yourbank/connector.go`
2. Implement the 5 interface methods
3. Register it in `cmd/server/main.go`
4. Add the auth route in `internal/api/handler.go`

## Monzo 90-day re-auth

Monzo's security model requires re-authorisation every 90 days. This is by design — it prevents long-lived token abuse. The app will surface a "re-connect" prompt when your token is about to expire.

## American Express

Amex does not participate in UK Open Banking (PSD2). The planned approach is a CSV import with auto-column detection. Amex lets you export transactions as CSV or OFX from their website.