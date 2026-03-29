# PRD: TradeLab

## 1. Product overview

TradeLab is a multi-asset crypto paper-trading platform for manual demo trading, automated strategies, and backtesting. XRP is the reference asset for the first onboarding and showcase flows, but the platform architecture is intentionally asset-agnostic.

The MVP combines three workflows:

- manual paper trading with virtual funds
- automated demo trading with configurable strategies
- performance analysis with transparent trading logs

The MVP runs in demo mode only. Real exchange integrations are a later expansion and not part of the initial release.

## 2. Problem

Most existing tools only solve one part of the workflow:

- price-tracking apps without trading simulation
- demo trading tools without automation
- bot products with poor transparency
- advanced platforms that are too complex for beginners

TradeLab should let users learn, test, and improve strategies without risking real capital.

## 3. Target users

### Primary users

- beginners who want to learn trading with virtual money
- users who want to try automated crypto strategies
- traders who want to validate a setup before using real money

### Secondary users

- creators who want to present strategies
- communities that want to compare setups and outcomes

## 4. Product goals

- deliver a credible multi-asset paper-trading experience
- make bot strategies configurable without code
- explain every automated action in a human-readable way
- use XRP as a fast entry point without hard-coding the product to XRP

## 5. MVP non-goals

- no real exchange trading
- no custody or real funds
- no social trading
- no leverage or advanced derivatives
- no AI-driven trading engine in the MVP

## 6. Core assumptions

- users benefit from a default market such as XRP/USDT on first launch
- early multi-asset support increases credibility and reuse
- simple rules such as dip-buy, take-profit, and stop-loss are enough for the first version
- transparent trade reasoning is a trust requirement

## 7. MVP scope

### 7.1 Authentication

- sign up and login
- one personal demo account
- starting balance, for example 10,000 USDT

### 7.2 Market overview

- multi-coin watchlist
- initial assets: XRP, BTC, ETH, SOL, ADA
- price display and percentage change
- market search and filtering

### 7.3 Asset detail view

- price chart
- market information
- demo buy and sell module
- open positions and latest trades

### 7.4 Demo trading

- market buy
- market sell
- limit orders later
- simulated fees
- basic slippage rules

### 7.5 Bot strategies

- bind a strategy to a market such as XRP/USDT
- activate and pause strategies
- initial strategy types:
  - dip buy
  - take profit
  - stop loss
  - DCA
  - SMA crossover
  - RSI trigger

### 7.6 Portfolio

- total portfolio value
- allocation by asset
- realized and unrealized PnL
- performance over time

### 7.7 Activity and trade log

- persist every executed order
- capture the order origin:
  - manual
  - strategy
  - system
- show a plain-English reason for every bot-triggered trade

### 7.8 Backtesting

- load historical market data
- run a strategy against a selected range
- show:
  - return
  - hit rate
  - max drawdown
  - trade count

## 8. User flows

### Flow A: user tries XRP first

1. The user signs up.
2. The app opens on XRP/USDT.
3. The user sees price, chart, and demo balance.
4. The user buys XRP with virtual capital.
5. The portfolio and positions update.

### Flow B: user activates a bot

1. The user chooses XRP/USDT or BTC/USDT.
2. The user creates a dip-buy strategy.
3. The user defines rules and position size.
4. The strategy is activated.
5. The bot executes virtual orders when conditions match.
6. The app shows logs and strategy performance.

### Flow C: user runs a backtest

1. The user picks a market and a date range.
2. The user chooses a strategy.
3. The backtest simulates trades.
4. The user compares outcome and risk.

## 9. Functional requirements

### Must-have

- support multiple assets and markets
- provide a demo wallet per user
- execute manual demo orders
- store and evaluate strategies
- store trades and positions with auditability
- show performance and activity logs

### Should-have

- basic backtesting
- real-time or near-real-time prices
- watchlist and market search

### Nice-to-have

- notifications
- strategy templates
- side-by-side strategy comparison

## 10. Quality requirements

- Auditability: every system decision must be traceable
- Extensibility: new assets and strategy types should not require a redesign
- Credibility: fees and slippage must not be ignored
- Security: the MVP should not require real exchange API keys
- Usability: a beginner should place a first demo order within minutes
- Testability: backend and frontend logic must be automated from the start
- Visual quality: the UI should feel premium and intentional, not like a generic dashboard

## 11. Technical guardrails

### Backend

- language: Go
- architecture: modular API with clear domains for markets, orders, strategies, portfolios, and backtests
- target database: PostgreSQL
- development database: local PostgreSQL
- key requirements:
  - strong unit-testability
  - strict separation of HTTP, domain logic, and persistence
  - data access behind an abstraction layer
  - future support for background workers

### Frontend

- technology is flexible as long as testability and visual quality stay first-class
- recommended MVP stack:
  - Next.js
  - TypeScript
  - Tailwind CSS or a token-based CSS system
  - component-driven architecture
- UX target:
  - modern trading UI with a distinct visual identity
  - responsive on desktop and mobile
  - clear hierarchy despite dense market data

### Testing is mandatory

- unit tests for Go domain logic
- API tests for critical endpoints
- component tests in the frontend
- end-to-end tests for the primary user journeys
- repository and integration tests against a local PostgreSQL database
- critical coverage for:
  - order execution
  - portfolio calculation
  - strategy signals
  - backtesting outcomes
  - essential UI flows

## 12. MVP KPIs

- percentage of users who place a first demo order
- percentage of users who activate at least one strategy
- average number of active strategies per user
- 7-day retention
- number of demo trades per active user

## 13. Risks

- poor market data quality could produce unrealistic simulation
- overly complex strategy setup could overwhelm beginners
- unrealistic fills could distort later expectations for live trading
- weak transparency around bot actions would reduce trust
- insufficient test coverage in trading logic could hide domain bugs
- a bland UI could make the product feel interchangeable despite strong functionality

## 14. Roadmap after MVP

### Phase 2

- limit orders
- improved backtesting
- notifications
- strategy templates

### Phase 3

- multiple paper-trading accounts
- advanced risk controls
- comparative strategy reports

### Phase 4

- optional real exchange integration
- API key management
- live trading only with explicit safety controls
