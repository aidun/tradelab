package postgres

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
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
	`, order.ID, order.UserID, order.WalletID, order.ID, "trade", "Demo buy recorded", "A demo market buy was created for "+order.MarketSymbol+".", order.CreatedAt); err != nil {
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
		summary.TotalValue += position.PositionValue
		summary.Positions = append(summary.Positions, position)
	}

	return summary, nil
}

func (r *PortfolioRepository) ListByWallet(ctx context.Context, walletID string, limit int) ([]domain.Order, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			o.id,
			o.user_id,
			o.wallet_id,
			o.market_id,
			m.symbol,
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
		LIMIT $2
	`, walletID, limit)
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
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, wallet_id, log_type, title, message, created_at
		FROM activity_logs
		WHERE wallet_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, walletID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ActivityLog
	for rows.Next() {
		var item domain.ActivityLog
		if err := rows.Scan(&item.ID, &item.WalletID, &item.LogType, &item.Title, &item.Message, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *PortfolioRepository) Create(ctx context.Context, order domain.Order) (domain.Order, error) {
	return domain.Order{}, fmt.Errorf("unsupported operation: use ApplyMarketBuy")
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
