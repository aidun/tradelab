# Data model: TradeLab

## 1. Design principles

The model is asset-agnostic. XRP is a reference asset, but all core entities are designed around generic assets, markets, wallets, orders, positions, and strategies.

Key principles:

- model assets separately from markets
- keep wallet balances separate from trades and positions
- log every strategy evaluation for traceability
- represent manual and automated orders with the same model

## 2. Core entities

### users

Stores user accounts.

- id
- email
- password_hash
- display_name
- created_at
- updated_at

### wallets

One demo wallet per user in the MVP, with support for more later.

- id
- user_id
- wallet_type
- base_currency
- starting_balance
- created_at
- updated_at

### wallet_balances

Current balances per asset inside a wallet.

- id
- wallet_id
- asset_id
- available_amount
- locked_amount
- average_entry_price
- updated_at

### assets

Master records for supported coins or tokens.

- id
- symbol
- name
- asset_type
- is_active
- created_at

Examples:

- XRP
- BTC
- ETH

### markets

Trading pairs such as XRP/USDT or BTC/USDT.

- id
- base_asset_id
- quote_asset_id
- symbol
- exchange_code
- tick_size
- min_order_size
- is_active
- created_at

### price_candles

Persisted market time series for charting and backtesting.

- id
- market_id
- timeframe
- timestamp
- open
- high
- low
- close
- volume

For the MVP, candle persistence is enough. Real-time caching can be added later.

### orders

Stores every order request and execution, whether manual or strategy-driven.

- id
- user_id
- wallet_id
- market_id
- strategy_id nullable
- order_source
- side
- order_type
- status
- requested_quantity
- requested_quote_amount nullable
- executed_quantity
- average_execution_price
- fee_amount
- fee_asset_id
- slippage_bps
- submitted_at
- executed_at nullable
- cancelled_at nullable

Suggested values:

- order_source: manual, strategy, system
- side: buy, sell
- order_type: market, limit
- status: pending, filled, partial, cancelled, rejected

### trades

Granular executions per order if split fills are needed later.

- id
- order_id
- market_id
- side
- quantity
- price
- fee_amount
- fee_asset_id
- executed_at

The MVP can start with one trade per order.

### positions

Aggregated view of open and closed market exposure.

- id
- user_id
- wallet_id
- market_id
- strategy_id nullable
- status
- opened_at
- closed_at nullable
- entry_quantity
- entry_price_avg
- exit_quantity
- exit_price_avg
- realized_pnl
- unrealized_pnl nullable

Suggested values:

- status: open, closed

### strategies

User-defined bot rules.

- id
- user_id
- wallet_id
- market_id
- name
- strategy_type
- config_json
- risk_config_json
- status
- created_at
- updated_at
- last_run_at nullable

Suggested strategy types:

- dip_buy
- take_profit_stop_loss
- dca
- sma_crossover
- rsi_trigger

Suggested status values:

- draft
- active
- paused
- archived

### strategy_runs

Logs each strategy evaluation by the engine.

- id
- strategy_id
- started_at
- finished_at
- outcome
- decision
- signal_strength nullable
- details_json

Examples:

- outcome: executed, skipped, errored
- decision: buy, sell, hold

### backtests

Stores backtest definitions and results.

- id
- user_id
- strategy_id nullable
- market_id
- name
- timeframe
- date_from
- date_to
- config_snapshot_json
- result_summary_json
- created_at

### activity_logs

Stores user-facing and system-facing audit events.

- id
- user_id
- wallet_id nullable
- strategy_id nullable
- order_id nullable
- log_type
- title
- message
- metadata_json nullable
- created_at

Suggested values:

- log_type: info, warning, trade, strategy, system

## 3. Relationships

- a user has many wallets
- a wallet has many wallet balances
- an asset can appear in many markets as base or quote
- a market has many price candles
- a wallet has many orders
- an order can belong to one strategy
- an order can have one or more trades
- a user and wallet can have many positions
- a user has many strategies
- a strategy has many strategy runs
- a strategy or market can have many backtests
- a user has many activity logs

## 4. Example: XRP demo trade

1. The user has 10,000 USDT in a demo wallet.
2. XRP/USDT exists in `markets`.
3. The user sends a market buy order.
4. A record is created in `orders` with source `manual`.
5. The execution is stored in `trades`.
6. `wallet_balances` is updated for USDT and XRP.
7. A `position` is opened or increased.
8. An `activity_log` captures the reason.

## 5. API surface implied by the model

The MVP will likely need these API areas:

- /auth
- /markets
- /prices
- /wallets
- /orders
- /positions
- /strategies
- /backtests
- /activity

## 6. Recommended backend architecture

### Go backend

Suggested layout:

- `cmd/api` for the HTTP entrypoint
- `internal/domain` for domain entities and rules
- `internal/service` for order execution, portfolios, strategies, and backtests
- `internal/store` for repository interfaces
- `internal/store/postgres` for PostgreSQL implementations
- `internal/http` for handlers, validation, and response mapping
- `internal/testutil` for fixtures and helpers

Important rule:
Order logic, position logic, PnL, and strategy rules must live in isolated services, not in handlers or SQL snippets.

### Persistence strategy

- target system: PostgreSQL
- development environment: local PostgreSQL
- all data access should go through an abstraction layer with explicit interfaces

Suggested approach:

- services depend on interfaces such as `OrderRepository`, `MarketRepository`, and `StrategyRepository`
- PostgreSQL implements those interfaces in `internal/store/postgres`
- SQL, migrations, and query details stay in the store layer
- domain logic stays testable with mocks, fakes, and test doubles

Useful repository interfaces:

- `UserRepository`
- `WalletRepository`
- `MarketRepository`
- `OrderRepository`
- `PositionRepository`
- `StrategyRepository`
- `BacktestRepository`

### Test strategy

- Go unit tests for:
  - order execution
  - fee and slippage calculations
  - position and PnL updates
  - strategy signal logic
  - backtesting calculations
- repository tests against a test database
- integration tests against a local PostgreSQL instance
- API tests for order creation and strategy activation flows

### Frontend recommendation

Suggested stack:

- Next.js with TypeScript
- React Testing Library for component tests
- Playwright for end-to-end tests
- Storybook later for visual component development if needed

UI direction:

- a distinct trading interface instead of a generic admin look
- a strong token-based visual system for color, spacing, type, and motion
- small, focused, testable components

## 7. Open decisions

- whether positions should use FIFO, average cost, or a hybrid model
- whether fees should always be simulated in the quote asset
- which timeframes should be persisted for candles
- whether real-time data should be cached before persistence
- how detailed backtest result storage should be
- which Go HTTP approach to use long term: stdlib only or a router package
- whether strategy workers should run inside the API binary or as a separate process
- whether local PostgreSQL should be native or container-based for the whole team

## 8. Recommended starting point

Keep the first build intentionally small:

- one demo wallet per user
- market orders first
- one execution per order
- candle data instead of tick-level data
- strategy configuration as JSON
- always log a human-readable reason
- put domain logic behind unit tests first
- start the frontend with a small but deliberate design system
- use local PostgreSQL with migrations from day one
- keep storage behind repository interfaces

That gives us a system that is simple enough for an MVP and still open for more assets, exchange connectors, and more advanced trading logic.
