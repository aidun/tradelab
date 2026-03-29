package postgres

import (
	"context"
	"database/sql"
	"fmt"

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
		SELECT wallet_balances.wallet_id, assets.symbol, wallet_balances.available_amount
		FROM wallet_balances
		JOIN assets ON assets.id = wallet_balances.asset_id
		WHERE wallet_balances.wallet_id = $1 AND assets.symbol = $2
	`, walletID, assetSymbol).Scan(&balance.WalletID, &balance.AssetSymbol, &balance.Available)
	if err != nil {
		return domain.Balance{}, err
	}

	return balance, nil
}

type PortfolioRepository struct {
	db *sql.DB
}

func NewPortfolioRepository(db *sql.DB) *PortfolioRepository {
	return &PortfolioRepository{db: db}
}

func (r *PortfolioRepository) ApplyMarketBuy(ctx context.Context, order domain.Order) (domain.Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Order{}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE wallet_balances wb
		SET available_amount = wb.available_amount - $1, updated_at = NOW()
		FROM assets a
		WHERE wb.wallet_id = $2
		  AND wb.asset_id = a.id
		  AND a.symbol = $3
	`, order.QuoteAmount, order.WalletID, order.QuoteAsset); err != nil {
		return domain.Order{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE wallet_balances wb
		SET available_amount = wb.available_amount + $1, average_entry_price = $2, updated_at = NOW()
		FROM assets a
		WHERE wb.wallet_id = $3
		  AND wb.asset_id = a.id
		  AND a.symbol = $4
	`, order.BaseQuantity, order.ExpectedPrice, order.WalletID, order.BaseAsset); err != nil {
		return domain.Order{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO positions (
			id,
			user_id,
			wallet_id,
			market_id,
			status,
			entry_quantity,
			entry_price_avg,
			opened_at,
			updated_at
		) VALUES ($1, $2, $3, $4, 'open', $5, $6, NOW(), NOW())
		ON CONFLICT (wallet_id, market_id, status)
		DO UPDATE SET
			entry_quantity = positions.entry_quantity + EXCLUDED.entry_quantity,
			entry_price_avg = (
				((positions.entry_quantity * positions.entry_price_avg) + (EXCLUDED.entry_quantity * EXCLUDED.entry_price_avg))
				/ NULLIF((positions.entry_quantity + EXCLUDED.entry_quantity), 0)
			),
			updated_at = NOW()
	`, order.ID, order.UserID, order.WalletID, order.MarketID, order.BaseQuantity, order.ExpectedPrice); err != nil {
		return domain.Order{}, err
	}

	if _, err := tx.ExecContext(ctx, `
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
			executed_quantity,
			average_execution_price,
			submitted_at,
			executed_at
		) VALUES ($1, $2, $3, $4, 'manual', $5, $6, $7, $8, $9, $10, $11, $12)
	`, order.ID, order.UserID, order.WalletID, order.MarketID, order.Side, order.Type, order.Status, order.QuoteAmount, order.BaseQuantity, order.ExpectedPrice, order.CreatedAt, order.ExecutedAt); err != nil {
		return domain.Order{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Order{}, err
	}

	return order, nil
}

func (r *PortfolioRepository) GetSummary(ctx context.Context, walletID string) (domain.PortfolioSummary, error) {
	var summary domain.PortfolioSummary
	summary.WalletID = walletID
	summary.BaseCurrency = "USDT"

	rows, err := r.db.QueryContext(ctx, `
		SELECT wb.wallet_id, a.symbol, wb.available_amount
		FROM wallet_balances wb
		JOIN assets a ON a.id = wb.asset_id
		WHERE wb.wallet_id = $1
		ORDER BY a.symbol ASC
	`, walletID)
	if err != nil {
		return domain.PortfolioSummary{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var balance domain.Balance
		if err := rows.Scan(&balance.WalletID, &balance.AssetSymbol, &balance.Available); err != nil {
			return domain.PortfolioSummary{}, err
		}

		if balance.AssetSymbol == summary.BaseCurrency {
			summary.CashBalance = balance.Available
			summary.TotalValue += balance.Available
		}

		summary.Balances = append(summary.Balances, balance)
	}

	positionRows, err := r.db.QueryContext(ctx, `
		SELECT
			p.id,
			p.user_id,
			p.wallet_id,
			p.market_id,
			m.symbol,
			base_asset.symbol,
			quote_asset.symbol,
			p.status,
			p.entry_quantity,
			p.entry_price_avg,
			p.opened_at
		FROM positions p
		JOIN markets m ON m.id = p.market_id
		JOIN assets base_asset ON base_asset.id = m.base_asset_id
		JOIN assets quote_asset ON quote_asset.id = m.quote_asset_id
		WHERE p.wallet_id = $1 AND p.status = 'open'
		ORDER BY p.opened_at DESC
	`, walletID)
	if err != nil {
		return domain.PortfolioSummary{}, err
	}
	defer positionRows.Close()

	for positionRows.Next() {
		var position domain.Position
		if err := positionRows.Scan(
			&position.ID,
			&position.UserID,
			&position.WalletID,
			&position.MarketID,
			&position.MarketSymbol,
			&position.BaseAsset,
			&position.QuoteAsset,
			&position.Status,
			&position.EntryQuantity,
			&position.EntryPriceAvg,
			&position.OpenedAt,
		); err != nil {
			return domain.PortfolioSummary{}, err
		}

		position.CurrentPrice = position.EntryPriceAvg
		position.PositionValue = position.EntryQuantity * position.EntryPriceAvg
		position.UnrealizedPnL = 0

		summary.TotalValue += position.PositionValue
		summary.Positions = append(summary.Positions, position)
	}

	return summary, nil
}

func (r *PortfolioRepository) Create(ctx context.Context, order domain.Order) (domain.Order, error) {
	return domain.Order{}, fmt.Errorf("unsupported operation: use ApplyMarketBuy")
}
