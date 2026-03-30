package backtest

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/aidun/tradelab/backend/internal/domain"
	"github.com/aidun/tradelab/backend/internal/service/tradingcalc"
	"github.com/aidun/tradelab/backend/internal/store"
)

const defaultInitialCash = 10_000.0

var (
	ErrInvalidRange      = errors.New("backtest start time must be before end time")
	ErrNoHistoricalData  = errors.New("no historical candles available for that range")
	ErrUnsupportedMarket = errors.New("market unavailable for backtest")
)

type HistoricalCandleProvider interface {
	ListHistoricalCandles(ctx context.Context, marketSymbol string, interval string, start time.Time, end time.Time) ([]domain.Candle, error)
}

type Service struct {
	markets store.MarketRepository
	candles HistoricalCandleProvider
}

type RunInput struct {
	MarketSymbol string
	Interval     string
	StartTime    time.Time
	EndTime      time.Time
	Config       domain.StrategyConfig
}

func NewService(markets store.MarketRepository, candles HistoricalCandleProvider) *Service {
	return &Service{markets: markets, candles: candles}
}

func (s *Service) Run(ctx context.Context, input RunInput) (domain.BacktestRun, error) {
	if !input.StartTime.Before(input.EndTime) {
		return domain.BacktestRun{}, ErrInvalidRange
	}

	if input.Interval == "" {
		input.Interval = "1h"
	}

	market, err := s.markets.GetBySymbol(ctx, input.MarketSymbol)
	if err != nil {
		return domain.BacktestRun{}, fmt.Errorf("%w: %v", ErrUnsupportedMarket, err)
	}

	candles, err := s.candles.ListHistoricalCandles(ctx, market.Symbol, input.Interval, input.StartTime.UTC(), input.EndTime.UTC())
	if err != nil {
		return domain.BacktestRun{}, err
	}
	if len(candles) == 0 {
		return domain.BacktestRun{}, ErrNoHistoricalData
	}

	initialCash := defaultInitialCash
	cashBalance := initialCash
	openQuantity := 0.0
	averageEntryPrice := 0.0
	referencePrice := 0.0
	peakEquity := initialCash
	maxDrawdownPct := 0.0
	orders := make([]domain.Order, 0)
	equityCurve := make([]domain.BacktestEquityPoint, 0, len(candles))

	for _, candle := range candles {
		price := candle.ClosePrice
		if price <= 0 {
			continue
		}

		if referencePrice == 0 || price > referencePrice {
			referencePrice = price
		}

		switch {
		case openQuantity > 0 && input.Config.StopLoss.Enabled && price <= averageEntryPrice*(1-input.Config.StopLoss.TriggerPercent/100):
			quoteAmount := openQuantity * price
			orders = append(orders, syntheticOrder(market, domain.OrderSideSell, openQuantity, quoteAmount, price, candle.CloseTime))
			cashBalance += quoteAmount
			openQuantity = 0
			averageEntryPrice = 0
			referencePrice = price
		case openQuantity > 0 && input.Config.TakeProfit.Enabled && price >= averageEntryPrice*(1+input.Config.TakeProfit.TriggerPercent/100):
			quoteAmount := openQuantity * price
			orders = append(orders, syntheticOrder(market, domain.OrderSideSell, openQuantity, quoteAmount, price, candle.CloseTime))
			cashBalance += quoteAmount
			openQuantity = 0
			averageEntryPrice = 0
			referencePrice = price
		case input.Config.DipBuy.Enabled && cashBalance >= input.Config.DipBuy.SpendQuoteAmount && price <= referencePrice*(1-input.Config.DipBuy.DipPercent/100):
			baseQuantity := input.Config.DipBuy.SpendQuoteAmount / price
			orders = append(orders, syntheticOrder(market, domain.OrderSideBuy, baseQuantity, input.Config.DipBuy.SpendQuoteAmount, price, candle.CloseTime))
			totalCost := (openQuantity * averageEntryPrice) + input.Config.DipBuy.SpendQuoteAmount
			openQuantity += baseQuantity
			if openQuantity > 0 {
				averageEntryPrice = totalCost / openQuantity
			}
			cashBalance -= input.Config.DipBuy.SpendQuoteAmount
			referencePrice = price
		}

		positionValue := openQuantity * price
		totalEquity := cashBalance + positionValue
		if totalEquity > peakEquity {
			peakEquity = totalEquity
		}
		drawdownPct := 0.0
		if peakEquity > 0 {
			drawdownPct = (peakEquity - totalEquity) / peakEquity * 100
		}
		if drawdownPct > maxDrawdownPct {
			maxDrawdownPct = drawdownPct
		}

		equityCurve = append(equityCurve, domain.BacktestEquityPoint{
			Time:          candle.CloseTime,
			Price:         price,
			CashBalance:   cashBalance,
			OpenQuantity:  openQuantity,
			PositionValue: positionValue,
			TotalEquity:   totalEquity,
			DrawdownPct:   drawdownPct,
		})
	}

	lastPrice := candles[len(candles)-1].ClosePrice
	analysis := tradingcalc.AnalyzeOrders(orders, domain.AccountingModeAverageCost, map[string]float64{market.Symbol: lastPrice})
	finalPositionVal := openQuantity * lastPrice
	finalEquity := cashBalance + finalPositionVal
	summary := summarizeRun(initialCash, finalEquity, analysis.Orders, maxDrawdownPct)

	return domain.BacktestRun{
		MarketSymbol:     market.Symbol,
		BaseAsset:        market.BaseAsset,
		QuoteAsset:       market.QuoteAsset,
		Interval:         input.Interval,
		StartTime:        input.StartTime.UTC(),
		EndTime:          input.EndTime.UTC(),
		InitialCash:      initialCash,
		FinalCash:        cashBalance,
		FinalPositionQty: openQuantity,
		FinalPositionVal: finalPositionVal,
		FinalEquity:      finalEquity,
		Orders:           analysis.Orders,
		EquityCurve:      equityCurve,
		Summary:          summary,
	}, nil
}

func syntheticOrder(market domain.Market, side domain.OrderSide, baseQuantity float64, quoteAmount float64, price float64, executedAt time.Time) domain.Order {
	return domain.Order{
		ID:            uuid.NewString(),
		MarketID:      market.ID,
		MarketSymbol:  market.Symbol,
		BaseAsset:     market.BaseAsset,
		QuoteAsset:    market.QuoteAsset,
		OrderSource:   domain.OrderSourceStrategy,
		Side:          side,
		Type:          domain.OrderTypeMarket,
		Status:        domain.OrderStatusFilled,
		BaseQuantity:  baseQuantity,
		QuoteAmount:   quoteAmount,
		ExpectedPrice: price,
		CreatedAt:     executedAt.UTC(),
		ExecutedAt:    executedAt.UTC(),
	}
}

func summarizeRun(initialCash float64, finalEquity float64, orders []domain.Order, maxDrawdownPct float64) domain.BacktestSummary {
	sellCount := 0
	winningTradeCount := 0
	for _, order := range orders {
		if order.Side != domain.OrderSideSell {
			continue
		}
		sellCount++
		if order.RealizedPnL > 0 {
			winningTradeCount++
		}
	}

	hitRatePercent := 0.0
	if sellCount > 0 {
		hitRatePercent = float64(winningTradeCount) / float64(sellCount) * 100
	}

	returnPct := 0.0
	if initialCash > 0 {
		returnPct = (finalEquity - initialCash) / initialCash * 100
	}

	return domain.BacktestSummary{
		ReturnPercent:      returnPct,
		TradeCount:         len(orders),
		SellCount:          sellCount,
		WinningTradeCount:  winningTradeCount,
		HitRatePercent:     hitRatePercent,
		MaxDrawdownPercent: maxDrawdownPct,
	}
}
