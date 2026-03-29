# TradeLab

[![CI](https://github.com/aidun/tradelab/actions/workflows/ci.yml/badge.svg)](https://github.com/aidun/tradelab/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/aidun/tradelab?display_name=tag)](https://github.com/aidun/tradelab/releases)
[![License](https://img.shields.io/github/license/aidun/tradelab)](LICENSE)

TradeLab is a multi-asset crypto paper-trading platform for demo execution, strategy experimentation, and trading workflow validation. The project is being developed as a product-grade sandbox with clear engineering standards, reproducible releases, and a public roadmap.

> [!IMPORTANT]
> TradeLab is a demo-only software product. It does not provide financial advice, investment recommendations, or live trading guarantees. This repository also includes AI-assisted code and documentation.

## Why TradeLab exists

TradeLab is designed to make trading-system experimentation easier to understand and safer to validate.

It helps teams and individual builders:

- simulate market actions with virtual balances
- validate execution and portfolio flows before live integrations exist
- inspect trading outcomes through a transparent UI and auditable backend behavior
- evolve toward a production-grade product with strong testing, delivery, and operational discipline

## Current product status

- `Stage`: active demo sandbox
- `Reference market`: `XRP/USDT`
- `Scope`: paper trading, strategy automation, dashboard UX, portfolio tracking, release automation
- `Not included`: live trading, custody, exchange account linking, financial advice

## Core capabilities

- multi-asset market list with `XRP/USDT` as the default reference flow
- demo-session based trading with isolated wallet state
- Clerk-backed registered demo accounts with Google and Apple as the initial social-login targets
- explicit guest-to-registered upgrade flow with preserve-versus-fresh account choice
- HttpOnly registered app sessions plus session-scoped guest storage
- server-authoritative market buy and sell execution in the Go backend
- portfolio, balances, orders, positions, realized PnL, unrealized PnL, and activity history
- selectable accounting modes across the portfolio surface: `Average cost`, `FIFO`, and `Hybrid`
- dedicated market detail routes for focused trading flows
- rule-based strategy automation with one bundle per wallet and market
- built-in `dip_buy`, `take_profit`, and `stop_loss` automation in the market detail experience
- market candle rendering with bounded stale-feed fallback behavior
- Kubernetes deployment assets, CI validation, and release automation

## Product walkthrough

TradeLab currently supports a compact but realistic user journey:

1. open the app and create a guest demo session automatically
2. inspect the default market and feed state
3. optionally sign in with Google or Apple after first value appears
4. choose whether to preserve guest demo data or start with a fresh registered account
5. configure and activate a strategy bundle for a market
6. execute manual or automated demo buys and sells
7. switch accounting modes globally to inspect valuation and PnL
8. review balances, positions, orders, activity, and bot reasoning across overview and market detail screens

Visual walkthrough:

![Dashboard overview](docs/screenshots/dashboard-overview.png)
![Market detail](docs/screenshots/market-detail-page.png)
![Automation active](docs/screenshots/strategy-automation-active.png)
![Demo sell success](docs/screenshots/demo-sell-success.png)
![Accounting mode switch](docs/screenshots/accounting-mode-switch.png)

The full user-facing walkthrough lives in [docs/user-guide.md](docs/user-guide.md).

For setup and first successful run guidance by audience, start with [docs/getting-started.md](docs/getting-started.md).

## Quality and delivery

TradeLab is maintained with a product-style engineering workflow:

- backend tests via `go test ./...`
- frontend unit tests via Vitest
- frontend end-to-end coverage via Playwright
- container build validation for backend and frontend
- Kubernetes manifest rendering validation
- automated PR validation, squash merges, and release automation

For the exact PR -> CI -> merge -> release sequence, see:

- [docs/system-operations.md](docs/system-operations.md)
- [docs/release-process.md](docs/release-process.md)

## Public roadmap

TradeLab is intentionally developed in the open with a curated product roadmap.

- roadmap:
  [docs/roadmap.md](docs/roadmap.md)
- release process:
  [docs/release-process.md](docs/release-process.md)
- GitHub rollout checklist:
  [docs/github-rollout.md](docs/github-rollout.md)

## Documentation index

- users:
  [docs/getting-started.md](docs/getting-started.md),
  [docs/user-guide.md](docs/user-guide.md),
  [docs/onboarding-requirements.md](docs/onboarding-requirements.md)
- product and planning:
  [docs/PRD.md](docs/PRD.md),
  [docs/authentication-model.md](docs/authentication-model.md),
  [docs/roadmap.md](docs/roadmap.md),
  [docs/security-model.md](docs/security-model.md)
- developers:
  [docs/developer-guide.md](docs/developer-guide.md),
  [docs/data-model.md](docs/data-model.md)
- operators:
  [docs/system-operations.md](docs/system-operations.md),
  [docs/deployment.md](docs/deployment.md),
  [docs/installation-validation.md](docs/installation-validation.md),
  [docs/release-process.md](docs/release-process.md)
- machine-readable metadata:
  [docs/ai-metadata.json](docs/ai-metadata.json)

## Project standards

- contributing guide:
  [CONTRIBUTING.md](CONTRIBUTING.md)
- security policy:
  [SECURITY.md](SECURITY.md)
- code of conduct:
  [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- support guidance:
  [SUPPORT.md](SUPPORT.md)

## Identity roadmap

TradeLab currently uses guest demo sessions for the low-friction first run. The next identity phase is defined around Clerk-backed registered accounts with Google and Apple sign-in, while keeping guest mode as the entry path.

- [docs/authentication-model.md](docs/authentication-model.md)
- [docs/clerk-architecture.md](docs/clerk-architecture.md)
- [docs/auth-flows.md](docs/auth-flows.md)
- [docs/account-lifecycle.md](docs/account-lifecycle.md)

## Local development

### Parameters before you start

For a default local setup you usually do not need to set anything manually.

Only set parameters if you are deviating from the defaults:

| Parameter | When to set it | Default |
| --- | --- | --- |
| `DATABASE_URL` | before backend startup or migrations if your local PostgreSQL is not the default local instance | `postgres://tradelab:tradelab@localhost:5432/tradelab?sslmode=disable` |
| `HTTP_ADDRESS` | before backend startup if you want the API on a different port/address | `:8080` |
| `MARKET_DATA_BASE_URL` | before backend startup if you want another upstream market-data endpoint | `https://api.binance.com` |
| `TRADESLAB_CLERK_ISSUER_URL` | before backend startup if you want registered-account verification against Clerk instead of guest-only mode | empty |
| `TRADESLAB_CLERK_JWKS_URL` | before backend startup if you want registered-account verification against Clerk instead of guest-only mode | empty |
| `TRADESLAB_AUTH_MOCK_MODE` | before backend startup for local or CI auth mocking without live Clerk configuration | `false` |
| `STRATEGY_ENGINE_ENABLED` | before backend startup if you want to disable the in-process strategy engine locally | `true` |
| `STRATEGY_ENGINE_TICK` | before backend startup if you want a strategy-evaluation interval other than `60s` | `60s` |
| `TRADESLAB_API_PROXY_TARGET` | before frontend startup if the backend is not running on `http://localhost:8080` | `http://localhost:8080` |
| `NEXT_PUBLIC_API_BASE_URL` | before frontend startup only if the browser should call another API origin directly | empty |
| `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` | before frontend startup when you want real Clerk UI instead of guest-only or mock auth mode | empty |
| `NEXT_PUBLIC_AUTH_MOCK_MODE` | before frontend startup for local or CI auth mocking without live Clerk setup | `false` |

For the current auth/session security posture and sensitive-data boundaries, see [docs/security-model.md](docs/security-model.md).

For the full parameter matrix, including Kubernetes development and production values, see [docs/deployment.md](docs/deployment.md).

For the full first-run path and acceptance checklist, see [docs/getting-started.md](docs/getting-started.md) and [docs/installation-validation.md](docs/installation-validation.md).

### Start PostgreSQL

```bash
docker compose up -d postgres
```

### Apply database migrations

```bash
cd backend
go run ./cmd/migrate up
```

### Run the backend

```bash
cd backend
go run ./cmd/api
```

### Run the frontend

```bash
cd frontend
npm run dev
```

## Testing

### Backend

```bash
cd backend
go test ./...
```

### Frontend

```bash
cd frontend
npm run test
npm run build
npm run test:e2e
```

## Repository layout

- `frontend/`: web application
- `backend/`: API, domain logic, repositories, and migrations
- `deploy/`: Kubernetes manifests and release-render helpers
- `docs/`: product, operational, contributor, and machine-readable documentation
- `.github/`: workflow automation and GitHub contribution templates
