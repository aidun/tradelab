package portfolio

import (
	"context"
	"log/slog"

	"github.com/aidun/tradelab/backend/internal/domain"
	"github.com/aidun/tradelab/backend/internal/service/tradingcalc"
	"github.com/aidun/tradelab/backend/internal/store"
)

type Service struct {
	portfolio store.PortfolioRepository
	prices    PriceProvider
	logger    *slog.Logger
}

type PriceProvider interface {
	GetSpotPrice(ctx context.Context, marketSymbol string) (float64, error)
}

func NewService(portfolio store.PortfolioRepository, prices PriceProvider, logger *slog.Logger) *Service {
	return &Service{portfolio: portfolio, prices: prices, logger: logger}
}

func (s *Service) GetSummary(ctx context.Context, walletID string, mode domain.AccountingMode) (domain.PortfolioSummary, error) {
	summary, err := s.portfolio.GetSummary(ctx, walletID)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("get_portfolio_summary.failed", "operation", "get_portfolio_summary.failed", "wallet_id", walletID, "error", err)
		}
		return domain.PortfolioSummary{}, err
	}

	orders, err := s.portfolio.ListByWallet(ctx, walletID, 0)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("get_portfolio_orders.failed", "operation", "get_portfolio_orders.failed", "wallet_id", walletID, "error", err)
		}
		return domain.PortfolioSummary{}, err
	}

	mode = domain.NormalizeAccountingMode(string(mode))
	summary.AccountingMode = mode

	currentPrices := map[string]float64{}
	for _, order := range orders {
		if _, ok := currentPrices[order.MarketSymbol]; ok {
			continue
		}

		price, priceErr := s.prices.GetSpotPrice(ctx, order.MarketSymbol)
		if priceErr != nil && s.logger != nil {
			s.logger.Warn("get_portfolio_summary.price_fallback", "operation", "get_portfolio_summary.price_fallback", "wallet_id", walletID, "market_symbol", order.MarketSymbol)
		}
		currentPrices[order.MarketSymbol] = price
	}

	analysis := tradingcalc.AnalyzeOrders(orders, mode, currentPrices)
	summary.Positions = analysis.Positions
	summary.RealizedPnL = analysis.Realized
	summary.UnrealizedPnL = analysis.Unrealized
	summary.TotalValue = summary.CashBalance
	summary.PositionValue = 0
	summary.Allocations = []domain.PortfolioAllocation{}

	for _, position := range summary.Positions {
		summary.PositionValue += position.PositionValue
		summary.TotalValue += position.PositionValue
	}

	for _, position := range summary.Positions {
		weight := 0.0
		if summary.PositionValue > 0 {
			weight = position.PositionValue / summary.PositionValue
		}
		summary.Allocations = append(summary.Allocations, domain.PortfolioAllocation{
			MarketSymbol: position.MarketSymbol,
			Value:        position.PositionValue,
			Weight:       weight,
		})
	}

	if s.logger != nil {
		s.logger.Info("get_portfolio_summary.success", "operation", "get_portfolio_summary.success", "wallet_id", walletID, "position_count", len(summary.Positions), "accounting_mode", mode)
	}

	return summary, nil
}
