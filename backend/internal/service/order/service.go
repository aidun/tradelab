package order

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/aidun/tradelab/backend/internal/domain"
	"github.com/aidun/tradelab/backend/internal/store"
)

var (
	ErrQuoteAmountTooLow       = errors.New("quote amount must be greater than zero")
	ErrBaseQuantityTooLow      = errors.New("base quantity must be greater than zero")
	ErrCurrentPriceUnavailable = errors.New("current market price is unavailable")
	ErrInsufficientFunds       = errors.New("insufficient quote balance")
	ErrInsufficientPosition    = errors.New("insufficient base balance")
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
	logger   *slog.Logger
	clock    Clock
}

type PriceProvider interface {
	GetSpotPrice(ctx context.Context, marketSymbol string) (float64, error)
}

func NewService(markets store.MarketRepository, balances store.BalanceRepository, orders store.PortfolioRepository, prices PriceProvider, logger *slog.Logger) *Service {
	return &Service{
		markets:  markets,
		balances: balances,
		orders:   orders,
		prices:   prices,
		logger:   logger,
		clock:    realClock{},
	}
}

type PlaceMarketOrderInput struct {
	UserID       string
	WalletID     string
	MarketSymbol string
	Side         domain.OrderSide
	QuoteAmount  float64
	BaseQuantity float64
}

func (s *Service) PlaceMarketOrder(ctx context.Context, input PlaceMarketOrderInput) (domain.Order, error) {
	s.logInfo("place_market_order.attempt", "wallet_id", input.WalletID, "market_symbol", input.MarketSymbol, "side", input.Side, "quote_amount", input.QuoteAmount, "base_quantity", input.BaseQuantity)

	market, err := s.markets.GetBySymbol(ctx, input.MarketSymbol)
	if err != nil {
		s.logError("place_market_order.market_lookup_failed", err, "wallet_id", input.WalletID, "market_symbol", input.MarketSymbol)
		return domain.Order{}, fmt.Errorf("get market: %w", err)
	}

	currentPrice, err := s.prices.GetSpotPrice(ctx, market.Symbol)
	if err != nil {
		s.logError("place_market_order.price_lookup_failed", err, "wallet_id", input.WalletID, "market_symbol", market.Symbol)
		return domain.Order{}, fmt.Errorf("get current price: %w", err)
	}

	if currentPrice <= 0 {
		s.logInfo("place_market_order.validation_failed", "wallet_id", input.WalletID, "market_symbol", market.Symbol, "reason", "price_unavailable")
		return domain.Order{}, ErrCurrentPriceUnavailable
	}

	switch input.Side {
	case domain.OrderSideBuy:
		if input.QuoteAmount <= 0 {
			s.logInfo("place_market_order.validation_failed", "wallet_id", input.WalletID, "reason", "quote_amount_too_low")
			return domain.Order{}, ErrQuoteAmountTooLow
		}

		if input.QuoteAmount < market.MinNotional {
			s.logInfo("place_market_order.validation_failed", "wallet_id", input.WalletID, "market_symbol", market.Symbol, "reason", "below_min_notional", "min_notional", market.MinNotional)
			return domain.Order{}, fmt.Errorf("quote amount below market minimum: %.2f", market.MinNotional)
		}

		quoteBalance, err := s.balances.GetByWalletAndAsset(ctx, input.WalletID, market.QuoteAsset)
		if err != nil {
			s.logError("place_market_order.balance_lookup_failed", err, "wallet_id", input.WalletID, "market_symbol", market.Symbol, "asset", market.QuoteAsset)
			return domain.Order{}, fmt.Errorf("get balance: %w", err)
		}

		if quoteBalance.Available < input.QuoteAmount {
			s.logInfo("place_market_order.insufficient_funds", "wallet_id", input.WalletID, "market_symbol", market.Symbol, "available_balance", quoteBalance.Available)
			return domain.Order{}, ErrInsufficientFunds
		}

		input.BaseQuantity = input.QuoteAmount / currentPrice
	case domain.OrderSideSell:
		if input.BaseQuantity <= 0 {
			s.logInfo("place_market_order.validation_failed", "wallet_id", input.WalletID, "reason", "base_quantity_too_low")
			return domain.Order{}, ErrBaseQuantityTooLow
		}

		baseBalance, err := s.balances.GetByWalletAndAsset(ctx, input.WalletID, market.BaseAsset)
		if err != nil {
			s.logError("place_market_order.balance_lookup_failed", err, "wallet_id", input.WalletID, "market_symbol", market.Symbol, "asset", market.BaseAsset)
			return domain.Order{}, fmt.Errorf("get balance: %w", err)
		}

		if baseBalance.Available < input.BaseQuantity {
			s.logInfo("place_market_order.insufficient_position", "wallet_id", input.WalletID, "market_symbol", market.Symbol, "available_balance", baseBalance.Available)
			return domain.Order{}, ErrInsufficientPosition
		}

		input.QuoteAmount = input.BaseQuantity * currentPrice
	default:
		return domain.Order{}, fmt.Errorf("unsupported order side: %s", input.Side)
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
		BaseQuantity:  input.BaseQuantity,
		ExpectedPrice: currentPrice,
		Side:          input.Side,
		Type:          domain.OrderTypeMarket,
		Status:        domain.OrderStatusFilled,
		CreatedAt:     s.clock.Now(),
		ExecutedAt:    s.clock.Now(),
	}

	var createdOrder domain.Order
	switch input.Side {
	case domain.OrderSideBuy:
		createdOrder, err = s.orders.ApplyMarketBuy(ctx, order)
	case domain.OrderSideSell:
		createdOrder, err = s.orders.ApplyMarketSell(ctx, order)
	}
	if err != nil {
		s.logError("place_market_order.apply_failed", err, "wallet_id", input.WalletID, "market_symbol", market.Symbol, "side", input.Side)
		return domain.Order{}, fmt.Errorf("create order: %w", err)
	}

	s.logInfo("place_market_order.success", "wallet_id", input.WalletID, "market_symbol", market.Symbol, "order_id", createdOrder.ID, "executed_price", createdOrder.ExpectedPrice, "side", createdOrder.Side)
	return createdOrder, nil
}

func (s *Service) logInfo(operation string, args ...any) {
	if s.logger == nil {
		return
	}

	s.logger.Info(operation, append([]any{"operation", operation}, args...)...)
}

func (s *Service) logError(operation string, err error, args ...any) {
	if s.logger == nil {
		return
	}

	s.logger.Error(operation, append([]any{"operation", operation, "error", err}, args...)...)
}
