package domain

import "time"

// Account represents a connected bank, card, or manually imported account.
type Account struct {
	ID             string    `json:"id"`
	Provider       string    `json:"provider"`
	ExternalID     string    `json:"externalId"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Currency       string    `json:"currency"`
	BalancePence   int64     `json:"balancePence"`
	LastSyncedAt   time.Time `json:"lastSyncedAt,omitempty"`
	ReconnectAfter time.Time `json:"reconnectAfter,omitempty"`
}

// Transaction is the normalized ledger entry used across all connectors.
type Transaction struct {
	ID           string    `json:"id"`
	AccountID    string    `json:"accountId"`
	ExternalID   string    `json:"externalId"`
	CategoryID   string    `json:"categoryId,omitempty"`
	AmountPence  int64     `json:"amountPence"`
	Currency     string    `json:"currency"`
	Description  string    `json:"description"`
	MerchantName string    `json:"merchantName,omitempty"`
	OccurredAt   time.Time `json:"occurredAt"`
	Pending      bool      `json:"pending"`
	Notes        string    `json:"notes,omitempty"`
	ImportedAt   time.Time `json:"importedAt"`
}

// Category groups transactions for reporting and budgeting.
type Category struct {
	ID       string `json:"id"`
	ParentID string `json:"parentId,omitempty"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Icon     string `json:"icon"`
}

// Budget stores a monthly plan for a category in pence.
type Budget struct {
	ID          string    `json:"id"`
	CategoryID  string    `json:"categoryId"`
	Month       time.Time `json:"month"`
	AmountPence int64     `json:"amountPence"`
}
