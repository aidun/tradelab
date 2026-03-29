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
