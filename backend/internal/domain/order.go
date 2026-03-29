package domain

import "time"

type OrderSide string
type OrderType string
type OrderStatus string
type AccountingMode string
type OrderSource string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"

	OrderTypeMarket OrderType = "market"

	OrderStatusPending OrderStatus = "pending"
	OrderStatusFilled  OrderStatus = "filled"

	OrderSourceManual   OrderSource = "manual"
	OrderSourceStrategy OrderSource = "strategy"
	OrderSourceSystem   OrderSource = "system"

	AccountingModeAverageCost AccountingMode = "average_cost"
	AccountingModeFIFO        AccountingMode = "fifo"
	AccountingModeHybrid      AccountingMode = "hybrid"
)

type Order struct {
	ID            string      `json:"id"`
	UserID        string      `json:"userID"`
	WalletID      string      `json:"walletID"`
	MarketID      string      `json:"marketID"`
	MarketSymbol  string      `json:"marketSymbol"`
	StrategyID    string      `json:"strategyID,omitempty"`
	OrderSource   OrderSource `json:"orderSource"`
	BaseAsset     string      `json:"baseAsset"`
	QuoteAsset    string      `json:"quoteAsset"`
	QuoteAmount   float64     `json:"quoteAmount"`
	BaseQuantity  float64     `json:"baseQuantity"`
	ExpectedPrice float64     `json:"expectedPrice"`
	Side          OrderSide   `json:"side"`
	Type          OrderType   `json:"type"`
	Status        OrderStatus `json:"status"`
	RealizedPnL   float64     `json:"realizedPnL"`
	PositionAfter float64     `json:"positionAfter"`
	CreatedAt     time.Time   `json:"createdAt"`
	ExecutedAt    time.Time   `json:"executedAt"`
}

type Balance struct {
	WalletID    string  `json:"walletID"`
	AssetSymbol string  `json:"assetSymbol"`
	Available   float64 `json:"available"`
}

type Position struct {
	ID             string    `json:"id"`
	UserID         string    `json:"userID"`
	WalletID       string    `json:"walletID"`
	MarketID       string    `json:"marketID"`
	MarketSymbol   string    `json:"marketSymbol"`
	BaseAsset      string    `json:"baseAsset"`
	QuoteAsset     string    `json:"quoteAsset"`
	Status         string    `json:"status"`
	OpenQuantity   float64   `json:"openQuantity"`
	EntryQuantity  float64   `json:"entryQuantity"`
	EntryPriceAvg  float64   `json:"entryPriceAvg"`
	CurrentPrice   float64   `json:"currentPrice"`
	CostBasisValue float64   `json:"costBasisValue"`
	PositionValue  float64   `json:"positionValue"`
	RealizedPnL    float64   `json:"realizedPnL"`
	UnrealizedPnL  float64   `json:"unrealizedPnL"`
	OpenedAt       time.Time `json:"openedAt"`
}

type PortfolioAllocation struct {
	MarketSymbol string  `json:"marketSymbol"`
	Value        float64 `json:"value"`
	Weight       float64 `json:"weight"`
}

type PortfolioSummary struct {
	WalletID       string                `json:"walletID"`
	BaseCurrency   string                `json:"baseCurrency"`
	AccountingMode AccountingMode        `json:"accountingMode"`
	TotalValue     float64               `json:"totalValue"`
	CashBalance    float64               `json:"cashBalance"`
	PositionValue  float64               `json:"positionValue"`
	RealizedPnL    float64               `json:"realizedPnL"`
	UnrealizedPnL  float64               `json:"unrealizedPnL"`
	Positions      []Position            `json:"positions"`
	Balances       []Balance             `json:"balances"`
	Allocations    []PortfolioAllocation `json:"allocations"`
}

type ActivityLog struct {
	ID           string    `json:"id"`
	WalletID     string    `json:"walletID"`
	MarketSymbol string    `json:"marketSymbol"`
	LogType      string    `json:"logType"`
	Title        string    `json:"title"`
	Message      string    `json:"message"`
	CreatedAt    time.Time `json:"createdAt"`
}

func NormalizeAccountingMode(raw string) AccountingMode {
	switch AccountingMode(raw) {
	case AccountingModeFIFO:
		return AccountingModeFIFO
	case AccountingModeHybrid:
		return AccountingModeHybrid
	default:
		return AccountingModeAverageCost
	}
}
