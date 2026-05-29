package monzo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pennypilot/pennypilot/backend/internal/domain"
	banksync "github.com/pennypilot/pennypilot/backend/internal/sync"
)

const providerName = "monzo"

// Connector is a placeholder for Monzo's direct OAuth2 integration.
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
	values.Set("state", state)

	return "https://auth.monzo.com/?" + values.Encode()
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.monzo.com/oauth2/token", strings.NewReader(values.Encode()))
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
		return nil, fmt.Errorf("monzo token request failed: %s", resp.Status)
	}

	var res struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		UserID       string `json:"user_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &banksync.Credentials{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(res.ExpiresIn) * time.Second),
		Metadata:     map[string]string{"user_id": res.UserID},
	}, nil
}

func (c Connector) FetchAccounts(ctx context.Context, creds *banksync.Credentials) ([]domain.Account, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.monzo.com/accounts", nil)
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
		return nil, fmt.Errorf("monzo fetch accounts failed: %s", resp.Status)
	}

	var res struct {
		Accounts []struct {
			ID          string `json:"id"`
			Description string `json:"description"`
			Type        string `json:"type"`
			Currency    string `json:"currency"`
			Closed      bool   `json:"closed"`
		} `json:"accounts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	accounts := make([]domain.Account, 0, len(res.Accounts))
	for _, a := range res.Accounts {
		if a.Closed {
			continue
		}
		accounts = append(accounts, domain.Account{
			ExternalID: a.ID,
			Provider:   providerName,
			Name:       a.Description,
			Type:       a.Type,
			Currency:   a.Currency,
		})
	}

	return accounts, nil
}

func (c Connector) FetchTransactions(ctx context.Context, creds *banksync.Credentials, accountID string, since time.Time) ([]domain.Transaction, error) {
	values := url.Values{}
	values.Set("account_id", accountID)
	values.Set("since", since.Format(time.RFC3339))
	values.Set("expand[]", "merchant")

	u := "https://api.monzo.com/transactions?" + values.Encode()
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
		return nil, fmt.Errorf("monzo fetch transactions failed: %s", resp.Status)
	}

	var res struct {
		Transactions []struct {
			ID          string    `json:"id"`
			Amount      int64     `json:"amount"`
			Currency    string    `json:"currency"`
			Description string    `json:"description"`
			Notes       string    `json:"notes"`
			Created     time.Time `json:"created"`
			IsLoad      bool      `json:"is_load"`
			Merchant    struct {
				Name string `json:"name"`
			} `json:"merchant"`
		} `json:"transactions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	transactions := make([]domain.Transaction, 0, len(res.Transactions))
	for _, t := range res.Transactions {
		transactions = append(transactions, domain.Transaction{
			ExternalID:   t.ID,
			AmountPence:  t.Amount, // Monzo amounts are already in pence
			Currency:     t.Currency,
			Description:  t.Description,
			MerchantName: t.Merchant.Name,
			Notes:        t.Notes,
			OccurredAt:   t.Created,
			Pending:      false, // Monzo transactions in this list are usually settled or we need to check declined
		})
	}

	return transactions, nil
}

func (c Connector) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.monzo.com/ping", nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		return fmt.Errorf("monzo api returned status %d", resp.StatusCode)
	}
	return nil
}
