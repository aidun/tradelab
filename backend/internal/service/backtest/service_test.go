package backtest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aidun/tradelab/backend/internal/domain"
)

func TestRunExecutesDipBuyAndTakeProfitDeterministically(t *testing.T) {
	service := NewService(fakeMarketRepo{}, fakeCandleProvider{
		candles: []domain.Candle{
			candleAt(0.70, 0),
			candleAt(0.66, 1),
			candleAt(0.80, 2),
		},
	})

	run, err := service.Run(context.Background(), RunInput{
		MarketSymbol: "XRP/USDT",
		Interval:     "1h",
		StartTime:    time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		EndTime:      time.Date(2026, 3, 1, 3, 0, 0, 0, time.UTC),
		Config: domain.StrategyConfig{
			DipBuy:     domain.DipBuyRule{Enabled: true, DipPercent: 5, SpendQuoteAmount: 100},
			TakeProfit: domain.TakeProfitRule{Enabled: true, TriggerPercent: 10},
			StopLoss:   domain.StopLossRule{Enabled: true, TriggerPercent: 3},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(run.Orders) != 2 {
		t.Fatalf("expected 2 simulated orders, got %d", len(run.Orders))
	}

	if run.Orders[0].Side != domain.OrderSideSell {
		t.Fatalf("expected newest annotated order to be a sell, got %s", run.Orders[0].Side)
	}

	if run.Summary.TradeCount != 2 {
		t.Fatalf("expected 2 trades, got %d", run.Summary.TradeCount)
	}

	if run.FinalEquity <= run.InitialCash {
		t.Fatalf("expected profitable backtest, got initial %f final %f", run.InitialCash, run.FinalEquity)
	}
}

func TestRunRejectsInvalidDateRange(t *testing.T) {
	service := NewService(fakeMarketRepo{}, fakeCandleProvider{})

	_, err := service.Run(context.Background(), RunInput{
		MarketSymbol: "XRP/USDT",
		StartTime:    time.Date(2026, 3, 1, 2, 0, 0, 0, time.UTC),
		EndTime:      time.Date(2026, 3, 1, 1, 0, 0, 0, time.UTC),
		Config: domain.StrategyConfig{
			DipBuy: domain.DipBuyRule{Enabled: true, DipPercent: 5, SpendQuoteAmount: 100},
		},
	})
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("expected invalid range error, got %v", err)
	}
}

func TestRunReturnsNoHistoricalDataError(t *testing.T) {
	service := NewService(fakeMarketRepo{}, fakeCandleProvider{})

	_, err := service.Run(context.Background(), RunInput{
		MarketSymbol: "XRP/USDT",
		StartTime:    time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		EndTime:      time.Date(2026, 3, 1, 1, 0, 0, 0, time.UTC),
		Config: domain.StrategyConfig{
			DipBuy: domain.DipBuyRule{Enabled: true, DipPercent: 5, SpendQuoteAmount: 100},
		},
	})
	if !errors.Is(err, ErrNoHistoricalData) {
		t.Fatalf("expected no data error, got %v", err)
	}
}

type fakeMarketRepo struct{}

func (fakeMarketRepo) List(context.Context) ([]domain.Market, error) { return nil, nil }
func (fakeMarketRepo) GetBySymbol(context.Context, string) (domain.Market, error) {
	return domain.Market{ID: "market-1", Symbol: "XRP/USDT", BaseAsset: "XRP", QuoteAsset: "USDT"}, nil
}

type fakeCandleProvider struct {
	candles []domain.Candle
	err     error
}

func (f fakeCandleProvider) ListHistoricalCandles(context.Context, string, string, time.Time, time.Time) ([]domain.Candle, error) {
	return f.candles, f.err
}

func candleAt(closePrice float64, offset int) domain.Candle {
	openTime := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(offset) * time.Hour)
	return domain.Candle{
		OpenTime:   openTime,
		CloseTime:  openTime.Add(time.Hour - time.Second),
		OpenPrice:  closePrice,
		HighPrice:  closePrice,
		LowPrice:   closePrice,
		ClosePrice: closePrice,
	}
}
