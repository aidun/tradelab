package portfolio

import (
	"context"

	"github.com/aidun/tradelab/backend/internal/domain"
	"github.com/aidun/tradelab/backend/internal/store"
)

type Service struct {
	portfolio store.PortfolioRepository
}

func NewService(portfolio store.PortfolioRepository) *Service {
	return &Service{portfolio: portfolio}
}

func (s *Service) GetSummary(ctx context.Context, walletID string) (domain.PortfolioSummary, error) {
	return s.portfolio.GetSummary(ctx, walletID)
}
