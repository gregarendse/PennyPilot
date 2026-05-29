package truelayer

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pennypilot/pennypilot/backend/internal/domain"
	banksync "github.com/pennypilot/pennypilot/backend/internal/sync"
)

const providerName = "truelayer"

// Connector is a placeholder for Barclays and Barclaycard Open Banking sync via TrueLayer.
type Connector struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func New(clientID, clientSecret, redirectURL string) Connector {
	return Connector{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
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
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("client_id", c.ClientID)
	values.Set("client_secret", c.ClientSecret)
	values.Set("redirect_uri", c.RedirectURL)
	values.Set("code", code)

	return c.doTokenRequest(ctx, values)
}

func (c Connector) RefreshCredentials(ctx context.Context, creds *banksync.Credentials) (*banksync.Credentials, error) {
	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("client_id", c.ClientID)
	values.Set("client_secret", c.ClientSecret)
	values.Set("refresh_token", creds.RefreshToken)

	return c.doTokenRequest(ctx, values)
}

func (c Connector) doTokenRequest(ctx context.Context, values url.Values) (*banksync.Credentials, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://auth.truelayer.com/connect/token", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("truelayer token request failed: %s", resp.Status)
	}

	var res struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &banksync.Credentials{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(res.ExpiresIn) * time.Second),
	}, nil
}

func (c Connector) FetchAccounts(ctx context.Context, creds *banksync.Credentials) ([]domain.Account, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.truelayer.com/data/v1/accounts", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+creds.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("truelayer fetch accounts failed: %s", resp.Status)
	}

	var res struct {
		Results []struct {
			AccountID   string `json:"account_id"`
			AccountType string `json:"account_type"`
			DisplayName string `json:"display_name"`
			Currency    string `json:"currency"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	accounts := make([]domain.Account, 0, len(res.Results))
	for _, a := range res.Results {
		accounts = append(accounts, domain.Account{
			ExternalID: a.AccountID,
			Provider:   providerName,
			Name:       a.DisplayName,
			Type:       a.AccountType,
			Currency:   a.Currency,
		})
	}

	return accounts, nil
}

func (c Connector) FetchTransactions(ctx context.Context, creds *banksync.Credentials, accountID string, since time.Time) ([]domain.Transaction, error) {
	u := fmt.Sprintf("https://api.truelayer.com/data/v1/accounts/%s/transactions", accountID)
	values := url.Values{}
	values.Set("from", since.Format(time.RFC3339))
	u += "?" + values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+creds.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("truelayer fetch transactions failed: %s", resp.Status)
	}

	var res struct {
		Results []struct {
			TransactionID string    `json:"transaction_id"`
			Timestamp     time.Time `json:"timestamp"`
			Description   string    `json:"description"`
			Amount        float64   `json:"amount"`
			Currency      string    `json:"currency"`
			MerchantName  string    `json:"merchant_name"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	transactions := make([]domain.Transaction, 0, len(res.Results))
	for _, t := range res.Results {
		transactions = append(transactions, domain.Transaction{
			ExternalID:   t.TransactionID,
			AmountPence:  int64(math.Round(t.Amount * 100)),
			Currency:     t.Currency,
			Description:  t.Description,
			MerchantName: t.MerchantName,
			OccurredAt:   t.Timestamp,
			Pending:      false,
		})
	}

	return transactions, nil
}

func (c Connector) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.truelayer.com/", nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		return fmt.Errorf("truelayer api returned status %d", resp.StatusCode)
	}
	return nil
}
