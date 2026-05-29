package monzo

import (
	"context"
	"errors"
	"net/url"
	"time"

	banksync "github.com/pennypilot/pennypilot/backend/internal/sync"
	"github.com/pennypilot/pennypilot/backend/internal/domain"
)

const providerName = "monzo"

// Connector is a placeholder for Monzo's direct OAuth2 integration.
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
	values.Set("state", state)

	return "https://auth.monzo.com/?" + values.Encode()
}

func (c Connector) Exchange(ctx context.Context, code string) (*banksync.Credentials, error) {
	return nil, errors.New("monzo credential exchange is not implemented yet")
}

func (c Connector) FetchTransactions(ctx context.Context, creds *banksync.Credentials, accountID string, since time.Time) ([]domain.Transaction, error) {
	return nil, errors.New("monzo transaction sync is not implemented yet")
}

func (c Connector) FetchAccounts(ctx context.Context, creds *banksync.Credentials) ([]domain.Account, error) {
	return nil, errors.New("monzo account sync is not implemented yet")
}

func (c Connector) RefreshCredentials(ctx context.Context, creds *banksync.Credentials) (*banksync.Credentials, error) {
	return nil, errors.New("monzo credential refresh is not implemented yet")
}
