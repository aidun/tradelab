package postgres

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/aidun/tradelab/backend/internal/domain"
)

func TestInsertOrderAndActivityWritesNullStrategyIDForManualOrders(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 3, 30, 8, 0, 0, 0, time.UTC)
	order := domain.Order{
		ID:            "order-1",
		UserID:        "user-1",
		WalletID:      "wallet-1",
		MarketID:      "market-1",
		OrderSource:   domain.OrderSourceManual,
		Side:          domain.OrderSideBuy,
		Type:          domain.OrderTypeMarket,
		Status:        domain.OrderStatusFilled,
		BaseQuantity:  10,
		QuoteAmount:   100,
		ExpectedPrice: 10,
		CreatedAt:     now,
		ExecutedAt:    now,
	}

	mock.ExpectBegin()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(`
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
	`)).WithArgs(
		order.ID,
		order.UserID,
		order.WalletID,
		order.MarketID,
		"",
		order.OrderSource,
		order.Side,
		order.Type,
		order.Status,
		order.BaseQuantity,
		order.QuoteAmount,
		order.BaseQuantity,
		order.ExpectedPrice,
		order.CreatedAt,
		order.ExecutedAt,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(regexp.QuoteMeta(`
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
	`)).WithArgs(
		sqlmock.AnyArg(),
		order.UserID,
		order.WalletID,
		order.ID,
		"trade",
		"Demo buy recorded",
		"Bought 10.0000  at 10.0000 . Position size is now updating in the portfolio view.",
		order.CreatedAt,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	if err := insertOrderAndActivity(context.Background(), tx, order, "Demo buy recorded", "Bought 10.0000  at 10.0000 . Position size is now updating in the portfolio view."); err != nil {
		t.Fatalf("insert order and activity: %v", err)
	}

	mock.ExpectCommit()
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestInsertOrderAndActivityWritesStrategyUUIDForStrategyOrders(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 3, 30, 8, 5, 0, 0, time.UTC)
	order := domain.Order{
		ID:            "order-2",
		UserID:        "user-1",
		WalletID:      "wallet-1",
		MarketID:      "market-1",
		StrategyID:    "11111111-1111-1111-1111-111111111111",
		OrderSource:   domain.OrderSourceStrategy,
		Side:          domain.OrderSideSell,
		Type:          domain.OrderTypeMarket,
		Status:        domain.OrderStatusFilled,
		BaseQuantity:  5,
		QuoteAmount:   50,
		ExpectedPrice: 10,
		CreatedAt:     now,
		ExecutedAt:    now,
	}

	mock.ExpectBegin()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(`
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
	`)).WithArgs(
		order.ID,
		order.UserID,
		order.WalletID,
		order.MarketID,
		order.StrategyID,
		order.OrderSource,
		order.Side,
		order.Type,
		order.Status,
		order.BaseQuantity,
		order.QuoteAmount,
		order.BaseQuantity,
		order.ExpectedPrice,
		order.CreatedAt,
		order.ExecutedAt,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(regexp.QuoteMeta(`
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
	`)).WithArgs(
		sqlmock.AnyArg(),
		order.UserID,
		order.WalletID,
		order.ID,
		"trade",
		"Strategy sell executed",
		"Strategy executed a sell for 5.0000  at 10.0000 .",
		order.CreatedAt,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	if err := insertOrderAndActivity(context.Background(), tx, order, "Strategy sell executed", "Strategy executed a sell for 5.0000  at 10.0000 ."); err != nil {
		t.Fatalf("insert order and activity: %v", err)
	}

	mock.ExpectCommit()
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListByWalletCastsStrategyIDToText(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPortfolioRepository(db)
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "wallet_id", "market_id", "symbol", "strategy_id", "order_source",
		"base_asset", "quote_asset", "requested_quote_amount", "executed_quantity",
		"average_execution_price", "side", "order_type", "status", "submitted_at",
	}).AddRow(
		"order-1", "user-1", "wallet-1", "market-1", "XRP/USDT", "11111111-1111-1111-1111-111111111111",
		"strategy", "XRP", "USDT", 100, 10, 10, "buy", "market", "filled", time.Date(2026, 3, 30, 8, 10, 0, 0, time.UTC),
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
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
		LIMIT $2
	`)).WithArgs("wallet-1", 1).WillReturnRows(rows)

	orders, err := repo.ListByWallet(context.Background(), "wallet-1", 1)
	if err != nil {
		t.Fatalf("list by wallet: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected one order, got %d", len(orders))
	}
	if strings.TrimSpace(orders[0].StrategyID) != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("expected strategy id to round-trip as text, got %q", orders[0].StrategyID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListActivityByWalletQualifiesJoinedColumns(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPortfolioRepository(db)
	rows := sqlmock.NewRows([]string{
		"id", "wallet_id", "market_symbol", "log_type", "title", "message", "created_at",
	}).AddRow(
		"log-1", "wallet-1", "XRP/USDT", "trade", "Demo buy recorded", "Bought 10 XRP.", time.Date(2026, 3, 30, 8, 15, 0, 0, time.UTC),
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT activity_logs.id, activity_logs.wallet_id, COALESCE(order_markets.symbol, strategy_markets.symbol, ''), activity_logs.log_type, activity_logs.title, activity_logs.message, activity_logs.created_at
		FROM activity_logs
		LEFT JOIN orders ON orders.id = activity_logs.order_id
		LEFT JOIN markets order_markets ON order_markets.id = orders.market_id
		LEFT JOIN strategies ON strategies.id = activity_logs.strategy_id
		LEFT JOIN markets strategy_markets ON strategy_markets.id = strategies.market_id
		WHERE activity_logs.wallet_id = $1
		ORDER BY activity_logs.created_at DESC
		LIMIT $2
	`)).WithArgs("wallet-1", 10).WillReturnRows(rows)

	activity, err := repo.ListActivityByWallet(context.Background(), "wallet-1", 10)
	if err != nil {
		t.Fatalf("list activity by wallet: %v", err)
	}
	if len(activity) != 1 {
		t.Fatalf("expected one activity log, got %d", len(activity))
	}
	if activity[0].Title != "Demo buy recorded" {
		t.Fatalf("expected activity title to round-trip, got %q", activity[0].Title)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
