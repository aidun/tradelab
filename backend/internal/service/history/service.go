package history

import (
	"context"
	"log/slog"

	"github.com/aidun/tradelab/backend/internal/domain"
	"github.com/aidun/tradelab/backend/internal/store"
)

type Service struct {
	orders store.PortfolioRepository
	logger *slog.Logger
}

func NewService(orders store.PortfolioRepository, logger *slog.Logger) *Service {
	return &Service{orders: orders, logger: logger}
}

func (s *Service) ListOrders(ctx context.Context, walletID string, limit int) ([]domain.Order, error) {
	orders, err := s.orders.ListByWallet(ctx, walletID, limit)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("list_order_history.failed", "operation", "list_order_history.failed", "wallet_id", walletID, "limit", limit, "error", err)
		}
		return nil, err
	}

	if s.logger != nil {
		s.logger.Info("list_order_history.success", "operation", "list_order_history.success", "wallet_id", walletID, "limit", limit, "result_count", len(orders))
	}

	return orders, nil
}

func (s *Service) ListActivity(ctx context.Context, walletID string, limit int) ([]domain.ActivityLog, error) {
	activity, err := s.orders.ListActivityByWallet(ctx, walletID, limit)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("list_activity_history.failed", "operation", "list_activity_history.failed", "wallet_id", walletID, "limit", limit, "error", err)
		}
		return nil, err
	}

	if s.logger != nil {
		s.logger.Info("list_activity_history.success", "operation", "list_activity_history.success", "wallet_id", walletID, "limit", limit, "result_count", len(activity))
	}

	return activity, nil
}
