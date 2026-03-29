package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/aidun/tradelab/backend/internal/domain"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(databaseURL string) (*sql.DB, error) {
	return sql.Open("pgx", databaseURL)
}

type MarketRepository struct {
	db *sql.DB
}

func NewMarketRepository(db *sql.DB) *MarketRepository {
	return &MarketRepository{db: db}
}

func (r *MarketRepository) List(ctx context.Context) ([]domain.Market, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT m.id, m.symbol, base_asset.symbol, quote_asset.symbol, m.min_order_size, m.exchange_code
		FROM markets m
		JOIN assets base_asset ON base_asset.id = m.base_asset_id
		JOIN assets quote_asset ON quote_asset.id = m.quote_asset_id
		WHERE m.is_active = TRUE
		ORDER BY m.symbol ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var markets []domain.Market

	for rows.Next() {
		var item domain.Market
		if err := rows.Scan(&item.ID, &item.Symbol, &item.BaseAsset, &item.QuoteAsset, &item.MinNotional, &item.Exchange); err != nil {
			return nil, err
		}

		markets = append(markets, item)
	}

	return markets, rows.Err()
}

func (r *MarketRepository) GetBySymbol(ctx context.Context, symbol string) (domain.Market, error) {
	var item domain.Market

	err := r.db.QueryRowContext(ctx, `
		SELECT m.id, m.symbol, base_asset.symbol, quote_asset.symbol, m.min_order_size, m.exchange_code
		FROM markets m
		JOIN assets base_asset ON base_asset.id = m.base_asset_id
		JOIN assets quote_asset ON quote_asset.id = m.quote_asset_id
		WHERE m.symbol = $1 AND m.is_active = TRUE
	`, symbol).Scan(&item.ID, &item.Symbol, &item.BaseAsset, &item.QuoteAsset, &item.MinNotional, &item.Exchange)
	if err != nil {
		return domain.Market{}, err
	}

	return item, nil
}

type BalanceRepository struct {
	db *sql.DB
}

func NewBalanceRepository(db *sql.DB) *BalanceRepository {
	return &BalanceRepository{db: db}
}

func (r *BalanceRepository) GetByWalletAndAsset(ctx context.Context, walletID string, assetSymbol string) (domain.Balance, error) {
	var balance domain.Balance

	err := r.db.QueryRowContext(ctx, `
		SELECT assets.symbol, wallet_balances.available_amount
		FROM wallet_balances
		JOIN assets ON assets.id = wallet_balances.asset_id
		WHERE wallet_balances.wallet_id = $1 AND assets.symbol = $2
	`, walletID, assetSymbol).Scan(&balance.AssetSymbol, &balance.Available)
	if err != nil {
		return domain.Balance{}, err
	}

	return balance, nil
}

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order domain.Order) (domain.Order, error) {
	order.CreatedAt = time.Now().UTC()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO orders (
			id,
			user_id,
			wallet_id,
			market_id,
			order_source,
			side,
			order_type,
			status,
			requested_quote_amount,
			average_execution_price,
			submitted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, order.ID, order.UserID, order.WalletID, order.MarketID, "manual", order.Side, order.Type, order.Status, order.QuoteAmount, order.ExpectedPrice, order.CreatedAt)
	if err != nil {
		return domain.Order{}, err
	}

	return order, nil
}
