package csv

import (
	"context"
	"errors"
	"io"

	"github.com/pennypilot/pennypilot/backend/internal/domain"
)

// Importer will normalize CSV and OFX exports from providers such as American Express.
type Importer struct{}

func NewImporter() Importer { return Importer{} }

func (i Importer) DetectColumns(ctx context.Context, reader io.Reader) (map[string]string, error) {
	return nil, errors.New("csv column detection is not implemented yet")
}

func (i Importer) ImportTransactions(ctx context.Context, accountID string, reader io.Reader) ([]domain.Transaction, error) {
	return nil, errors.New("csv transaction import is not implemented yet")
}
