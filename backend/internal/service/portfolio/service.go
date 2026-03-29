package portfolio

import (
	"context"
	"log/slog"

	"github.com/aidun/tradelab/backend/internal/domain"
	"github.com/aidun/tradelab/backend/internal/store"
)

type Service struct {
	portfolio store.PortfolioRepository
	logger    *slog.Logger
}

func NewService(portfolio store.PortfolioRepository, logger *slog.Logger) *Service {
	return &Service{portfolio: portfolio, logger: logger}
}

func (s *Service) GetSummary(ctx context.Context, walletID string) (domain.PortfolioSummary, error) {
	summary, err := s.portfolio.GetSummary(ctx, walletID)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("get_portfolio_summary.failed", "operation", "get_portfolio_summary.failed", "wallet_id", walletID, "error", err)
		}
		return domain.PortfolioSummary{}, err
	}

	if s.logger != nil {
		s.logger.Info("get_portfolio_summary.success", "operation", "get_portfolio_summary.success", "wallet_id", walletID, "position_count", len(summary.Positions))
	}

	return summary, nil
}
