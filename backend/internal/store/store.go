package store

import (
	"context"

	"github.com/aidun/tradelab/backend/internal/domain"
)

type MarketRepository interface {
	List(ctx context.Context) ([]domain.Market, error)
	GetBySymbol(ctx context.Context, symbol string) (domain.Market, error)
}

type BalanceRepository interface {
	GetByWalletAndAsset(ctx context.Context, walletID string, assetSymbol string) (domain.Balance, error)
}

type DemoSessionRepository interface {
	CreateDemoSession(ctx context.Context) (domain.DemoSession, error)
	GetByToken(ctx context.Context, token string) (domain.DemoSession, error)
}

type AppSessionRepository interface {
	CreateRegisteredSession(ctx context.Context, userID string, walletID string) (domain.AppSession, error)
	GetRegisteredSessionByToken(ctx context.Context, token string) (domain.AppSession, error)
	RevokeRegisteredSessionByToken(ctx context.Context, token string) error
}

type RegisteredAccountRepository interface {
	GetByClerkUserID(ctx context.Context, clerkUserID string) (domain.RegisteredAccount, error)
	BootstrapRegisteredAccount(ctx context.Context, identity domain.RegisteredIdentity) (domain.RegisteredAccount, error)
	UpgradeGuestSession(ctx context.Context, guestToken string, identity domain.RegisteredIdentity, preserveGuestData bool) (domain.RegisteredAccount, error)
}

type OrderRepository interface {
	Create(ctx context.Context, order domain.Order) (domain.Order, error)
	ListByWallet(ctx context.Context, walletID string, limit int) ([]domain.Order, error)
	ListActivityByWallet(ctx context.Context, walletID string, limit int) ([]domain.ActivityLog, error)
}

type PortfolioRepository interface {
	ApplyMarketBuy(ctx context.Context, order domain.Order) (domain.Order, error)
	ApplyMarketSell(ctx context.Context, order domain.Order) (domain.Order, error)
	GetSummary(ctx context.Context, walletID string) (domain.PortfolioSummary, error)
	ListByWallet(ctx context.Context, walletID string, limit int) ([]domain.Order, error)
	ListActivityByWallet(ctx context.Context, walletID string, limit int) ([]domain.ActivityLog, error)
}
