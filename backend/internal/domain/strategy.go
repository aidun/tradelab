package domain

import "time"

type StrategyStatus string
type StrategyDecision string
type StrategyOutcome string

const (
	StrategyStatusDraft    StrategyStatus = "draft"
	StrategyStatusActive   StrategyStatus = "active"
	StrategyStatusPaused   StrategyStatus = "paused"
	StrategyStatusArchived StrategyStatus = "archived"

	StrategyDecisionBuy  StrategyDecision = "buy"
	StrategyDecisionSell StrategyDecision = "sell"
	StrategyDecisionHold StrategyDecision = "hold"

	StrategyOutcomeExecuted StrategyOutcome = "executed"
	StrategyOutcomeSkipped  StrategyOutcome = "skipped"
	StrategyOutcomeErrored  StrategyOutcome = "errored"
)

// StrategyConfig groups the automation rules that can be evaluated for one market.
type StrategyConfig struct {
	DipBuy     DipBuyRule     `json:"dipBuy"`
	TakeProfit TakeProfitRule `json:"takeProfit"`
	StopLoss   StopLossRule   `json:"stopLoss"`
}

type DipBuyRule struct {
	Enabled          bool    `json:"enabled"`
	DipPercent       float64 `json:"dipPercent"`
	SpendQuoteAmount float64 `json:"spendQuoteAmount"`
}

type TakeProfitRule struct {
	Enabled        bool    `json:"enabled"`
	TriggerPercent float64 `json:"triggerPercent"`
}

type StopLossRule struct {
	Enabled        bool    `json:"enabled"`
	TriggerPercent float64 `json:"triggerPercent"`
}

// Strategy represents one automation bundle for a specific wallet and market.
type Strategy struct {
	ID                   string           `json:"id"`
	UserID               string           `json:"userID"`
	WalletID             string           `json:"walletID"`
	MarketID             string           `json:"marketID"`
	MarketSymbol         string           `json:"marketSymbol"`
	Status               StrategyStatus   `json:"status"`
	Config               StrategyConfig   `json:"config"`
	ReferencePrice       float64          `json:"referencePrice"`
	LastRunAt            *time.Time       `json:"lastRunAt,omitempty"`
	LastDecision         StrategyDecision `json:"lastDecision"`
	LastOutcome          StrategyOutcome  `json:"lastOutcome"`
	LastReason           string           `json:"lastReason"`
	EvaluationClaimToken string           `json:"-"`
	EvaluationClaimedAt  *time.Time       `json:"-"`
	CreatedAt            time.Time        `json:"createdAt"`
	UpdatedAt            time.Time        `json:"updatedAt"`
}

// StrategyRun captures one evaluation attempt of a strategy bundle.
type StrategyRun struct {
	ID                   string           `json:"id"`
	StrategyID           string           `json:"strategyID"`
	Decision             StrategyDecision `json:"decision"`
	Outcome              StrategyOutcome  `json:"outcome"`
	Reason               string           `json:"reason"`
	DetailsJSON          string           `json:"detailsJSON"`
	EvaluationDurationMS int64            `json:"evaluationDurationMs"`
	StartedAt            time.Time        `json:"startedAt"`
	FinishedAt           time.Time        `json:"finishedAt"`
}
