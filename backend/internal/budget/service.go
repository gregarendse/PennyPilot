package budget

import (
	"context"
	"time"

	"github.com/pennypilot/pennypilot/backend/internal/domain"
)

// Service coordinates budget calculations once transactions and categories are persisted.
type Service struct{}

func NewService() Service { return Service{} }

func (s Service) MonthlySummary(ctx context.Context, month time.Time) ([]domain.Budget, error) {
	return []domain.Budget{}, nil
}
