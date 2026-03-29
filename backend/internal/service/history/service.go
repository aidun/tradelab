package history

import (
	"context"

	"github.com/aidun/tradelab/backend/internal/domain"
	"github.com/aidun/tradelab/backend/internal/store"
)

type Service struct {
	orders store.PortfolioRepository
}

func NewService(orders store.PortfolioRepository) *Service {
	return &Service{orders: orders}
}

func (s *Service) ListOrders(ctx context.Context, walletID string, limit int) ([]domain.Order, error) {
	return s.orders.ListByWallet(ctx, walletID, limit)
}

func (s *Service) ListActivity(ctx context.Context, walletID string, limit int) ([]domain.ActivityLog, error) {
	return s.orders.ListActivityByWallet(ctx, walletID, limit)
}
