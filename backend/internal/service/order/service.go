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
	ErrQuoteAmountTooLow = errors.New("quote amount must be greater than zero")
	ErrExpectedPriceLow  = errors.New("expected price must be greater than zero")
	ErrInsufficientFunds = errors.New("insufficient quote balance")
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
	clock    Clock
}

func NewService(markets store.MarketRepository, balances store.BalanceRepository, orders store.PortfolioRepository) *Service {
	return &Service{
		markets:  markets,
		balances: balances,
		orders:   orders,
		clock:    realClock{},
	}
}

type PlaceMarketBuyInput struct {
	UserID        string
	WalletID      string
	MarketSymbol  string
	QuoteAmount   float64
	ExpectedPrice float64
}

func (s *Service) PlaceMarketBuy(ctx context.Context, input PlaceMarketBuyInput) (domain.Order, error) {
	if input.QuoteAmount <= 0 {
		return domain.Order{}, ErrQuoteAmountTooLow
	}

	if input.ExpectedPrice <= 0 {
		return domain.Order{}, ErrExpectedPriceLow
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

	order := domain.Order{
		ID:            uuid.NewString(),
		UserID:        input.UserID,
		WalletID:      input.WalletID,
		MarketID:      market.ID,
		MarketSymbol:  market.Symbol,
		BaseAsset:     market.BaseAsset,
		QuoteAsset:    market.QuoteAsset,
		QuoteAmount:   input.QuoteAmount,
		BaseQuantity:  input.QuoteAmount / input.ExpectedPrice,
		ExpectedPrice: input.ExpectedPrice,
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
