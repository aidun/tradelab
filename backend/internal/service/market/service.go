package market

import (
	"context"

	"github.com/aidun/tradelab/backend/internal/domain"
	"github.com/aidun/tradelab/backend/internal/store"
)

type Service struct {
	markets store.MarketRepository
}

func NewService(markets store.MarketRepository) *Service {
	return &Service{markets: markets}
}

func (s *Service) List(ctx context.Context) ([]domain.Market, error) {
	return s.markets.List(ctx)
}
