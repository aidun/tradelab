package tradingcalc

import (
	"testing"
	"time"

	"github.com/aidun/tradelab/backend/internal/domain"
)

func TestAnalyzeOrdersComputesAverageCostRealizedPnL(t *testing.T) {
	result := AnalyzeOrders(sampleOrders(), domain.AccountingModeAverageCost, map[string]float64{"XRP/USDT": 1.0})

	if len(result.Positions) != 1 {
		t.Fatalf("expected 1 open position, got %d", len(result.Positions))
	}

	if result.Realized <= 0 {
		t.Fatalf("expected positive realized pnl, got %f", result.Realized)
	}

	if result.Positions[0].UnrealizedPnL <= 0 {
		t.Fatalf("expected positive unrealized pnl, got %f", result.Positions[0].UnrealizedPnL)
	}
}

func TestAnalyzeOrdersComputesFIFOPnLDifferently(t *testing.T) {
	averageResult := AnalyzeOrders(sampleOrders(), domain.AccountingModeAverageCost, map[string]float64{"XRP/USDT": 1.0})
	fifoResult := AnalyzeOrders(sampleOrders(), domain.AccountingModeFIFO, map[string]float64{"XRP/USDT": 1.0})

	if averageResult.Realized == fifoResult.Realized {
		t.Fatal("expected realized pnl to differ between average cost and fifo")
	}
}

func TestAnalyzeOrdersUsesHybridAverageValuationAndFIFORealized(t *testing.T) {
	hybridResult := AnalyzeOrders(sampleOrders(), domain.AccountingModeHybrid, map[string]float64{"XRP/USDT": 1.0})
	averageResult := AnalyzeOrders(sampleOrders(), domain.AccountingModeAverageCost, map[string]float64{"XRP/USDT": 1.0})
	fifoResult := AnalyzeOrders(sampleOrders(), domain.AccountingModeFIFO, map[string]float64{"XRP/USDT": 1.0})

	if hybridResult.Realized != fifoResult.Realized {
		t.Fatalf("expected hybrid realized pnl to match fifo, got %f and %f", hybridResult.Realized, fifoResult.Realized)
	}

	if hybridResult.Positions[0].UnrealizedPnL != averageResult.Positions[0].UnrealizedPnL {
		t.Fatalf("expected hybrid unrealized pnl to match average cost, got %f and %f", hybridResult.Positions[0].UnrealizedPnL, averageResult.Positions[0].UnrealizedPnL)
	}
}

func sampleOrders() []domain.Order {
	now := time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)
	return []domain.Order{
		{
			ID:            "order-1",
			MarketID:      "market-1",
			MarketSymbol:  "XRP/USDT",
			BaseAsset:     "XRP",
			QuoteAsset:    "USDT",
			Side:          domain.OrderSideBuy,
			BaseQuantity:  100,
			QuoteAmount:   50,
			ExpectedPrice: 0.5,
			CreatedAt:     now,
		},
		{
			ID:            "order-2",
			MarketID:      "market-1",
			MarketSymbol:  "XRP/USDT",
			BaseAsset:     "XRP",
			QuoteAsset:    "USDT",
			Side:          domain.OrderSideBuy,
			BaseQuantity:  100,
			QuoteAmount:   80,
			ExpectedPrice: 0.8,
			CreatedAt:     now.Add(time.Minute),
		},
		{
			ID:            "order-3",
			MarketID:      "market-1",
			MarketSymbol:  "XRP/USDT",
			BaseAsset:     "XRP",
			QuoteAsset:    "USDT",
			Side:          domain.OrderSideSell,
			BaseQuantity:  100,
			QuoteAmount:   90,
			ExpectedPrice: 0.9,
			CreatedAt:     now.Add(2 * time.Minute),
		},
	}
}
