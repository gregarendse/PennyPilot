package gocardless

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
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
	// In GoCardless, 'code' here is likely the requisition ID passed back in the redirect.
	// We'll store it in Metadata so we can use it to fetch accounts.
	token, err := c.getNewToken(ctx)
	if err != nil {
		return nil, err
	}

	return &banksync.Credentials{
		AccessToken:  token.Access,
		RefreshToken: token.Refresh,
		ExpiresAt:    time.Now().Add(time.Duration(token.AccessExpires) * time.Second),
		Metadata:     map[string]string{"requisition_id": code},
	}, nil
}

type tokenResponse struct {
	Access        string `json:"access"`
	AccessExpires int    `json:"access_expires"`
	Refresh       string `json:"refresh"`
	RefreshExpires int   `json:"refresh_expires"`
}

func (c Connector) getNewToken(ctx context.Context) (*tokenResponse, error) {
	payload := map[string]string{
		"secret_id":  c.SecretID,
		"secret_key": c.SecretKey,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://bankaccountdata.gocardless.com/api/v2/token/new/", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gocardless token request failed: %s", resp.Status)
	}

	var res tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c Connector) FetchAccounts(ctx context.Context, creds *banksync.Credentials) ([]domain.Account, error) {
	requisitionID := creds.Metadata["requisition_id"]
	if requisitionID == "" {
		return nil, errors.New("missing requisition_id in credentials")
	}

	u := fmt.Sprintf("https://bankaccountdata.gocardless.com/api/v2/requisitions/%s/", requisitionID)
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

	var res struct {
		Accounts []string `json:"accounts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	accounts := make([]domain.Account, 0, len(res.Accounts))
	for _, id := range res.Accounts {
		accounts = append(accounts, domain.Account{
			ExternalID: id,
			Provider:   providerName,
			Name:       "GoCardless Account", // Detailed info requires another API call per account
		})
	}

	return accounts, nil
}

func (c Connector) FetchTransactions(ctx context.Context, creds *banksync.Credentials, accountID string, since time.Time) ([]domain.Transaction, error) {
	u := fmt.Sprintf("https://bankaccountdata.gocardless.com/api/v2/accounts/%s/transactions/", accountID)
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

	var res struct {
		Transactions struct {
			Booked []struct {
				TransactionID     string `json:"transactionId"`
				BookingDate       string `json:"bookingDate"`
				ValueDate         string `json:"valueDate"`
				TransactionAmount struct {
					Amount   string `json:"amount"`
					Currency string `json:"currency"`
				} `json:"transactionAmount"`
				RemittanceInformationUnstructured string `json:"remittanceInformationUnstructured"`
			} `json:"booked"`
		} `json:"transactions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	transactions := make([]domain.Transaction, 0, len(res.Transactions.Booked))
	for _, t := range res.Transactions.Booked {
		occurredAt, _ := time.Parse("2006-01-02", t.BookingDate)
		amountFloat, _ := strconv.ParseFloat(t.TransactionAmount.Amount, 64)
		amountPence := int64(math.Round(amountFloat * 100))

		transactions = append(transactions, domain.Transaction{
			ExternalID:  t.TransactionID,
			AmountPence: amountPence,
			Currency:    t.TransactionAmount.Currency,
			Description: t.RemittanceInformationUnstructured,
			OccurredAt:  occurredAt,
		})
	}

	return transactions, nil
}

func (c Connector) RefreshCredentials(ctx context.Context, creds *banksync.Credentials) (*banksync.Credentials, error) {
	payload := map[string]string{
		"refresh": creds.RefreshToken,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://bankaccountdata.gocardless.com/api/v2/token/refresh/", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res struct {
		Access        string `json:"access"`
		AccessExpires int    `json:"access_expires"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	creds.AccessToken = res.Access
	creds.ExpiresAt = time.Now().Add(time.Duration(res.AccessExpires) * time.Second)

	return creds, nil
}

func (c Connector) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://bankaccountdata.gocardless.com/api/v2/", nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		return fmt.Errorf("gocardless api returned status %d", resp.StatusCode)
	}
	return nil
}
