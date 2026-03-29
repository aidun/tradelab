# TradeLab

TradeLab is a multi-asset crypto paper-trading platform with automated strategy support, backtesting, and a polished web experience. XRP is the reference market for the first user flows, but the platform is designed for multiple assets from day one.

## Stack

- Frontend: Next.js, TypeScript, Tailwind CSS
- Backend: Go
- Database: PostgreSQL
- Testing: Go test, Vitest, React Testing Library, Playwright

## Repository layout

- `frontend/` web application
- `backend/` API, domain logic, repositories, migrations
- `deploy/` Kubernetes deployment manifests and overlays
- `docs/` product and architecture documentation

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

### Build container images locally

```bash
docker build -f backend/Dockerfile -t tradelab-backend:local .
docker build -f frontend/Dockerfile -t tradelab-frontend:local .
```

## Testing

### Backend

```bash
cd backend
go test ./...
```

### Demo seed identifiers

- Demo user ID: `cfbf7c8f-eaf9-47fa-8674-2a29fed1fcc9`
- Demo wallet ID: `1ddb1c1c-827f-4bf0-b85a-3d5786c3b26c`

### Sample API calls

```bash
curl http://localhost:8080/api/v1/markets
```

```bash
curl http://localhost:8080/api/v1/portfolios/1ddb1c1c-827f-4bf0-b85a-3d5786c3b26c
```

```bash
curl "http://localhost:8080/api/v1/orders?wallet_id=1ddb1c1c-827f-4bf0-b85a-3d5786c3b26c"
```

```bash
curl "http://localhost:8080/api/v1/activity?wallet_id=1ddb1c1c-827f-4bf0-b85a-3d5786c3b26c"
```

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "cfbf7c8f-eaf9-47fa-8674-2a29fed1fcc9",
    "wallet_id": "1ddb1c1c-827f-4bf0-b85a-3d5786c3b26c",
    "market_symbol": "XRP/USDT",
    "quote_amount": 50,
    "expected_price": 0.67
  }'
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

## Kubernetes deployment

TradeLab ships with a complete Kubernetes deployment under `deploy/kubernetes`.

Quick start for the development overlay:

```bash
kubectl apply -k deploy/kubernetes/overlays/development
```

This deploys PostgreSQL, the Go backend, the Next.js frontend, and an ingress that routes `/api` to the backend and `/` to the frontend.

Full deployment notes live in [docs/deployment.md](docs/deployment.md).

## Delivery automation

- Pull requests are validated by GitHub Actions.
- Successful PR checks can be auto-merged into `master`.
- Every successful `master` run builds release artifacts, publishes container images to GitHub Container Registry, and creates a GitHub Release.
