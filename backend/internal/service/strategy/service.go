package strategy

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/aidun/tradelab/backend/internal/domain"
	orderservice "github.com/aidun/tradelab/backend/internal/service/order"
	"github.com/aidun/tradelab/backend/internal/store"
)

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now().UTC()
}

type PriceProvider interface {
	GetSpotPrice(ctx context.Context, marketSymbol string) (float64, error)
}

type PortfolioSummaryProvider interface {
	GetSummary(ctx context.Context, walletID string, mode domain.AccountingMode) (domain.PortfolioSummary, error)
}

type OrderExecutor interface {
	PlaceMarketOrder(ctx context.Context, input orderservice.PlaceMarketOrderInput) (domain.Order, error)
}

type Service struct {
	markets    store.MarketRepository
	repository store.StrategyRepository
	prices     PriceProvider
	portfolios PortfolioSummaryProvider
	orders     OrderExecutor
	logger     *slog.Logger
	clock      Clock
	claimStale time.Duration
}

func NewService(markets store.MarketRepository, repository store.StrategyRepository, prices PriceProvider, portfolios PortfolioSummaryProvider, orders OrderExecutor, logger *slog.Logger) *Service {
	return &Service{
		markets:    markets,
		repository: repository,
		prices:     prices,
		portfolios: portfolios,
		orders:     orders,
		logger:     logger,
		clock:      realClock{},
		claimStale: 2 * time.Minute,
	}
}

type UpsertStrategyInput struct {
	UserID       string
	WalletID     string
	MarketSymbol string
	Status       domain.StrategyStatus
	Config       domain.StrategyConfig
}

type PatchStrategyInput struct {
	UserID   string
	WalletID string
	ID       string
	Status   domain.StrategyStatus
	Config   domain.StrategyConfig
}

func (s *Service) ListStrategies(ctx context.Context, walletID string, marketSymbol string) ([]domain.Strategy, error) {
	items, err := s.repository.ListByWallet(ctx, walletID, marketSymbol)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []domain.Strategy{}, nil
	}
	return items, nil
}

func (s *Service) UpsertStrategy(ctx context.Context, input UpsertStrategyInput) (domain.Strategy, error) {
	if err := validateConfig(input.Config); err != nil {
		return domain.Strategy{}, err
	}

	market, err := s.markets.GetBySymbol(ctx, input.MarketSymbol)
	if err != nil {
		return domain.Strategy{}, fmt.Errorf("get market: %w", err)
	}

	existing, err := s.repository.ListByWallet(ctx, input.WalletID, input.MarketSymbol)
	if err != nil {
		return domain.Strategy{}, err
	}

	status := normalizeStatus(input.Status)
	strategy := domain.Strategy{
		ID:           firstStrategyID(existing),
		UserID:       input.UserID,
		WalletID:     input.WalletID,
		MarketID:     market.ID,
		MarketSymbol: market.Symbol,
		Status:       status,
		Config:       input.Config,
	}

	saved, err := s.repository.UpsertForWalletMarket(ctx, strategy)
	if err != nil {
		return domain.Strategy{}, err
	}

	title, message := lifecycleMessage(existing, saved, false)
	if title != "" {
		_ = s.repository.RecordLifecycleActivity(ctx, saved, title, message)
	}

	return saved, nil
}

func (s *Service) PatchStrategy(ctx context.Context, input PatchStrategyInput) (domain.Strategy, error) {
	if err := validateConfig(input.Config); err != nil {
		return domain.Strategy{}, err
	}

	current, err := s.repository.GetByIDForWallet(ctx, input.WalletID, input.ID)
	if err != nil {
		return domain.Strategy{}, err
	}

	current.Config = input.Config
	current.Status = normalizeStatus(input.Status)

	saved, err := s.repository.UpsertForWalletMarket(ctx, current)
	if err != nil {
		return domain.Strategy{}, err
	}

	title, message := lifecycleMessage([]domain.Strategy{current}, saved, true)
	if title != "" {
		_ = s.repository.RecordLifecycleActivity(ctx, saved, title, message)
	}

	return saved, nil
}

func (s *Service) RunOnce(ctx context.Context, limit int) error {
	claimToken := uuid.NewString()
	strategies, err := s.repository.ClaimActiveStrategies(ctx, claimToken, limit, s.clock.Now().Add(-s.claimStale))
	if err != nil {
		return err
	}

	for _, strategy := range strategies {
		s.evaluateStrategy(ctx, strategy)
	}

	return nil
}

func (s *Service) evaluateStrategy(ctx context.Context, strategy domain.Strategy) {
	startedAt := s.clock.Now()

	price, err := s.prices.GetSpotPrice(ctx, strategy.MarketSymbol)
	if err != nil {
		s.recordEvaluationError(ctx, strategy, startedAt, fmt.Sprintf("Market price unavailable for %s.", strategy.MarketSymbol), map[string]any{
			"marketSymbol": strategy.MarketSymbol,
		})
		return
	}

	portfolio, err := s.portfolios.GetSummary(ctx, strategy.WalletID, domain.AccountingModeAverageCost)
	if err != nil {
		s.recordEvaluationError(ctx, strategy, startedAt, "Portfolio state unavailable during strategy evaluation.", nil)
		return
	}

	position := findPosition(portfolio, strategy.MarketSymbol)
	nextReference := strategy.ReferencePrice
	if price > nextReference {
		nextReference = price
	}
	if nextReference == 0 {
		nextReference = price
	}

	decision := domain.StrategyDecisionHold
	outcome := domain.StrategyOutcomeSkipped
	reason := "No strategy condition matched."
	details := map[string]any{
		"marketSymbol":    strategy.MarketSymbol,
		"currentPrice":    price,
		"referencePrice":  nextReference,
		"openPositionQty": 0.0,
	}
	if position != nil {
		details["openPositionQty"] = position.OpenQuantity
		details["entryPriceAvg"] = position.EntryPriceAvg
	}

	// Exit rules run before dip-buy so the engine never buys and sells the same market in one evaluation tick.
	if position != nil && position.OpenQuantity > 0 && strategy.Config.StopLoss.Enabled && price <= position.EntryPriceAvg*(1-strategy.Config.StopLoss.TriggerPercent/100) {
		decision = domain.StrategyDecisionSell
		reason = fmt.Sprintf("Stop-loss fired because %s fell to %.4f, below the %.2f%% threshold from %.4f.", strategy.MarketSymbol, price, strategy.Config.StopLoss.TriggerPercent, position.EntryPriceAvg)
		if _, err := s.orders.PlaceMarketOrder(ctx, orderservice.PlaceMarketOrderInput{
			UserID:       strategy.UserID,
			WalletID:     strategy.WalletID,
			MarketSymbol: strategy.MarketSymbol,
			Side:         domain.OrderSideSell,
			BaseQuantity: position.OpenQuantity,
			OrderSource:  domain.OrderSourceStrategy,
			StrategyID:   strategy.ID,
		}); err != nil {
			s.recordEvaluationError(ctx, strategy, startedAt, fmt.Sprintf("Stop-loss evaluation failed: %s", err.Error()), details)
			return
		}
		outcome = domain.StrategyOutcomeExecuted
		nextReference = price
	} else if position != nil && position.OpenQuantity > 0 && strategy.Config.TakeProfit.Enabled && price >= position.EntryPriceAvg*(1+strategy.Config.TakeProfit.TriggerPercent/100) {
		decision = domain.StrategyDecisionSell
		reason = fmt.Sprintf("Take-profit fired because %s reached %.4f, above the %.2f%% target from %.4f.", strategy.MarketSymbol, price, strategy.Config.TakeProfit.TriggerPercent, position.EntryPriceAvg)
		if _, err := s.orders.PlaceMarketOrder(ctx, orderservice.PlaceMarketOrderInput{
			UserID:       strategy.UserID,
			WalletID:     strategy.WalletID,
			MarketSymbol: strategy.MarketSymbol,
			Side:         domain.OrderSideSell,
			BaseQuantity: position.OpenQuantity,
			OrderSource:  domain.OrderSourceStrategy,
			StrategyID:   strategy.ID,
		}); err != nil {
			s.recordEvaluationError(ctx, strategy, startedAt, fmt.Sprintf("Take-profit evaluation failed: %s", err.Error()), details)
			return
		}
		outcome = domain.StrategyOutcomeExecuted
		nextReference = price
	} else if strategy.Config.DipBuy.Enabled {
		dipThreshold := nextReference * (1 - strategy.Config.DipBuy.DipPercent/100)
		details["dipThreshold"] = dipThreshold
		if price <= dipThreshold {
			decision = domain.StrategyDecisionBuy
			reason = fmt.Sprintf("Dip-buy fired because %s dropped to %.4f, %.2f%% below the reference price %.4f.", strategy.MarketSymbol, price, strategy.Config.DipBuy.DipPercent, nextReference)
			if _, err := s.orders.PlaceMarketOrder(ctx, orderservice.PlaceMarketOrderInput{
				UserID:       strategy.UserID,
				WalletID:     strategy.WalletID,
				MarketSymbol: strategy.MarketSymbol,
				Side:         domain.OrderSideBuy,
				QuoteAmount:  strategy.Config.DipBuy.SpendQuoteAmount,
				OrderSource:  domain.OrderSourceStrategy,
				StrategyID:   strategy.ID,
			}); err != nil {
				s.recordEvaluationError(ctx, strategy, startedAt, fmt.Sprintf("Dip-buy evaluation failed: %s", err.Error()), details)
				return
			}
			outcome = domain.StrategyOutcomeExecuted
			nextReference = price
		} else if price > strategy.ReferencePrice {
			// The highest seen reference price is the moving anchor for future dip checks until the next strategy buy resets it.
			reason = fmt.Sprintf("Reference price moved up to %.4f while waiting for a %.2f%% dip.", price, strategy.Config.DipBuy.DipPercent)
		} else {
			reason = fmt.Sprintf("Dip-buy held because %s at %.4f has not reached the %.4f trigger yet.", strategy.MarketSymbol, price, dipThreshold)
		}
	}

	detailsJSON, _ := json.Marshal(details)
	finishedAt := s.clock.Now()
	run := domain.StrategyRun{
		ID:                   uuid.NewString(),
		StrategyID:           strategy.ID,
		Decision:             decision,
		Outcome:              outcome,
		Reason:               reason,
		DetailsJSON:          string(detailsJSON),
		EvaluationDurationMS: finishedAt.Sub(startedAt).Milliseconds(),
		StartedAt:            startedAt,
		FinishedAt:           finishedAt,
	}

	if err := s.repository.RecordEvaluation(ctx, strategy, run, nextReference); err != nil {
		s.logError("strategy.record_evaluation_failed", err, "strategy_id", strategy.ID, "market_symbol", strategy.MarketSymbol)
		return
	}

	s.logInfo("strategy.evaluated", "strategy_id", strategy.ID, "market_symbol", strategy.MarketSymbol, "decision", decision, "outcome", outcome, "evaluation_duration_ms", run.EvaluationDurationMS)
}

func (s *Service) recordEvaluationError(ctx context.Context, strategy domain.Strategy, startedAt time.Time, reason string, details map[string]any) {
	detailsJSON, _ := json.Marshal(details)
	finishedAt := s.clock.Now()
	run := domain.StrategyRun{
		ID:                   uuid.NewString(),
		StrategyID:           strategy.ID,
		Decision:             domain.StrategyDecisionHold,
		Outcome:              domain.StrategyOutcomeErrored,
		Reason:               reason,
		DetailsJSON:          string(detailsJSON),
		EvaluationDurationMS: finishedAt.Sub(startedAt).Milliseconds(),
		StartedAt:            startedAt,
		FinishedAt:           finishedAt,
	}

	if err := s.repository.RecordEvaluationError(ctx, strategy, run); err != nil {
		s.logError("strategy.record_error_failed", err, "strategy_id", strategy.ID, "market_symbol", strategy.MarketSymbol)
	}
}

func validateConfig(config domain.StrategyConfig) error {
	if config.DipBuy.Enabled {
		if config.DipBuy.DipPercent <= 0 || config.DipBuy.SpendQuoteAmount <= 0 {
			return fmt.Errorf("dip-buy requires positive dip_percent and spend_quote_amount")
		}
	}
	if config.TakeProfit.Enabled && config.TakeProfit.TriggerPercent <= 0 {
		return fmt.Errorf("take-profit requires a positive trigger_percent")
	}
	if config.StopLoss.Enabled && config.StopLoss.TriggerPercent <= 0 {
		return fmt.Errorf("stop-loss requires a positive trigger_percent")
	}
	if !config.DipBuy.Enabled && !config.TakeProfit.Enabled && !config.StopLoss.Enabled {
		return fmt.Errorf("at least one strategy rule must be enabled")
	}
	return nil
}

func normalizeStatus(status domain.StrategyStatus) domain.StrategyStatus {
	switch status {
	case domain.StrategyStatusActive, domain.StrategyStatusPaused, domain.StrategyStatusArchived:
		return status
	default:
		return domain.StrategyStatusDraft
	}
}

func firstStrategyID(items []domain.Strategy) string {
	if len(items) == 0 {
		return ""
	}
	return items[0].ID
}

func findPosition(summary domain.PortfolioSummary, marketSymbol string) *domain.Position {
	for _, position := range summary.Positions {
		if position.MarketSymbol == marketSymbol {
			copy := position
			return &copy
		}
	}
	return nil
}

func lifecycleMessage(existing []domain.Strategy, saved domain.Strategy, isPatch bool) (string, string) {
	action := "Strategy saved"
	message := fmt.Sprintf("Automation rules were saved for %s.", saved.MarketSymbol)
	if len(existing) == 0 {
		action = "Strategy created"
		message = fmt.Sprintf("Automation rules were created for %s.", saved.MarketSymbol)
	} else if isPatch {
		action = "Strategy updated"
		message = fmt.Sprintf("Automation rules were updated for %s.", saved.MarketSymbol)
	}

	if len(existing) > 0 && existing[0].Status != saved.Status {
		switch saved.Status {
		case domain.StrategyStatusActive:
			return "Strategy activated", fmt.Sprintf("Automation is now active for %s.", saved.MarketSymbol)
		case domain.StrategyStatusPaused:
			return "Strategy paused", fmt.Sprintf("Automation is paused for %s.", saved.MarketSymbol)
		case domain.StrategyStatusArchived:
			return "Strategy archived", fmt.Sprintf("Automation was archived for %s.", saved.MarketSymbol)
		}
	}

	return action, message
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
