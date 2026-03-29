package domain

import "time"

type OrderSide string
type OrderType string
type OrderStatus string

const (
	OrderSideBuy OrderSide = "buy"

	OrderTypeMarket OrderType = "market"

	OrderStatusPending OrderStatus = "pending"
	OrderStatusFilled  OrderStatus = "filled"
)

type Order struct {
	ID            string
	UserID        string
	WalletID      string
	MarketID      string
	MarketSymbol  string
	BaseAsset     string
	QuoteAsset    string
	QuoteAmount   float64
	BaseQuantity  float64
	ExpectedPrice float64
	Side          OrderSide
	Type          OrderType
	Status        OrderStatus
	CreatedAt     time.Time
	ExecutedAt    time.Time
}

type Balance struct {
	WalletID     string
	AssetSymbol string
	Available   float64
}

type Position struct {
	ID             string
	UserID         string
	WalletID       string
	MarketID       string
	MarketSymbol   string
	BaseAsset      string
	QuoteAsset     string
	Status         string
	EntryQuantity  float64
	EntryPriceAvg  float64
	CurrentPrice   float64
	PositionValue  float64
	UnrealizedPnL  float64
	OpenedAt       time.Time
}

type PortfolioSummary struct {
	WalletID     string
	BaseCurrency string
	TotalValue   float64
	CashBalance  float64
	Positions    []Position
	Balances     []Balance
}
