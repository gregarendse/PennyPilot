package gocardless

import (
	"context"
	"errors"
	"time"

	"github.com/pennypilot/pennypilot/backend/internal/domain"
	banksync "github.com/pennypilot/pennypilot/backend/internal/sync"
)

const providerName = "gocardless"

// Connector is a placeholder for GoCardless (Nordigen) Open Banking sync.
type Connector struct {
	SecretID  string
	SecretKey string
}

func New(secretID, secretKey string) Connector {
	return Connector{SecretID: secretID, SecretKey: secretKey}
}

func (c Connector) Name() string { return providerName }

func (c Connector) AuthURL(state string) string {
	// GoCardless (Nordigen) uses a multi-step process for requisitions:
	// 1. Get an access token using Secret ID and Secret Key.
	// 2. Create a requisition to get a redirect link to the bank.
	// For now, this returns a placeholder as full implementation requires API calls.
	return "https://bankaccountdata.gocardless.com/api/v2/requisitions/?state=" + state
}

func (c Connector) Exchange(ctx context.Context, code string) (*banksync.Credentials, error) {
	return nil, errors.New("gocardless credential exchange is not implemented yet")
}

func (c Connector) FetchTransactions(ctx context.Context, creds *banksync.Credentials, accountID string, since time.Time) ([]domain.Transaction, error) {
	return nil, errors.New("gocardless transaction sync is not implemented yet")
}

func (c Connector) FetchAccounts(ctx context.Context, creds *banksync.Credentials) ([]domain.Account, error) {
	return nil, errors.New("gocardless account sync is not implemented yet")
}

func (c Connector) RefreshCredentials(ctx context.Context, creds *banksync.Credentials) (*banksync.Credentials, error) {
	return nil, errors.New("gocardless credential refresh is not implemented yet")
}
