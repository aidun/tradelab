package store

import (
	"context"

	"github.com/aidun/tradelab/backend/internal/domain"
)

type MarketRepository interface {
	GetBySymbol(ctx context.Context, symbol string) (domain.Market, error)
}

type BalanceRepository interface {
	GetByWalletAndAsset(ctx context.Context, walletID string, assetSymbol string) (domain.Balance, error)
}

type OrderRepository interface {
	Create(ctx context.Context, order domain.Order) (domain.Order, error)
}
