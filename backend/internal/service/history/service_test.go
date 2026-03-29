package history

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aidun/tradelab/backend/internal/domain"
	"github.com/aidun/tradelab/backend/internal/logging"
)

func TestListOrdersReturnsRepositoryOrders(t *testing.T) {
	service := NewService(fakePortfolioRepository{
		orders: []domain.Order{{ID: "order-1", WalletID: "wallet-1"}},
	}, logging.NewDiscardLogger("history_service_test"))

	orders, err := service.ListOrders(context.Background(), "wallet-1", 10, "", domain.AccountingModeAverageCost)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
}

func TestListActivityReturnsRepositoryEntries(t *testing.T) {
	service := NewService(fakePortfolioRepository{
		activity: []domain.ActivityLog{{ID: "log-1", WalletID: "wallet-1", CreatedAt: time.Now()}},
	}, logging.NewDiscardLogger("history_service_test"))

	activity, err := service.ListActivity(context.Background(), "wallet-1", 10, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(activity) != 1 {
		t.Fatalf("expected 1 activity item, got %d", len(activity))
	}
}

func TestListOrdersReturnsRepositoryError(t *testing.T) {
	service := NewService(fakePortfolioRepository{
		err: errors.New("db unavailable"),
	}, logging.NewDiscardLogger("history_service_test"))

	if _, err := service.ListOrders(context.Background(), "wallet-1", 10, "", domain.AccountingModeAverageCost); err == nil {
		t.Fatal("expected an error, got nil")
	}
}

type fakePortfolioRepository struct {
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
	return domain.PortfolioSummary{}, f.err
}

func (f fakePortfolioRepository) ListByWallet(context.Context, string, int) ([]domain.Order, error) {
	return f.orders, f.err
}

func (f fakePortfolioRepository) ListActivityByWallet(context.Context, string, int) ([]domain.ActivityLog, error) {
	return f.activity, f.err
}
