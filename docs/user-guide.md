# User Guide

## Purpose

This guide is for end users and first-time readers who want to understand how to use TradeLab as a demo trading sandbox.

TradeLab is a simulation-focused product. It helps users explore market flows, inspect portfolio changes, and test trading actions with virtual balances.

> [!IMPORTANT]
> TradeLab is demo-only software. It does not provide financial advice, investment recommendations, or live brokerage functionality.

## Product flow at a glance

The current user experience now uses two connected surfaces:

1. a demo session is created automatically
2. the overview dashboard loads with `XRP/USDT` as the reference pair
3. global accounting mode can be switched between `Average cost`, `FIFO`, and `Hybrid`
4. after the dashboard shows value, the user can choose to keep going as a guest or sign in with Google or Apple
5. the user can open a dedicated market detail page
6. the user can configure a rule-based strategy bundle for the selected market
7. the user can submit manual demo market buys and sells or let automation execute them
8. balances, positions, orders, activity, and PnL update after execution
9. the market detail page can replay the active strategy over a chosen historical range and compare recent backtest runs

## Dashboard overview

The dashboard is the overview surface:

- market selection
- quick buy ticket
- accounting mode switch
- wallet and PnL summary
- balances
- positions
- recent orders
- activity log

![TradeLab dashboard overview](screenshots/dashboard-overview.png)

## Market detail page

TradeLab now uses a dedicated market route for focused trading work.

The market detail page combines:

- live chart
- current position state
- demo buy ticket
- demo sell ticket
- strategy automation card
- filtered order history
- filtered activity for the selected market

![TradeLab market detail page](screenshots/market-detail-page.png)

## Core workflows

### 1. Start a demo session

When the app opens, TradeLab creates a demo session in the background and keeps it only for the current browser session.

What this enables:

- access to a demo wallet
- access to protected portfolio and order routes
- repeat use during the current browser session until it expires

### 2. Upgrade to a registered demo account

Once the dashboard has loaded, TradeLab can offer a durable-account prompt.

What this enables:

- Google or Apple sign-in through Clerk
- a durable registered demo account that can be restored across sessions
- an explicit choice to keep guest demo data or start with a fresh registered account

If you sign in from a guest session, the upgrade prompt asks whether to:

- keep guest demo data
- start fresh

Registered access uses a secure server-managed session after sign-in. The browser UI does not need a visible registered account token.

### 3. Inspect a market

Use the overview to choose a reference pair such as `XRP/USDT`, then open the dedicated market detail page. The chart updates independently from the rest of the product state so you can switch intervals without clearing portfolio panels.

What to watch:

- `Last close`
- `Feed state`
- `Current position`
- `Recent market-specific orders`
- `Recent market-specific activity`

### 4. Execute a demo market buy or sell

Use the market detail page to place virtual market trades:

1. choose a market
2. choose `buy` with a quote amount or `sell` with a base quantity
3. optionally use `Max position` for a full exit
4. submit the order

TradeLab executes the trade in the backend and then refreshes:

- wallet summary
- balances
- open positions
- recent orders
- activity log
- realized and unrealized PnL

![TradeLab demo sell success state](screenshots/demo-sell-success.png)

### 5. Configure strategy automation

TradeLab now supports one strategy bundle per wallet and market.

The automation card on the market detail page can manage:

- `Dip buy`
- `Take profit`
- `Stop loss`

What you can do:

1. adjust thresholds and spend size
2. save the bundle as a draft
3. activate the bundle
4. pause the bundle later
5. inspect the latest decision, outcome, and plain-English reason

Automated trades use the same demo wallet and order system as manual trades. In history they appear as `Automated` trades, and the activity log explains why the strategy fired.

![TradeLab automation active state](screenshots/strategy-automation-active.png)

### 6. Switch accounting modes

TradeLab applies one global accounting mode across overview and market detail:

- `Average cost`
- `FIFO`
- `Hybrid`

This lets you compare the same simulated trades under different valuation logic without changing the underlying demo trades themselves.

![TradeLab accounting mode switch](screenshots/accounting-mode-switch.png)

### 7. Review the outcome

After an order succeeds, use the overview and market detail together.

Use it to answer:

- did the wallet value change as expected
- did the position increase, shrink, or close
- was a new order recorded
- did realized PnL appear correctly on sells
- did the activity log explain what happened

### 8. Run a historical backtest

The market detail page now includes a read-only backtesting section.

What you can do:

1. choose a start date and end date
2. keep the current strategy bundle and interval
3. run a historical replay without touching the live demo wallet
4. inspect the result summary, equity curve, simulated trades, and recent run comparison table

This is useful when you want to compare how the same configuration would have behaved over different recent windows.

## Feed states

TradeLab currently exposes market-data freshness in the UI:

- `Fresh`: the backend fetched current market data successfully
- `Stale`: the backend is temporarily using a bounded cached fallback

The stale indicator helps users understand that the chart still works during short upstream problems, but the feed is no longer fully current.

## Current limitations

The current product is still intentionally narrow:

- no live trading
- no real-money wallets
- no exchange account linking
- no financial advice
- one automation bundle per market and wallet only
- no multi-strategy conflict management yet
- no live account linking beyond demo-only registered accounts

## Related documentation

- Repository overview:
  [README.md](../README.md)
- Getting started:
  [getting-started.md](getting-started.md)
- Product intent:
  [PRD.md](PRD.md)
- First-launch behavior:
  [onboarding-requirements.md](onboarding-requirements.md)
- Runtime and operations:
  [system-operations.md](system-operations.md)
- Developer onboarding:
  [developer-guide.md](developer-guide.md)
