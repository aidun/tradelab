package order

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aidun/tradelab/backend/internal/domain"
	"github.com/aidun/tradelab/backend/internal/logging"
)

func TestPlaceMarketBuyReturnsErrorForZeroQuoteAmount(t *testing.T) {
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
		fakeBalanceRepository{},
		fakePortfolioRepository{},
		fakePriceProvider{price: 0.67},
		logging.NewDiscardLogger("order_service_test"),
	)

	_, err := service.PlaceMarketOrder(context.Background(), PlaceMarketOrderInput{
		UserID:       "user-1",
		WalletID:     "wallet-1",
		MarketSymbol: "XRP/USDT",
		Side:         domain.OrderSideBuy,
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
		fakePortfolioRepository{},
		fakePriceProvider{price: 0.65},
		logging.NewDiscardLogger("order_service_test"),
	)

	_, err := service.PlaceMarketOrder(context.Background(), PlaceMarketOrderInput{
		UserID:       "user-1",
		WalletID:     "wallet-1",
		MarketSymbol: "XRP/USDT",
		Side:         domain.OrderSideBuy,
		QuoteAmount:  200,
	})

	if !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("expected ErrInsufficientFunds, got %v", err)
	}
}

func TestPlaceMarketBuyCreatesFilledOrder(t *testing.T) {
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
		fakePortfolioRepository{},
		fakePriceProvider{price: 0.67},
		logging.NewDiscardLogger("order_service_test"),
	)
	service.clock = fakeClock{now: now}

	order, err := service.PlaceMarketOrder(context.Background(), PlaceMarketOrderInput{
		UserID:       "user-1",
		WalletID:     "wallet-1",
		MarketSymbol: "XRP/USDT",
		Side:         domain.OrderSideBuy,
		QuoteAmount:  100,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if order.Status != domain.OrderStatusFilled {
		t.Fatalf("expected filled status, got %s", order.Status)
	}

	if order.BaseQuantity <= 0 {
		t.Fatalf("expected positive base quantity, got %f", order.BaseQuantity)
	}

	if !order.CreatedAt.Equal(now) {
		t.Fatalf("expected created time %s, got %s", now, order.CreatedAt)
	}

	if order.ExpectedPrice != 0.67 {
		t.Fatalf("expected server-side price 0.67, got %f", order.ExpectedPrice)
	}
}

func TestPlaceMarketBuyReturnsErrorWhenPriceProviderHasNoPrice(t *testing.T) {
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
		fakePortfolioRepository{},
		fakePriceProvider{price: 0},
		logging.NewDiscardLogger("order_service_test"),
	)

	_, err := service.PlaceMarketOrder(context.Background(), PlaceMarketOrderInput{
		UserID:       "user-1",
		WalletID:     "wallet-1",
		MarketSymbol: "XRP/USDT",
		Side:         domain.OrderSideBuy,
		QuoteAmount:  100,
	})
	if !errors.Is(err, ErrCurrentPriceUnavailable) {
		t.Fatalf("expected ErrCurrentPriceUnavailable, got %v", err)
	}
}

func TestPlaceMarketBuyReturnsErrorWhenPriceProviderFails(t *testing.T) {
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
		fakePortfolioRepository{},
		fakePriceProvider{err: errors.New("upstream failed")},
		logging.NewDiscardLogger("order_service_test"),
	)

	_, err := service.PlaceMarketOrder(context.Background(), PlaceMarketOrderInput{
		UserID:       "user-1",
		WalletID:     "wallet-1",
		MarketSymbol: "XRP/USDT",
		Side:         domain.OrderSideBuy,
		QuoteAmount:  100,
	})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestPlaceMarketBuyReturnsErrorWhenBelowMarketMinimum(t *testing.T) {
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
		fakePortfolioRepository{},
		fakePriceProvider{price: 0.67},
		logging.NewDiscardLogger("order_service_test"),
	)

	_, err := service.PlaceMarketOrder(context.Background(), PlaceMarketOrderInput{
		UserID:       "user-1",
		WalletID:     "wallet-1",
		MarketSymbol: "XRP/USDT",
		Side:         domain.OrderSideBuy,
		QuoteAmount:  5,
	})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestPlaceMarketSellRejectsOversizedPosition(t *testing.T) {
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
				AssetSymbol: "XRP",
				Available:   10,
			},
		},
		fakePortfolioRepository{},
		fakePriceProvider{price: 0.67},
		logging.NewDiscardLogger("order_service_test"),
	)

	_, err := service.PlaceMarketOrder(context.Background(), PlaceMarketOrderInput{
		UserID:       "user-1",
		WalletID:     "wallet-1",
		MarketSymbol: "XRP/USDT",
		Side:         domain.OrderSideSell,
		BaseQuantity: 25,
	})
	if !errors.Is(err, ErrInsufficientPosition) {
		t.Fatalf("expected ErrInsufficientPosition, got %v", err)
	}
}

func TestPlaceMarketSellCreatesFilledOrder(t *testing.T) {
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
				AssetSymbol: "XRP",
				Available:   250,
			},
		},
		fakePortfolioRepository{},
		fakePriceProvider{price: 0.7},
		logging.NewDiscardLogger("order_service_test"),
	)
	service.clock = fakeClock{now: now}

	order, err := service.PlaceMarketOrder(context.Background(), PlaceMarketOrderInput{
		UserID:       "user-1",
		WalletID:     "wallet-1",
		MarketSymbol: "XRP/USDT",
		Side:         domain.OrderSideSell,
		BaseQuantity: 100,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if order.Side != domain.OrderSideSell {
		t.Fatalf("expected sell side, got %s", order.Side)
	}

	if order.QuoteAmount != 70 {
		t.Fatalf("expected quote proceeds 70, got %f", order.QuoteAmount)
	}
}

type fakeClock struct{ now time.Time }

func (f fakeClock) Now() time.Time { return f.now }

type fakeMarketRepository struct {
	markets []domain.Market
	market  domain.Market
	err     error
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

type fakePortfolioRepository struct {
	err error
}

func (f fakePortfolioRepository) ApplyMarketBuy(_ context.Context, order domain.Order) (domain.Order, error) {
	if f.err != nil {
		return domain.Order{}, f.err
	}
	return order, nil
}

func (f fakePortfolioRepository) ApplyMarketSell(_ context.Context, order domain.Order) (domain.Order, error) {
	if f.err != nil {
		return domain.Order{}, f.err
	}
	return order, nil
}

func (f fakePortfolioRepository) GetSummary(context.Context, string) (domain.PortfolioSummary, error) {
	return domain.PortfolioSummary{}, f.err
}

func (f fakePortfolioRepository) ListByWallet(context.Context, string, int) ([]domain.Order, error) {
	return nil, f.err
}

func (f fakePortfolioRepository) ListActivityByWallet(context.Context, string, int) ([]domain.ActivityLog, error) {
	return nil, f.err
}

type fakePriceProvider struct {
	price float64
	err   error
}

func (f fakePriceProvider) GetSpotPrice(context.Context, string) (float64, error) {
	return f.price, f.err
}
