# User Guide

## Purpose

This guide is for end users and first-time readers who want to understand how to use TradeLab as a demo trading sandbox.

TradeLab is a simulation-focused product. It helps users explore market flows, inspect portfolio changes, and test trading actions with virtual balances.

> [!IMPORTANT]
> TradeLab is demo-only software. It does not provide financial advice, investment recommendations, or live brokerage functionality.

## Product flow at a glance

The current user experience centers on one dashboard:

1. a demo session is created automatically
2. the market list loads with `XRP/USDT` as the reference pair
3. live candle data and feed status appear in the chart area
4. the user can switch intervals or markets
5. the user can submit a demo market buy
6. balances, positions, orders, and activity update after execution

## Dashboard overview

The dashboard combines the main trading surfaces in one place:

- market selection
- demo buy ticket
- live chart
- wallet summary
- balances
- positions
- recent orders
- activity log

![TradeLab dashboard overview](screenshots/dashboard-overview.png)

## Core workflows

### 1. Start a demo session

When the app opens, TradeLab creates a demo session in the background and stores it locally in the browser for the current user journey.

What this enables:

- access to a demo wallet
- access to protected portfolio and order routes
- repeat visits without re-creating the session until it expires

### 2. Inspect a market

Use the market list to choose a reference pair such as `XRP/USDT`. The chart updates independently from the rest of the dashboard so you can switch intervals without clearing portfolio panels.

What to watch:

- `Last close`
- `Feed state`
- `Session high`
- `Session low`
- `Last volume`

![TradeLab stale-feed interval view](screenshots/chart-stale-feed.png)

### 3. Execute a demo market buy

Use the demo buy ticket to place a virtual market buy:

1. choose a market
2. enter a quote amount
3. review the server-side execution pricing hint
4. submit the order

TradeLab executes the buy in the backend and then refreshes:

- wallet summary
- balances
- open positions
- recent orders
- activity log

![TradeLab demo buy success state](screenshots/demo-buy-success.png)

### 4. Review the outcome

After an order succeeds, the dashboard becomes the main review surface.

Use it to answer:

- did the wallet value change as expected
- did a new position open
- was a new order recorded
- did the activity log explain what happened

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
- no strategy automation in the user workflow yet

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
