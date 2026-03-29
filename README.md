# TradeLab

TradeLab is a multi-asset crypto paper-trading platform for demo execution, strategy experimentation, and trading workflow validation. XRP is the reference market for the first experience, but the platform is intentionally built for multiple assets.

> [!IMPORTANT]
> This repository includes AI-assisted code and documentation. TradeLab is an educational and simulation-oriented software project. It does not provide financial advice, investment recommendations, or trading guarantees, and it must not be treated as a substitute for independent financial decision-making.

## Product overview

TradeLab is designed to help users:

- explore crypto markets with virtual balances
- test execution and portfolio flows without risking real funds
- understand automated strategy behavior in a transparent UI
- evolve toward a production-grade trading sandbox with strong testing and operational discipline

The current system is still a demo environment. It is not a live brokerage, not a custody product, and not a managed investment service.

## Current capabilities

- multi-asset market list with XRP as the default showcase market
- demo-session based trading flow with isolated demo wallets
- server-authoritative market-buy execution in the Go backend
- portfolio, balances, orders, and activity history
- market candle rendering with cached market-data fallback behavior
- Kubernetes deployment assets and release automation

## Quick links

- Users and first-time readers:
  [README](README.md),
  [PRD.md](docs/PRD.md)
- Operators:
  [system-operations.md](docs/system-operations.md),
  [deployment.md](docs/deployment.md)
- Developers:
  [developer-guide.md](docs/developer-guide.md),
  [data-model.md](docs/data-model.md)
- AI tooling / structured metadata:
  [ai-metadata.json](docs/ai-metadata.json)

## Repository layout

- `frontend/` web application
- `backend/` API, domain logic, repositories, and migrations
- `deploy/` Kubernetes manifests and release-render helpers
- `docs/` product, operational, developer, and machine-readable documentation

## Local development

### Start PostgreSQL

```bash
docker compose up -d postgres
```

### Run the backend

```bash
cd backend
go run ./cmd/api
```

### Apply database migrations

```bash
cd backend
go run ./cmd/migrate up
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

### Frontend unit tests

```bash
cd frontend
npm run test
```

### Frontend end-to-end tests

```bash
cd frontend
npm run test:e2e
```

## Sample API flow

Create a demo session:

```bash
curl -X POST http://localhost:8080/api/v1/sessions/demo
```

Use the returned bearer token to access protected routes:

```bash
curl http://localhost:8080/api/v1/markets
```

```bash
curl "http://localhost:8080/api/v1/markets/XRP%2FUSDT/candles?interval=1h&limit=48"
```

Use the `wallet_id` returned by the demo session for portfolio access:

```bash
curl http://localhost:8080/api/v1/portfolios/<wallet-id> \
  -H "Authorization: Bearer <demo-token>"
```

```bash
curl http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer <demo-token>"
```

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer <demo-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "market_symbol": "XRP/USDT",
    "quote_amount": 50
  }'
```

## Deployment and delivery

TradeLab ships with Kubernetes deployment assets, immutable release-manifest rendering, GitHub Actions CI, and GitHub release automation.

- Deployment quick start:
  [deployment.md](docs/deployment.md)
- Runtime and operations guide:
  [system-operations.md](docs/system-operations.md)
- Contributor workflow:
  [developer-guide.md](docs/developer-guide.md)

## Delivery automation

- pull requests are validated by GitHub Actions
- successful PR checks can be auto-merged into `master`
- successful `master` runs publish release artifacts, container images, and a GitHub Release
