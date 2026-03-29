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
- `Scope`: paper trading, dashboard UX, portfolio tracking, release automation
- `Not included`: live trading, custody, exchange account linking, financial advice

## Core capabilities

- multi-asset market list with `XRP/USDT` as the default reference flow
- demo-session based trading with isolated wallet state
- server-authoritative market-buy execution in the Go backend
- portfolio, balances, orders, positions, and activity history
- market candle rendering with bounded stale-feed fallback behavior
- Kubernetes deployment assets, CI validation, and release automation

## Product walkthrough

TradeLab currently supports a compact but realistic user journey:

1. open the app and create a demo session automatically
2. inspect the default market and feed state
3. switch intervals or markets without wiping the portfolio panels
4. execute a demo market buy
5. review balances, positions, orders, and activity in one screen

Visual walkthrough:

![Dashboard overview](docs/screenshots/dashboard-overview.png)
![Stale feed state](docs/screenshots/chart-stale-feed.png)
![Demo buy success](docs/screenshots/demo-buy-success.png)

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
  [docs/roadmap.md](docs/roadmap.md)
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

## Local development

### Parameters before you start

For a default local setup you usually do not need to set anything manually.

Only set parameters if you are deviating from the defaults:

| Parameter | When to set it | Default |
| --- | --- | --- |
| `DATABASE_URL` | before backend startup or migrations if your local PostgreSQL is not the default local instance | `postgres://tradelab:tradelab@localhost:5432/tradelab?sslmode=disable` |
| `HTTP_ADDRESS` | before backend startup if you want the API on a different port/address | `:8080` |
| `MARKET_DATA_BASE_URL` | before backend startup if you want another upstream market-data endpoint | `https://api.binance.com` |
| `TRADESLAB_API_PROXY_TARGET` | before frontend startup if the backend is not running on `http://localhost:8080` | `http://localhost:8080` |
| `NEXT_PUBLIC_API_BASE_URL` | before frontend startup only if the browser should call another API origin directly | empty |

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
