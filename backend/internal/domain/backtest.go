package domain

import "time"

// BacktestSummary captures the first product-facing performance metrics for one run.
type BacktestSummary struct {
	ReturnPercent      float64 `json:"returnPercent"`
	TradeCount         int     `json:"tradeCount"`
	SellCount          int     `json:"sellCount"`
	WinningTradeCount  int     `json:"winningTradeCount"`
	HitRatePercent     float64 `json:"hitRatePercent"`
	MaxDrawdownPercent float64 `json:"maxDrawdownPercent"`
}

// BacktestEquityPoint records the simulated account state at one candle close.
type BacktestEquityPoint struct {
	Time          time.Time `json:"time"`
	Price         float64   `json:"price"`
	CashBalance   float64   `json:"cashBalance"`
	OpenQuantity  float64   `json:"openQuantity"`
	PositionValue float64   `json:"positionValue"`
	TotalEquity   float64   `json:"totalEquity"`
	DrawdownPct   float64   `json:"drawdownPercent"`
}

// BacktestRun is the read-only historical simulation result returned to the UI.
type BacktestRun struct {
	MarketSymbol     string                `json:"marketSymbol"`
	BaseAsset        string                `json:"baseAsset"`
	QuoteAsset       string                `json:"quoteAsset"`
	Interval         string                `json:"interval"`
	StartTime        time.Time             `json:"startTime"`
	EndTime          time.Time             `json:"endTime"`
	InitialCash      float64               `json:"initialCash"`
	FinalCash        float64               `json:"finalCash"`
	FinalPositionQty float64               `json:"finalPositionQty"`
	FinalPositionVal float64               `json:"finalPositionValue"`
	FinalEquity      float64               `json:"finalEquity"`
	Orders           []Order               `json:"orders"`
	EquityCurve      []BacktestEquityPoint `json:"equityCurve"`
	Summary          BacktestSummary       `json:"summary"`
}
