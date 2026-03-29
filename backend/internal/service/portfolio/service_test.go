package portfolio

import (
	"context"
	"errors"
	"testing"

	"github.com/aidun/tradelab/backend/internal/domain"
	"github.com/aidun/tradelab/backend/internal/logging"
)

func TestGetSummaryReturnsRepositoryResult(t *testing.T) {
	service := NewService(fakePortfolioRepository{
		summary: domain.PortfolioSummary{
			WalletID:     "wallet-1",
			CashBalance:  10000,
			BaseCurrency: "USDT",
		},
	}, fakePortfolioPriceProvider{}, logging.NewDiscardLogger("portfolio_service_test"))

	summary, err := service.GetSummary(context.Background(), "wallet-1", domain.AccountingModeAverageCost)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if summary.WalletID != "wallet-1" {
		t.Fatalf("expected wallet-1, got %s", summary.WalletID)
	}
}

func TestGetSummaryReturnsRepositoryError(t *testing.T) {
	service := NewService(fakePortfolioRepository{
		err: errors.New("db unavailable"),
	}, fakePortfolioPriceProvider{}, logging.NewDiscardLogger("portfolio_service_test"))

	if _, err := service.GetSummary(context.Background(), "wallet-1", domain.AccountingModeAverageCost); err == nil {
		t.Fatal("expected an error, got nil")
	}
}

type fakePortfolioRepository struct {
	summary  domain.PortfolioSummary
	err      error
	orders   []domain.Order
	activity []domain.ActivityLog
}

func (f fakePortfolioRepository) ApplyMarketBuy(context.Context, domain.Order) (domain.Order, error) {
	return domain.Order{}, nil
}

func (f fakePortfolioRepository) ApplyMarketSell(context.Context, domain.Order) (domain.Order, error) {
	return domain.Order{}, nil
}

func (f fakePortfolioRepository) GetSummary(context.Context, string) (domain.PortfolioSummary, error) {
	return f.summary, f.err
}

func (f fakePortfolioRepository) ListByWallet(context.Context, string, int) ([]domain.Order, error) {
	return f.orders, f.err
}

func (f fakePortfolioRepository) ListActivityByWallet(context.Context, string, int) ([]domain.ActivityLog, error) {
	return f.activity, f.err
}

type fakePortfolioPriceProvider struct {
	price float64
	err   error
}

func (f fakePortfolioPriceProvider) GetSpotPrice(context.Context, string) (float64, error) {
	return f.price, f.err
}
