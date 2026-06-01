package csv

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/pennypilot/pennypilot/backend/internal/domain"
	banksync "github.com/pennypilot/pennypilot/backend/internal/sync"
)

const providerName = "csv"

// Importer will normalize CSV and OFX exports from providers such as American Express.
type Importer struct{}

func NewImporter() Importer { return Importer{} }

func (i Importer) Name() string { return providerName }

func (i Importer) AuthURL(state string) string {
	return ""
}

func (i Importer) Exchange(ctx context.Context, code string) (*banksync.Credentials, error) {
	return nil, errors.New("csv does not support oauth exchange")
}

func (i Importer) FetchTransactions(ctx context.Context, creds *banksync.Credentials, accountID string, since time.Time) ([]domain.Transaction, error) {
	return nil, errors.New("csv transactions must be imported via file upload")
}

func (i Importer) FetchAccounts(ctx context.Context, creds *banksync.Credentials) ([]domain.Account, error) {
	return []domain.Account{}, nil
}

func (i Importer) RefreshCredentials(ctx context.Context, creds *banksync.Credentials) (*banksync.Credentials, error) {
	return creds, nil
}

func (i Importer) Ping(ctx context.Context) error {
	return nil
}

func (i Importer) DetectColumns(ctx context.Context, reader io.Reader) (map[string]string, error) {
	r := csv.NewReader(reader)
	header, err := r.Read()
	if err != nil {
		return nil, err
	}

	return i.getMapping(header), nil
}

func (i Importer) getMapping(header []string) map[string]string {
	mapping := make(map[string]string)
	for i, col := range header {
		col = strings.ToLower(col)
		if strings.Contains(col, "date") && mapping["date"] == "" {
			mapping["date"] = strconv.Itoa(i)
		} else if strings.Contains(col, "amount") && mapping["amount"] == "" {
			mapping["amount"] = strconv.Itoa(i)
		} else if (strings.Contains(col, "description") || strings.Contains(col, "memo")) && mapping["description"] == "" {
			mapping["description"] = strconv.Itoa(i)
		}
	}
	return mapping
}

func (i Importer) ImportTransactions(ctx context.Context, accountID string, reader io.Reader) ([]domain.Transaction, error) {
	r := csv.NewReader(reader)
	header, err := r.Read()
	if err != nil {
		return nil, err
	}

	mapping := i.getMapping(header)
	dateIdx, _ := strconv.Atoi(mapping["date"])
	amountIdx, _ := strconv.Atoi(mapping["amount"])
	descIdx, _ := strconv.Atoi(mapping["description"])

	if mapping["date"] == "" || mapping["amount"] == "" || mapping["description"] == "" {
		return nil, errors.New("could not detect required CSV columns (date, amount, description)")
	}

	var transactions []domain.Transaction
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		occurredAt, err := parseDate(record[dateIdx])
		if err != nil {
			return nil, fmt.Errorf("failed to parse date %q: %w", record[dateIdx], err)
		}
		amountFloat, err := strconv.ParseFloat(record[amountIdx], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse amount %q: %w", record[amountIdx], err)
		}
		amountPence := int64(math.Round(amountFloat * 100))

		transactions = append(transactions, domain.Transaction{
			AccountID:   accountID,
			AmountPence: amountPence,
			Description: record[descIdx],
			OccurredAt:  occurredAt,
			ImportedAt:  time.Now(),
		})
	}

	return transactions, nil
}

func parseDate(s string) (time.Time, error) {
	formats := []string{"2006-01-02", "02/01/2006", "01/02/2006", "2006/01/02"}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse date: %s", s)
}
