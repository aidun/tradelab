package order

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aidun/tradelab/backend/internal/domain"
)

func TestPlaceMarketBuyReturnsErrorForZeroQuoteAmount(t *testing.T) {
	service := &Service{}

	_, err := service.PlaceMarketBuy(context.Background(), PlaceMarketBuyInput{
		UserID:       "user-1",
		WalletID:     "wallet-1",
		MarketSymbol: "XRP/USDT",
		QuoteAmount:  0,
	})

	if !errors.Is(err, ErrQuoteAmountTooLow) {
		t.Fatalf("expected ErrQuoteAmountTooLow, got %v", err)
	}
}

func TestPlaceMarketBuyReturnsErrorWhenFundsAreInsufficient(t *testing.T) {
	service := NewService(
		fakeMarketRepository{
			market: domain.Market{
				ID:          "market-1",
				Symbol:      "XRP/USDT",
				BaseAsset:   "XRP",
				QuoteAsset:  "USDT",
				MinNotional: 10,
			},
		},
		fakeBalanceRepository{
			balance: domain.Balance{
				AssetSymbol: "USDT",
				Available:   50,
			},
		},
		fakeOrderRepository{},
	)

	_, err := service.PlaceMarketBuy(context.Background(), PlaceMarketBuyInput{
		UserID:        "user-1",
		WalletID:      "wallet-1",
		MarketSymbol:  "XRP/USDT",
		QuoteAmount:   200,
		ExpectedPrice: 0.65,
	})

	if !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("expected ErrInsufficientFunds, got %v", err)
	}
}

func TestPlaceMarketBuyCreatesPendingOrder(t *testing.T) {
	now := time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)

	service := NewService(
		fakeMarketRepository{
			market: domain.Market{
				ID:          "market-1",
				Symbol:      "XRP/USDT",
				BaseAsset:   "XRP",
				QuoteAsset:  "USDT",
				MinNotional: 10,
			},
		},
		fakeBalanceRepository{
			balance: domain.Balance{
				AssetSymbol: "USDT",
				Available:   500,
			},
		},
		fakeOrderRepository{},
	)
	service.clock = fakeClock{now: now}

	order, err := service.PlaceMarketBuy(context.Background(), PlaceMarketBuyInput{
		UserID:        "user-1",
		WalletID:      "wallet-1",
		MarketSymbol:  "XRP/USDT",
		QuoteAmount:   100,
		ExpectedPrice: 0.67,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if order.Status != domain.OrderStatusPending {
		t.Fatalf("expected pending status, got %s", order.Status)
	}

	if order.UserID != "user-1" {
		t.Fatalf("expected user-1, got %s", order.UserID)
	}

	if order.MarketSymbol != "XRP/USDT" {
		t.Fatalf("expected XRP/USDT market, got %s", order.MarketSymbol)
	}

	if !order.CreatedAt.Equal(now) {
		t.Fatalf("expected created time %s, got %s", now, order.CreatedAt)
	}
}

type fakeClock struct {
	now time.Time
}

func (f fakeClock) Now() time.Time {
	return f.now
}

type fakeMarketRepository struct {
	markets []domain.Market
	market domain.Market
	err    error
}

func (f fakeMarketRepository) List(context.Context) ([]domain.Market, error) {
	return f.markets, f.err
}

func (f fakeMarketRepository) GetBySymbol(context.Context, string) (domain.Market, error) {
	return f.market, f.err
}

type fakeBalanceRepository struct {
	balance domain.Balance
	err     error
}

func (f fakeBalanceRepository) GetByWalletAndAsset(context.Context, string, string) (domain.Balance, error) {
	return f.balance, f.err
}

type fakeOrderRepository struct {
	err error
}

func (f fakeOrderRepository) Create(_ context.Context, order domain.Order) (domain.Order, error) {
	if f.err != nil {
		return domain.Order{}, f.err
	}

	order.ID = "order-1"
	return order, nil
}
