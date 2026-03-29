# Getting Started

## Purpose

This guide explains how to get TradeLab running for three different audiences:

- developers running the product locally
- operators validating a deployment target
- first-time product reviewers opening the app for the first time

TradeLab is currently a demo-only trading sandbox. The first-run experience should always make that clear.

## Choose your path

### Local developer setup

Use this path when you want to run the full stack on your machine and work on code.

You will typically need:

- Docker or a local PostgreSQL instance
- Go
- Node.js and npm

Follow this order:

1. start PostgreSQL
2. apply database migrations
3. start the backend
4. start the frontend
5. open the app and confirm that a demo session is created

The detailed local parameter reference lives in [deployment.md](deployment.md).

### Operator deployment setup

Use this path when you want to validate or run TradeLab in Kubernetes.

You will typically need:

- a Kubernetes cluster
- `kubectl` with kustomize support
- ingress support
- access to GHCR images
- the required secret values for the chosen environment

Follow this order:

1. prepare environment-specific secrets and hosts
2. render or apply the Kubernetes overlay
3. verify backend health, frontend reachability, and session creation
4. confirm the first protected API flow works

The deployment-specific steps live in [deployment.md](deployment.md), and the runtime model lives in [system-operations.md](system-operations.md).

### First-time product review

Use this path when you are evaluating the product rather than modifying or operating it.

What you should expect on first launch:

1. the app loads the dashboard
2. a guest demo session is created automatically
3. the market list appears with `XRP/USDT` as the reference pair
4. the chart loads with a feed freshness indicator and a global accounting-mode switch
5. portfolio panels, recent orders, activity, and PnL become visible
6. market cards can open a focused market detail route
7. after first value appears, the app can offer a durable-account upgrade through Google or Apple login

The full walkthrough is in [user-guide.md](user-guide.md).

## First local run

### Step 1. Start PostgreSQL

```bash
docker compose up -d postgres
```

### Step 2. Apply migrations

```bash
cd backend
go run ./cmd/migrate up
```

### Step 3. Start the backend

```bash
cd backend
go run ./cmd/api
```

### Step 4. Start the frontend

```bash
cd frontend
npm run dev
```

### Step 5. Open the app

Open `http://localhost:3000` and confirm the dashboard loads.

On a successful first run:

- the frontend loads without a blank error screen
- a demo session is created automatically
- markets are visible
- the chart renders candles
- the wallet panels load

## Configuration decisions before first run

Most local runs can use the defaults.

Only set configuration values when you intentionally deviate from the default topology:

| Parameter | When to set it | Typical reason |
| --- | --- | --- |
| `DATABASE_URL` | before migrations or backend startup | your PostgreSQL host, port, user, password, or DB name differs |
| `HTTP_ADDRESS` | before backend startup | you want the API on another address or port |
| `MARKET_DATA_BASE_URL` | before backend startup | you want another upstream market-data base URL |
| `TRADESLAB_CLERK_ISSUER_URL` | before backend startup | you want real Clerk-backed registered-account verification |
| `TRADESLAB_CLERK_JWKS_URL` | before backend startup | you want real Clerk-backed registered-account verification |
| `TRADESLAB_AUTH_MOCK_MODE` | before backend startup | you want local or CI auth mocking instead of live Clerk verification |
| `TRADESLAB_API_PROXY_TARGET` | before frontend startup | the frontend should proxy `/api` to a non-default backend origin |
| `NEXT_PUBLIC_API_BASE_URL` | before frontend startup | the browser should call a different API origin directly |
| `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` | before frontend startup | you want the real Clerk UI and social-login surface |
| `NEXT_PUBLIC_AUTH_MOCK_MODE` | before frontend startup | you want local or CI auth mocking instead of live Clerk UI |

For the complete matrix, see [deployment.md](deployment.md).

## First-run acceptance checklist

Use this checklist after setup:

- PostgreSQL is reachable
- migrations complete successfully
- backend responds at `/healthz`
- frontend renders the dashboard
- a demo session is created automatically
- the durable-account prompt appears after the dashboard has loaded when auth is available
- the market list loads
- the chart loads candle data
- portfolio panels load successfully
- a demo buy can be submitted and reflected in the dashboard
- a market detail page opens for the selected pair
- a demo sell can be submitted from the market detail page

The validation details for local and deployed environments live in [installation-validation.md](installation-validation.md).

## Related documentation

- [README.md](../README.md)
- [deployment.md](deployment.md)
- [developer-guide.md](developer-guide.md)
- [system-operations.md](system-operations.md)
- [user-guide.md](user-guide.md)
- [installation-validation.md](installation-validation.md)
- [onboarding-requirements.md](onboarding-requirements.md)
