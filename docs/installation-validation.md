# Installation Validation

## Purpose

This document defines the expected smoke-test path for a correct TradeLab installation in local and deployed environments.

The goal is simple: a new engineer or operator should be able to validate a first successful run without guessing what "working" means.

## Local validation

### Preconditions

- PostgreSQL is running
- database migrations were applied successfully
- backend is running
- frontend is running

### Required checks

1. `GET /healthz` returns a successful response from the backend.
2. The frontend root page loads at the expected local address.
3. Opening the app creates a guest demo session automatically.
4. The market list renders at least one market and shows `XRP/USDT` as the default reference pair.
5. The chart renders candle data for the selected market.
6. Portfolio panels load without authentication or data errors.
7. A demo market buy succeeds and creates visible changes in:
   - wallet summary
   - balances
   - positions
   - recent orders
   - activity

### Local validation commands

Backend health:

```bash
curl http://localhost:8080/healthz
```

Manual demo session creation:

```bash
curl -X POST http://localhost:8080/api/v1/sessions/demo
```

Protected portfolio access with the returned token:

```bash
curl http://localhost:8080/api/v1/portfolios/<wallet-id> \
  -H "Authorization: Bearer <demo-token>"
```

Demo buy:

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer <demo-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "market_symbol": "XRP/USDT",
    "quote_amount": 50
  }'
```

### Local success criteria

The installation is considered valid when:

- backend health is green
- frontend loads the dashboard
- session creation works automatically or manually
- protected API access succeeds with the session token
- a demo buy updates the visible dashboard state

## Deployment validation

### Preconditions

- Kubernetes manifests were applied successfully
- required secrets are present for the target environment
- ingress host routing is configured correctly

### Required checks

1. Backend pods are healthy.
2. Frontend pods are healthy.
3. PostgreSQL is healthy and reachable from the backend.
4. Migration init container completed successfully.
5. The public frontend loads.
6. The frontend can reach `/api`.
7. A guest demo session is created automatically in the browser.
8. Protected portfolio routes succeed after session creation.
9. A demo buy succeeds and updates the user-facing dashboard state.

### Deployment success criteria

The deployment is considered valid when:

- all runtime components are healthy
- the first protected product flow works end-to-end
- no configuration, secret, or ingress mismatch blocks the demo experience

## Failure classification

Use this quick classification when validation fails:

| Symptom | First place to check |
| --- | --- |
| backend health fails | backend logs, pod status, database connectivity |
| frontend loads but `/api` fails | ingress routing, frontend proxy config, backend service |
| demo session fails | backend logs, database migrations, session tables |
| chart fails | backend market-data logs, upstream reachability, stale fallback state |
| portfolio or orders fail | auth headers, session validity, backend logs, database state |

## Related documentation

- [getting-started.md](getting-started.md)
- [deployment.md](deployment.md)
- [system-operations.md](system-operations.md)
- [user-guide.md](user-guide.md)
