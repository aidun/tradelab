package postgres

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aidun/tradelab/backend/internal/domain"
	"github.com/google/uuid"

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

type DemoSessionRepository struct {
	db *sql.DB
}

func NewDemoSessionRepository(db *sql.DB) *DemoSessionRepository {
	return &DemoSessionRepository{db: db}
}

func (r *DemoSessionRepository) CreateDemoSession(ctx context.Context) (domain.DemoSession, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.DemoSession{}, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	expiresAt := now.Add(30 * 24 * time.Hour)
	userID := newUUID()
	walletID := newUUID()
	sessionID := newUUID()
	token, tokenHash, err := newSessionToken()
	if err != nil {
		return domain.DemoSession{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, display_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
	`, userID, "demo+"+userID+"@tradelab.local", "demo-session", "Demo "+userID[:8], now); err != nil {
		return domain.DemoSession{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO wallets (id, user_id, wallet_type, base_currency, starting_balance, created_at, updated_at)
		VALUES ($1, $2, 'paper', 'USDT', 10000, $3, $3)
	`, walletID, userID, now); err != nil {
		return domain.DemoSession{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO wallet_balances (id, wallet_id, asset_id, available_amount, locked_amount, average_entry_price, updated_at)
		SELECT $1, $2, id, 10000, 0, 1, $3
		FROM assets
		WHERE symbol = 'USDT'
	`, newUUID(), walletID, now); err != nil {
		return domain.DemoSession{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO demo_sessions (id, user_id, wallet_id, token_hash, expires_at, created_at, last_used_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
	`, sessionID, userID, walletID, tokenHash, expiresAt, now); err != nil {
		return domain.DemoSession{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.DemoSession{}, err
	}

	return domain.DemoSession{
		ID:         sessionID,
		UserID:     userID,
		WalletID:   walletID,
		Token:      token,
		ExpiresAt:  expiresAt,
		CreatedAt:  now,
		LastUsedAt: now,
	}, nil
}

func (r *DemoSessionRepository) GetByToken(ctx context.Context, token string) (domain.DemoSession, error) {
	var session domain.DemoSession
	tokenHash := hashToken(token)

	err := r.db.QueryRowContext(ctx, `
		UPDATE demo_sessions
		SET last_used_at = NOW()
		WHERE token_hash = $1
		RETURNING id, user_id, wallet_id, expires_at, created_at, last_used_at
	`, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.WalletID,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.LastUsedAt,
	)
	if err != nil {
		return domain.DemoSession{}, err
	}

	return session, nil
}

type AppSessionRepository struct {
	db *sql.DB
}

func NewAppSessionRepository(db *sql.DB) *AppSessionRepository {
	return &AppSessionRepository{db: db}
}

func (r *AppSessionRepository) CreateRegisteredSession(ctx context.Context, userID string, walletID string) (domain.AppSession, error) {
	now := time.Now().UTC()
	idleExpiresAt := now.Add(7 * 24 * time.Hour)
	absoluteExpiresAt := now.Add(30 * 24 * time.Hour)
	sessionID := newUUID()
	token, tokenHash, err := newSessionToken()
	if err != nil {
		return domain.AppSession{}, err
	}

	if _, err := r.db.ExecContext(ctx, `
		UPDATE app_sessions
		SET revoked_at = NOW()
		WHERE user_id = $1 AND principal_kind = $2 AND revoked_at IS NULL
	`, userID, domain.PrincipalKindRegistered); err != nil {
		return domain.AppSession{}, err
	}

	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO app_sessions (
			id,
			user_id,
			wallet_id,
			principal_kind,
			token_hash,
			idle_expires_at,
			absolute_expires_at,
			created_at,
			last_used_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
	`, sessionID, userID, walletID, domain.PrincipalKindRegistered, tokenHash, idleExpiresAt, absoluteExpiresAt, now); err != nil {
		return domain.AppSession{}, err
	}

	return domain.AppSession{
		ID:                sessionID,
		UserID:            userID,
		WalletID:          walletID,
		PrincipalKind:     domain.PrincipalKindRegistered,
		Token:             token,
		IdleExpiresAt:     idleExpiresAt,
		AbsoluteExpiresAt: absoluteExpiresAt,
		CreatedAt:         now,
		LastUsedAt:        now,
	}, nil
}

func (r *AppSessionRepository) GetRegisteredSessionByToken(ctx context.Context, token string) (domain.AppSession, error) {
	var session domain.AppSession
	tokenHash := hashToken(token)

	err := r.db.QueryRowContext(ctx, `
		UPDATE app_sessions
		SET last_used_at = NOW(), idle_expires_at = NOW() + INTERVAL '7 days'
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		RETURNING id, user_id, wallet_id, principal_kind, idle_expires_at, absolute_expires_at, created_at, last_used_at, revoked_at
	`, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.WalletID,
		&session.PrincipalKind,
		&session.IdleExpiresAt,
		&session.AbsoluteExpiresAt,
		&session.CreatedAt,
		&session.LastUsedAt,
		&session.RevokedAt,
	)
	if err != nil {
		return domain.AppSession{}, err
	}

	return session, nil
}

func (r *AppSessionRepository) RevokeRegisteredSessionByToken(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE app_sessions
		SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, hashToken(token))
	return err
}

type RegisteredAccountRepository struct {
	db *sql.DB
}

func NewRegisteredAccountRepository(db *sql.DB) *RegisteredAccountRepository {
	return &RegisteredAccountRepository{db: db}
}

func (r *RegisteredAccountRepository) GetByClerkUserID(ctx context.Context, clerkUserID string) (domain.RegisteredAccount, error) {
	var account domain.RegisteredAccount

	err := r.db.QueryRowContext(ctx, `
		SELECT users.id, wallets.id, users.clerk_user_id, COALESCE(users.email, ''), users.display_name
		FROM users
		JOIN wallets ON wallets.user_id = users.id
		WHERE users.clerk_user_id = $1
		ORDER BY wallets.created_at ASC
		LIMIT 1
	`, clerkUserID).Scan(&account.UserID, &account.WalletID, &account.ClerkUserID, &account.Email, &account.DisplayName)
	if err != nil {
		return domain.RegisteredAccount{}, err
	}

	return account, nil
}

func (r *RegisteredAccountRepository) BootstrapRegisteredAccount(ctx context.Context, identity domain.RegisteredIdentity) (domain.RegisteredAccount, error) {
	existing, err := r.GetByClerkUserID(ctx, identity.ClerkUserID)
	if err == nil {
		return existing, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return domain.RegisteredAccount{}, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.RegisteredAccount{}, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	userID := newUUID()
	walletID := newUUID()
	email := accountEmail(identity)
	displayName := accountDisplayName(identity)

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO users (id, clerk_user_id, email, password_hash, display_name, auth_provider, created_at, updated_at)
		VALUES ($1, $2, $3, NULL, $4, 'clerk', $5, $5)
	`, userID, identity.ClerkUserID, nullableText(email), displayName, now); err != nil {
		return domain.RegisteredAccount{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO wallets (id, user_id, wallet_type, base_currency, starting_balance, created_at, updated_at)
		VALUES ($1, $2, 'paper', 'USDT', 10000, $3, $3)
	`, walletID, userID, now); err != nil {
		return domain.RegisteredAccount{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO wallet_balances (id, wallet_id, asset_id, available_amount, locked_amount, average_entry_price, updated_at)
		SELECT $1, $2, id, 10000, 0, 1, $3
		FROM assets
		WHERE symbol = 'USDT'
	`, newUUID(), walletID, now); err != nil {
		return domain.RegisteredAccount{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.RegisteredAccount{}, err
	}

	return domain.RegisteredAccount{
		UserID:      userID,
		WalletID:    walletID,
		ClerkUserID: identity.ClerkUserID,
		Email:       email,
		DisplayName: displayName,
	}, nil
}

func (r *RegisteredAccountRepository) UpgradeGuestSession(ctx context.Context, guestToken string, identity domain.RegisteredIdentity, preserveGuestData bool) (domain.RegisteredAccount, error) {
	existing, err := r.GetByClerkUserID(ctx, identity.ClerkUserID)
	if err == nil {
		if preserveGuestData {
			return domain.RegisteredAccount{}, fmt.Errorf("registered account already exists")
		}

		if err := r.deleteGuestSessionByToken(ctx, guestToken); err != nil {
			return domain.RegisteredAccount{}, err
		}

		return existing, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return domain.RegisteredAccount{}, err
	}

	if !preserveGuestData {
		account, err := r.BootstrapRegisteredAccount(ctx, identity)
		if err != nil {
			return domain.RegisteredAccount{}, err
		}

		if err := r.deleteGuestSessionByToken(ctx, guestToken); err != nil {
			return domain.RegisteredAccount{}, err
		}

		return account, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.RegisteredAccount{}, err
	}
	defer tx.Rollback()

	tokenHash := hashToken(guestToken)
	var sessionID, userID, walletID string
	err = tx.QueryRowContext(ctx, `
		SELECT id, user_id, wallet_id
		FROM demo_sessions
		WHERE token_hash = $1
	`, tokenHash).Scan(&sessionID, &userID, &walletID)
	if err != nil {
		return domain.RegisteredAccount{}, err
	}

	email := accountEmail(identity)
	displayName := accountDisplayName(identity)

	if _, err := tx.ExecContext(ctx, `
		UPDATE users
		SET clerk_user_id = $1,
			email = $2,
			password_hash = NULL,
			display_name = $3,
			auth_provider = 'clerk',
			updated_at = NOW()
		WHERE id = $4
	`, identity.ClerkUserID, nullableText(email), displayName, userID); err != nil {
		return domain.RegisteredAccount{}, err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM demo_sessions WHERE id = $1`, sessionID); err != nil {
		return domain.RegisteredAccount{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.RegisteredAccount{}, err
	}

	return domain.RegisteredAccount{
		UserID:      userID,
		WalletID:    walletID,
		ClerkUserID: identity.ClerkUserID,
		Email:       email,
		DisplayName: displayName,
	}, nil
}

func (r *RegisteredAccountRepository) deleteGuestSessionByToken(ctx context.Context, guestToken string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM demo_sessions WHERE token_hash = $1`, hashToken(guestToken))
	return err
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
		INSERT INTO wallet_balances (id, wallet_id, asset_id, available_amount, locked_amount, average_entry_price, updated_at)
		SELECT $1, $2, a.id, $3, 0, $4, NOW()
		FROM assets a
		WHERE a.symbol = $5
		ON CONFLICT (wallet_id, asset_id)
		DO UPDATE SET
			available_amount = wallet_balances.available_amount + EXCLUDED.available_amount,
			average_entry_price = (
				((wallet_balances.available_amount * wallet_balances.average_entry_price) + (EXCLUDED.available_amount * EXCLUDED.average_entry_price))
				/ NULLIF(wallet_balances.available_amount + EXCLUDED.available_amount, 0)
			),
			updated_at = NOW()
	`, newUUID(), order.WalletID, order.BaseQuantity, order.ExpectedPrice, order.BaseAsset); err != nil {
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

	if err := insertOrderAndActivity(ctx, tx, order, defaultOrderTitle(order), defaultOrderMessage(order)); err != nil {
		return domain.Order{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Order{}, err
	}

	return order, nil
}

func (r *PortfolioRepository) ApplyMarketSell(ctx context.Context, order domain.Order) (domain.Order, error) {
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
	`, order.BaseQuantity, order.WalletID, order.BaseAsset); err != nil {
		return domain.Order{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE wallet_balances wb
		SET available_amount = wb.available_amount + $1, updated_at = NOW()
		FROM assets a
		WHERE wb.wallet_id = $2
		  AND wb.asset_id = a.id
		  AND a.symbol = $3
	`, order.QuoteAmount, order.WalletID, order.QuoteAsset); err != nil {
		return domain.Order{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE positions
		SET
			entry_quantity = GREATEST(0, entry_quantity - $1),
			status = CASE WHEN entry_quantity - $1 <= 0 THEN 'closed' ELSE status END,
			updated_at = NOW()
		WHERE wallet_id = $2 AND market_id = $3 AND status = 'open'
	`, order.BaseQuantity, order.WalletID, order.MarketID); err != nil {
		return domain.Order{}, err
	}

	if err := insertOrderAndActivity(ctx, tx, order, defaultOrderTitle(order), defaultOrderMessage(order)); err != nil {
		return domain.Order{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Order{}, err
	}

	return order, nil
}

func defaultOrderTitle(order domain.Order) string {
	switch order.OrderSource {
	case domain.OrderSourceStrategy:
		if order.Side == domain.OrderSideSell {
			return "Strategy sell executed"
		}
		return "Strategy buy executed"
	default:
		if order.Side == domain.OrderSideSell {
			return "Demo sell recorded"
		}
		return "Demo buy recorded"
	}
}

func defaultOrderMessage(order domain.Order) string {
	switch {
	case order.OrderSource == domain.OrderSourceStrategy && order.Side == domain.OrderSideBuy:
		return fmt.Sprintf(
			"Strategy executed a buy for %.4f %s at %.4f %s.",
			order.BaseQuantity,
			order.BaseAsset,
			order.ExpectedPrice,
			order.QuoteAsset,
		)
	case order.OrderSource == domain.OrderSourceStrategy && order.Side == domain.OrderSideSell:
		return fmt.Sprintf(
			"Strategy executed a sell for %.4f %s at %.4f %s.",
			order.BaseQuantity,
			order.BaseAsset,
			order.ExpectedPrice,
			order.QuoteAsset,
		)
	case order.Side == domain.OrderSideSell:
		return fmt.Sprintf(
			"Sold %.4f %s at %.4f %s. Review realized PnL in the updated portfolio and order history.",
			order.BaseQuantity,
			order.BaseAsset,
			order.ExpectedPrice,
			order.QuoteAsset,
		)
	default:
		return fmt.Sprintf(
			"Bought %.4f %s at %.4f %s. Position size is now updating in the portfolio view.",
			order.BaseQuantity,
			order.BaseAsset,
			order.ExpectedPrice,
			order.QuoteAsset,
		)
	}
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

	return summary, nil
}

func (r *PortfolioRepository) ListByWallet(ctx context.Context, walletID string, limit int) ([]domain.Order, error) {
	query := `
		SELECT
			o.id,
			o.user_id,
			o.wallet_id,
			o.market_id,
			m.symbol,
			COALESCE(o.strategy_id::text, ''),
			o.order_source,
			base_asset.symbol,
			quote_asset.symbol,
			COALESCE(o.requested_quote_amount, 0),
			COALESCE(o.executed_quantity, 0),
			COALESCE(o.average_execution_price, 0),
			o.side,
			o.order_type,
			o.status,
			o.submitted_at
		FROM orders o
		JOIN markets m ON m.id = o.market_id
		JOIN assets base_asset ON base_asset.id = m.base_asset_id
		JOIN assets quote_asset ON quote_asset.id = m.quote_asset_id
		WHERE o.wallet_id = $1
		ORDER BY o.submitted_at DESC
	`

	var rows *sql.Rows
	var err error
	if limit > 0 {
		query += "\nLIMIT $2"
		rows, err = r.db.QueryContext(ctx, query, walletID, limit)
	} else {
		rows, err = r.db.QueryContext(ctx, query, walletID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Order
	for rows.Next() {
		var item domain.Order
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.WalletID,
			&item.MarketID,
			&item.MarketSymbol,
			&item.StrategyID,
			&item.OrderSource,
			&item.BaseAsset,
			&item.QuoteAsset,
			&item.QuoteAmount,
			&item.BaseQuantity,
			&item.ExpectedPrice,
			&item.Side,
			&item.Type,
			&item.Status,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *PortfolioRepository) ListActivityByWallet(ctx context.Context, walletID string, limit int) ([]domain.ActivityLog, error) {
	query := `
		SELECT activity_logs.id, activity_logs.wallet_id, COALESCE(order_markets.symbol, strategy_markets.symbol, ''), activity_logs.log_type, activity_logs.title, activity_logs.message, activity_logs.created_at
		FROM activity_logs
		LEFT JOIN orders ON orders.id = activity_logs.order_id
		LEFT JOIN markets order_markets ON order_markets.id = orders.market_id
		LEFT JOIN strategies ON strategies.id = activity_logs.strategy_id
		LEFT JOIN markets strategy_markets ON strategy_markets.id = strategies.market_id
		WHERE activity_logs.wallet_id = $1
		ORDER BY activity_logs.created_at DESC
	`

	var rows *sql.Rows
	var err error
	if limit > 0 {
		query += "\nLIMIT $2"
		rows, err = r.db.QueryContext(ctx, query, walletID, limit)
	} else {
		rows, err = r.db.QueryContext(ctx, query, walletID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ActivityLog
	for rows.Next() {
		var item domain.ActivityLog
		if err := rows.Scan(&item.ID, &item.WalletID, &item.MarketSymbol, &item.LogType, &item.Title, &item.Message, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *PortfolioRepository) Create(ctx context.Context, order domain.Order) (domain.Order, error) {
	return domain.Order{}, fmt.Errorf("unsupported operation: use ApplyMarketBuy or ApplyMarketSell")
}

func insertOrderAndActivity(ctx context.Context, tx *sql.Tx, order domain.Order, title string, message string) error {
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO orders (
			id,
			user_id,
			wallet_id,
			market_id,
			strategy_id,
			order_source,
			side,
			order_type,
			status,
			requested_quantity,
			requested_quote_amount,
			executed_quantity,
			average_execution_price,
			submitted_at,
			executed_at
		) VALUES ($1, $2, $3, $4, NULLIF($5, '')::uuid, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, order.ID, order.UserID, order.WalletID, order.MarketID, order.StrategyID, order.OrderSource, order.Side, order.Type, order.Status, order.BaseQuantity, order.QuoteAmount, order.BaseQuantity, order.ExpectedPrice, order.CreatedAt, order.ExecutedAt); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO activity_logs (
			id,
			user_id,
			wallet_id,
			order_id,
			log_type,
			title,
			message,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, newUUID(), order.UserID, order.WalletID, order.ID, "trade", title, strings.TrimSpace(message), order.CreatedAt); err != nil {
		return err
	}

	return nil
}

type StrategyRepository struct {
	db *sql.DB
}

func NewStrategyRepository(db *sql.DB) *StrategyRepository {
	return &StrategyRepository{db: db}
}

func (r *StrategyRepository) ListByWallet(ctx context.Context, walletID string, marketSymbol string) ([]domain.Strategy, error) {
	query := `
		SELECT s.id, s.user_id, s.wallet_id, s.market_id, m.symbol, s.status, s.config_json, s.reference_price,
			s.last_run_at, COALESCE(s.last_decision, ''), COALESCE(s.last_outcome, ''), COALESCE(s.last_reason, ''),
			s.created_at, s.updated_at
		FROM strategies s
		JOIN markets m ON m.id = s.market_id
		WHERE s.wallet_id = $1
	`
	args := []any{walletID}
	if marketSymbol != "" {
		query += ` AND m.symbol = $2`
		args = append(args, marketSymbol)
	}
	query += ` ORDER BY m.symbol ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Strategy
	for rows.Next() {
		item, err := scanStrategy(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *StrategyRepository) UpsertForWalletMarket(ctx context.Context, strategy domain.Strategy) (domain.Strategy, error) {
	configJSON, err := json.Marshal(strategy.Config)
	if err != nil {
		return domain.Strategy{}, err
	}

	var item domain.Strategy
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO strategies (
			id, user_id, wallet_id, market_id, status, config_json, reference_price,
			last_decision, last_outcome, last_reason, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, '', '', '', NOW(), NOW())
		ON CONFLICT (wallet_id, market_id)
		DO UPDATE SET
			status = EXCLUDED.status,
			config_json = EXCLUDED.config_json,
			updated_at = NOW()
		RETURNING id, user_id, wallet_id, market_id, (SELECT symbol FROM markets WHERE id = strategies.market_id),
			status, config_json, reference_price, last_run_at, COALESCE(last_decision, ''), COALESCE(last_outcome, ''),
			COALESCE(last_reason, ''), created_at, updated_at
	`, zeroIfEmpty(strategy.ID, newUUID()), strategy.UserID, strategy.WalletID, strategy.MarketID, strategy.Status, configJSON, strategy.ReferencePrice).Scan(
		&item.ID,
		&item.UserID,
		&item.WalletID,
		&item.MarketID,
		&item.MarketSymbol,
		&item.Status,
		&configJSON,
		&item.ReferencePrice,
		&item.LastRunAt,
		&item.LastDecision,
		&item.LastOutcome,
		&item.LastReason,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return domain.Strategy{}, err
	}
	if err := json.Unmarshal(configJSON, &item.Config); err != nil {
		return domain.Strategy{}, err
	}
	return item, nil
}

func (r *StrategyRepository) GetByIDForWallet(ctx context.Context, walletID string, strategyID string) (domain.Strategy, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT s.id, s.user_id, s.wallet_id, s.market_id, m.symbol, s.status, s.config_json, s.reference_price,
			s.last_run_at, COALESCE(s.last_decision, ''), COALESCE(s.last_outcome, ''), COALESCE(s.last_reason, ''),
			s.created_at, s.updated_at
		FROM strategies s
		JOIN markets m ON m.id = s.market_id
		WHERE s.wallet_id = $1 AND s.id = $2
	`, walletID, strategyID)
	return scanStrategy(row)
}

func (r *StrategyRepository) ClaimActiveStrategies(ctx context.Context, claimToken string, limit int, staleBefore time.Time) ([]domain.Strategy, error) {
	if limit <= 0 {
		limit = 10
	}

	// Claiming happens in SQL so multiple API replicas can compete safely without double-evaluating the same bundle.
	rows, err := r.db.QueryContext(ctx, `
		WITH candidates AS (
			SELECT s.id
			FROM strategies s
			WHERE s.status = $1
			  AND (s.evaluation_claim_token IS NULL OR s.evaluation_claimed_at IS NULL OR s.evaluation_claimed_at < $2)
			ORDER BY COALESCE(s.last_run_at, s.created_at) ASC
			LIMIT $3
			FOR UPDATE SKIP LOCKED
		)
		UPDATE strategies s
		SET evaluation_claim_token = $4, evaluation_claimed_at = NOW()
		FROM candidates
		WHERE s.id = candidates.id
		RETURNING s.id, s.user_id, s.wallet_id, s.market_id, (SELECT symbol FROM markets WHERE id = s.market_id), s.status, s.config_json, s.reference_price,
			s.last_run_at, COALESCE(s.last_decision, ''), COALESCE(s.last_outcome, ''), COALESCE(s.last_reason, ''),
			s.created_at, s.updated_at
	`, domain.StrategyStatusActive, staleBefore, limit, claimToken)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Strategy
	for rows.Next() {
		item, err := scanStrategy(rows)
		if err != nil {
			return nil, err
		}
		item.EvaluationClaimToken = claimToken
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *StrategyRepository) RecordEvaluation(ctx context.Context, strategy domain.Strategy, run domain.StrategyRun, nextReferencePrice float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := r.insertStrategyRun(ctx, tx, run); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE strategies
		SET
			reference_price = $1,
			last_run_at = $2,
			last_decision = $3,
			last_outcome = $4,
			last_reason = $5,
			evaluation_claim_token = NULL,
			evaluation_claimed_at = NULL,
			updated_at = NOW()
		WHERE id = $6 AND evaluation_claim_token = $7
	`, nextReferencePrice, run.FinishedAt, run.Decision, run.Outcome, run.Reason, strategy.ID, strategy.EvaluationClaimToken); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *StrategyRepository) RecordEvaluationError(ctx context.Context, strategy domain.Strategy, run domain.StrategyRun) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := r.insertStrategyRun(ctx, tx, run); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE strategies
		SET
			last_run_at = $1,
			last_decision = $2,
			last_outcome = $3,
			last_reason = $4,
			evaluation_claim_token = NULL,
			evaluation_claimed_at = NULL,
			updated_at = NOW()
		WHERE id = $5 AND evaluation_claim_token = $6
	`, run.FinishedAt, run.Decision, run.Outcome, run.Reason, strategy.ID, strategy.EvaluationClaimToken); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO activity_logs (id, user_id, wallet_id, strategy_id, log_type, title, message, created_at)
		VALUES ($1, $2, $3, $4, 'strategy', 'Strategy evaluation failed', $5, $6)
	`, newUUID(), strategy.UserID, strategy.WalletID, strategy.ID, run.Reason, run.FinishedAt); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *StrategyRepository) RecordLifecycleActivity(ctx context.Context, strategy domain.Strategy, title string, message string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO activity_logs (id, user_id, wallet_id, strategy_id, log_type, title, message, created_at)
		VALUES ($1, $2, $3, $4, 'strategy', $5, $6, NOW())
	`, newUUID(), strategy.UserID, strategy.WalletID, strategy.ID, title, strings.TrimSpace(message))
	return err
}

func (r *StrategyRepository) insertStrategyRun(ctx context.Context, tx *sql.Tx, run domain.StrategyRun) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO strategy_runs (id, strategy_id, decision, outcome, reason, details_json, evaluation_duration_ms, started_at, finished_at)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, $8, $9)
	`, run.ID, run.StrategyID, run.Decision, run.Outcome, run.Reason, run.DetailsJSON, run.EvaluationDurationMS, run.StartedAt, run.FinishedAt)
	return err
}

func scanStrategy(scanner interface {
	Scan(dest ...any) error
}) (domain.Strategy, error) {
	var item domain.Strategy
	var configJSON []byte
	err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&item.WalletID,
		&item.MarketID,
		&item.MarketSymbol,
		&item.Status,
		&configJSON,
		&item.ReferencePrice,
		&item.LastRunAt,
		&item.LastDecision,
		&item.LastOutcome,
		&item.LastReason,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return domain.Strategy{}, err
	}
	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &item.Config); err != nil {
			return domain.Strategy{}, err
		}
	}
	return item, nil
}

func zeroIfEmpty(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func newUUID() string {
	return uuid.NewString()
}

func newSessionToken() (string, string, error) {
	entropy := make([]byte, 32)
	if _, err := rand.Read(entropy); err != nil {
		return "", "", err
	}

	token := base64.RawURLEncoding.EncodeToString(entropy)
	return token, hashToken(token), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func accountEmail(identity domain.RegisteredIdentity) string {
	if identity.Email != "" {
		return identity.Email
	}

	return ""
}

func accountDisplayName(identity domain.RegisteredIdentity) string {
	if identity.DisplayName != "" {
		return identity.DisplayName
	}

	return "Trader " + truncate(identity.ClerkUserID)
}

func nullableText(value string) any {
	if value == "" {
		return nil
	}

	return value
}

func truncate(value string) string {
	if len(value) <= 8 {
		return value
	}

	return value[:8]
}
