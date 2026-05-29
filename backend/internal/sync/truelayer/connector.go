package truelayer

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/pennypilot/pennypilot/backend/internal/domain"
	banksync "github.com/pennypilot/pennypilot/backend/internal/sync"
)

const providerName = "truelayer"

// Connector is a placeholder for Barclays and Barclaycard Open Banking sync via TrueLayer.
type Connector struct {
	ClientID    string
	RedirectURL string
}

func New(clientID, redirectURL string) Connector {
	return Connector{ClientID: clientID, RedirectURL: redirectURL}
}

func (c Connector) Name() string { return providerName }

func (c Connector) AuthURL(state string) string {
	values := url.Values{}
	values.Set("client_id", c.ClientID)
	values.Set("redirect_uri", c.RedirectURL)
	values.Set("response_type", "code")
	values.Set("scope", "info accounts balance cards transactions offline_access")
	values.Set("state", state)

	return "https://auth.truelayer.com/?" + values.Encode()
}

func (c Connector) Exchange(ctx context.Context, code string) (*banksync.Credentials, error) {
	return nil, errors.New("truelayer credential exchange is not implemented yet")
}

func (c Connector) FetchTransactions(ctx context.Context, creds *banksync.Credentials, accountID string, since time.Time) ([]domain.Transaction, error) {
	return nil, errors.New("truelayer transaction sync is not implemented yet")
}

func (c Connector) FetchAccounts(ctx context.Context, creds *banksync.Credentials) ([]domain.Account, error) {
	return nil, errors.New("truelayer account sync is not implemented yet")
}

func (c Connector) RefreshCredentials(ctx context.Context, creds *banksync.Credentials) (*banksync.Credentials, error) {
	return nil, errors.New("truelayer credential refresh is not implemented yet")
}
