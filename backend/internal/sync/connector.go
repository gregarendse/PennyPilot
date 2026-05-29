package sync

import (
	"context"
	"time"

	"github.com/pennypilot/pennypilot/backend/internal/domain"
)

// Credentials holds OAuth and import metadata for a bank connector. Values must be encrypted before persistence.
type Credentials struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	Scopes       []string
	Metadata     map[string]string
}

// BankConnector is implemented by every bank or import source integration.
type BankConnector interface {
	Name() string
	AuthURL(state string) string
	Exchange(ctx context.Context, code string) (*Credentials, error)
	FetchTransactions(ctx context.Context, creds *Credentials, accountID string, since time.Time) ([]domain.Transaction, error)
	FetchAccounts(ctx context.Context, creds *Credentials) ([]domain.Account, error)
	RefreshCredentials(ctx context.Context, creds *Credentials) (*Credentials, error)
}
