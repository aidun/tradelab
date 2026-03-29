package strategy

import (
	"context"
	"testing"
	"time"

	"github.com/aidun/tradelab/backend/internal/domain"
	orderservice "github.com/aidun/tradelab/backend/internal/service/order"
)

type fakeMarketRepository struct {
	market domain.Market
}

func (f fakeMarketRepository) List(context.Context) ([]domain.Market, error) {
	return []domain.Market{f.market}, nil
}

func (f fakeMarketRepository) GetBySymbol(context.Context, string) (domain.Market, error) {
	return f.market, nil
}

type fakeStrategyRepository struct {
	strategies []domain.Strategy
	recorded   []domain.StrategyRun
	reasons    []string
	reference  float64
}

func (f *fakeStrategyRepository) ListByWallet(context.Context, string, string) ([]domain.Strategy, error) {
	return f.strategies, nil
}

func (f *fakeStrategyRepository) UpsertForWalletMarket(_ context.Context, strategy domain.Strategy) (domain.Strategy, error) {
	if strategy.ID == "" {
		strategy.ID = "strategy-1"
	}
	return strategy, nil
}

func (f *fakeStrategyRepository) GetByIDForWallet(context.Context, string, string) (domain.Strategy, error) {
	return f.strategies[0], nil
}

func (f *fakeStrategyRepository) ClaimActiveStrategies(context.Context, string, int, time.Time) ([]domain.Strategy, error) {
	return f.strategies, nil
}

func (f *fakeStrategyRepository) RecordEvaluation(_ context.Context, _ domain.Strategy, run domain.StrategyRun, nextReferencePrice float64) error {
	f.recorded = append(f.recorded, run)
	f.reasons = append(f.reasons, run.Reason)
	f.reference = nextReferencePrice
	return nil
}

func (f *fakeStrategyRepository) RecordEvaluationError(_ context.Context, _ domain.Strategy, run domain.StrategyRun) error {
	f.recorded = append(f.recorded, run)
	f.reasons = append(f.reasons, run.Reason)
	return nil
}

func (f *fakeStrategyRepository) RecordLifecycleActivity(context.Context, domain.Strategy, string, string) error {
	return nil
}

type fakePriceProvider struct {
	price float64
}

func (f fakePriceProvider) GetSpotPrice(context.Context, string) (float64, error) {
	return f.price, nil
}

type fakePortfolioProvider struct {
	summary domain.PortfolioSummary
}

func (f fakePortfolioProvider) GetSummary(context.Context, string, domain.AccountingMode) (domain.PortfolioSummary, error) {
	return f.summary, nil
}

type fakeOrderExecutor struct {
	inputs []orderservice.PlaceMarketOrderInput
}

func (f *fakeOrderExecutor) PlaceMarketOrder(_ context.Context, input orderservice.PlaceMarketOrderInput) (domain.Order, error) {
	f.inputs = append(f.inputs, input)
	return domain.Order{ID: "order-1", Side: input.Side}, nil
}

func TestRunOnceExecutesDipBuy(t *testing.T) {
	repo := &fakeStrategyRepository{
		strategies: []domain.Strategy{
			{
				ID:             "strategy-1",
				UserID:         "user-1",
				WalletID:       "wallet-1",
				MarketID:       "market-1",
				MarketSymbol:   "XRP/USDT",
				Status:         domain.StrategyStatusActive,
				ReferencePrice: 1.00,
				Config: domain.StrategyConfig{
					DipBuy: domain.DipBuyRule{Enabled: true, DipPercent: 5, SpendQuoteAmount: 100},
				},
			},
		},
	}
	orders := &fakeOrderExecutor{}
	service := NewService(
		fakeMarketRepository{market: domain.Market{ID: "market-1", Symbol: "XRP/USDT", BaseAsset: "XRP", QuoteAsset: "USDT"}},
		repo,
		fakePriceProvider{price: 0.94},
		fakePortfolioProvider{},
		orders,
		nil,
	)

	if err := service.RunOnce(context.Background(), 10); err != nil {
		t.Fatalf("expected run once to succeed, got %v", err)
	}

	if len(orders.inputs) != 1 {
		t.Fatalf("expected one strategy order, got %d", len(orders.inputs))
	}
	if orders.inputs[0].Side != domain.OrderSideBuy {
		t.Fatalf("expected buy order, got %s", orders.inputs[0].Side)
	}
	if orders.inputs[0].OrderSource != domain.OrderSourceStrategy {
		t.Fatalf("expected strategy order source, got %s", orders.inputs[0].OrderSource)
	}
	if orders.inputs[0].StrategyID != "strategy-1" {
		t.Fatalf("expected strategy id to be forwarded, got %q", orders.inputs[0].StrategyID)
	}
	if len(repo.recorded) != 1 || repo.recorded[0].Outcome != domain.StrategyOutcomeExecuted {
		t.Fatalf("expected executed strategy run, got %#v", repo.recorded)
	}
}

func TestRunOnceExecutesTakeProfitSell(t *testing.T) {
	repo := &fakeStrategyRepository{
		strategies: []domain.Strategy{
			{
				ID:             "strategy-1",
				UserID:         "user-1",
				WalletID:       "wallet-1",
				MarketID:       "market-1",
				MarketSymbol:   "XRP/USDT",
				Status:         domain.StrategyStatusActive,
				ReferencePrice: 0.70,
				Config: domain.StrategyConfig{
					TakeProfit: domain.TakeProfitRule{Enabled: true, TriggerPercent: 10},
				},
			},
		},
	}
	orders := &fakeOrderExecutor{}
	service := NewService(
		fakeMarketRepository{market: domain.Market{ID: "market-1", Symbol: "XRP/USDT", BaseAsset: "XRP", QuoteAsset: "USDT"}},
		repo,
		fakePriceProvider{price: 0.80},
		fakePortfolioProvider{
			summary: domain.PortfolioSummary{
				Positions: []domain.Position{
					{MarketSymbol: "XRP/USDT", OpenQuantity: 50, EntryPriceAvg: 0.70},
				},
			},
		},
		orders,
		nil,
	)

	if err := service.RunOnce(context.Background(), 10); err != nil {
		t.Fatalf("expected run once to succeed, got %v", err)
	}

	if len(orders.inputs) != 1 {
		t.Fatalf("expected one strategy sell order, got %d", len(orders.inputs))
	}
	if orders.inputs[0].Side != domain.OrderSideSell {
		t.Fatalf("expected sell order, got %s", orders.inputs[0].Side)
	}
	if orders.inputs[0].BaseQuantity != 50 {
		t.Fatalf("expected full-position sell, got %f", orders.inputs[0].BaseQuantity)
	}
	if orders.inputs[0].StrategyID != "strategy-1" {
		t.Fatalf("expected strategy id to be forwarded, got %q", orders.inputs[0].StrategyID)
	}
}

func TestRunOnceSkipsWhenNoRuleMatches(t *testing.T) {
	repo := &fakeStrategyRepository{
		strategies: []domain.Strategy{
			{
				ID:             "strategy-1",
				UserID:         "user-1",
				WalletID:       "wallet-1",
				MarketID:       "market-1",
				MarketSymbol:   "XRP/USDT",
				Status:         domain.StrategyStatusActive,
				ReferencePrice: 1.00,
				Config: domain.StrategyConfig{
					DipBuy: domain.DipBuyRule{Enabled: true, DipPercent: 5, SpendQuoteAmount: 100},
				},
			},
		},
	}
	orders := &fakeOrderExecutor{}
	service := NewService(
		fakeMarketRepository{market: domain.Market{ID: "market-1", Symbol: "XRP/USDT", BaseAsset: "XRP", QuoteAsset: "USDT"}},
		repo,
		fakePriceProvider{price: 0.98},
		fakePortfolioProvider{},
		orders,
		nil,
	)

	if err := service.RunOnce(context.Background(), 10); err != nil {
		t.Fatalf("expected run once to succeed, got %v", err)
	}

	if len(orders.inputs) != 0 {
		t.Fatalf("expected no order, got %d", len(orders.inputs))
	}
	if len(repo.recorded) != 1 || repo.recorded[0].Outcome != domain.StrategyOutcomeSkipped {
		t.Fatalf("expected skipped strategy run, got %#v", repo.recorded)
	}
}
