package domain

import "time"

type OrderSide string
type OrderType string
type OrderStatus string

const (
	OrderSideBuy OrderSide = "buy"

	OrderTypeMarket OrderType = "market"

	OrderStatusPending OrderStatus = "pending"
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
	ExpectedPrice float64
	Side          OrderSide
	Type          OrderType
	Status        OrderStatus
	CreatedAt     time.Time
}

type Balance struct {
	AssetSymbol string
	Available   float64
}
