package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/aidun/tradelab/backend/internal/domain"
	"github.com/aidun/tradelab/backend/internal/store"
)

var (
	ErrQuoteAmountTooLow       = errors.New("quote amount must be greater than zero")
	ErrCurrentPriceUnavailable = errors.New("current market price is unavailable")
	ErrInsufficientFunds       = errors.New("insufficient quote balance")
)

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now().UTC()
}

type Service struct {
	markets  store.MarketRepository
	balances store.BalanceRepository
	orders   store.PortfolioRepository
	prices   PriceProvider
	clock    Clock
}

type PriceProvider interface {
	GetSpotPrice(ctx context.Context, marketSymbol string) (float64, error)
}

func NewService(markets store.MarketRepository, balances store.BalanceRepository, orders store.PortfolioRepository, prices PriceProvider) *Service {
	return &Service{
		markets:  markets,
		balances: balances,
		orders:   orders,
		prices:   prices,
		clock:    realClock{},
	}
}

type PlaceMarketBuyInput struct {
	UserID       string
	WalletID     string
	MarketSymbol string
	QuoteAmount  float64
}

func (s *Service) PlaceMarketBuy(ctx context.Context, input PlaceMarketBuyInput) (domain.Order, error) {
	if input.QuoteAmount <= 0 {
		return domain.Order{}, ErrQuoteAmountTooLow
	}

	market, err := s.markets.GetBySymbol(ctx, input.MarketSymbol)
	if err != nil {
		return domain.Order{}, fmt.Errorf("get market: %w", err)
	}

	if input.QuoteAmount < market.MinNotional {
		return domain.Order{}, fmt.Errorf("quote amount below market minimum: %.2f", market.MinNotional)
	}

	quoteBalance, err := s.balances.GetByWalletAndAsset(ctx, input.WalletID, market.QuoteAsset)
	if err != nil {
		return domain.Order{}, fmt.Errorf("get balance: %w", err)
	}

	if quoteBalance.Available < input.QuoteAmount {
		return domain.Order{}, ErrInsufficientFunds
	}

	currentPrice, err := s.prices.GetSpotPrice(ctx, market.Symbol)
	if err != nil {
		return domain.Order{}, fmt.Errorf("get current price: %w", err)
	}

	if currentPrice <= 0 {
		return domain.Order{}, ErrCurrentPriceUnavailable
	}

	order := domain.Order{
		ID:            uuid.NewString(),
		UserID:        input.UserID,
		WalletID:      input.WalletID,
		MarketID:      market.ID,
		MarketSymbol:  market.Symbol,
		BaseAsset:     market.BaseAsset,
		QuoteAsset:    market.QuoteAsset,
		QuoteAmount:   input.QuoteAmount,
		BaseQuantity:  input.QuoteAmount / currentPrice,
		ExpectedPrice: currentPrice,
		Side:          domain.OrderSideBuy,
		Type:          domain.OrderTypeMarket,
		Status:        domain.OrderStatusFilled,
		CreatedAt:     s.clock.Now(),
		ExecutedAt:    s.clock.Now(),
	}

	createdOrder, err := s.orders.ApplyMarketBuy(ctx, order)
	if err != nil {
		return domain.Order{}, fmt.Errorf("create order: %w", err)
	}

	return createdOrder, nil
}
