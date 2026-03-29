# Developer Guide

## Purpose

This guide explains how to work in the TradeLab repository safely and efficiently as a contributor.

## Core references

- Product intent:
  [PRD.md](PRD.md)
- Data model:
  [data-model.md](data-model.md)
- Runtime and operations:
  [system-operations.md](system-operations.md)
- Deployment details:
  [deployment.md](deployment.md)
- Machine-readable repo metadata:
  [ai-metadata.json](ai-metadata.json)

## Required tools

- Go
- Node.js and npm
- PostgreSQL or Docker for local PostgreSQL
- Git
- optional but useful:
  - `kubectl` with kustomize support
  - GitHub CLI

## Local workflow

### Start dependencies

```bash
docker compose up -d postgres
```

### Run backend

```bash
cd backend
go run ./cmd/api
```

### Run migrations

```bash
cd backend
go run ./cmd/migrate up
```

### Run frontend

```bash
cd frontend
npm run dev
```

## Test commands

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

### Optional deployment validation

```bash
kubectl kustomize deploy/kubernetes/overlays/development
kubectl kustomize deploy/kubernetes/overlays/production
```

## Architecture map

### Backend

- `backend/cmd/api`: API entrypoint
- `backend/cmd/migrate`: migration entrypoint
- `backend/internal/domain`: domain types
- `backend/internal/service`: business logic
- `backend/internal/store`: repository interfaces
- `backend/internal/store/postgres`: PostgreSQL implementations
- `backend/internal/http`: routing and API response shaping

### Frontend

- `frontend/app`: app entrypoints
- `frontend/components`: UI components, including the trading dashboard
- `frontend/lib`: API client code and shared helpers
- `frontend/__tests__`: component and workflow tests

### Delivery and deployment

- `.github/workflows`: CI, auto-merge, and release automation
- `deploy/kubernetes`: base manifests, overlays, and release rendering helper

## Contribution model

TradeLab currently follows a PR-first workflow:

- changes are grouped into focused branches and pull requests
- CI must pass before merge
- merges are expected to remain reviewable and reasonably scoped
- `master` is the release branch
- successful `master` runs produce release artifacts and published container images

## Contribution expectations

When making changes:

- keep code and documentation in English
- add or update tests when behavior changes
- keep structured logging and operational visibility aligned with the runtime behavior when backend flows change
- keep documentation aligned with implementation, especially API and deployment behavior
- update `.github/workflows` when quality gates or delivery behavior need to change
- reference the relevant issue in the PR where possible
- prefer updating existing docs over creating redundant parallel explanations

## Documentation expectations

- the root README is for repository landing-page readers
- `docs/system-operations.md` is the runtime and operator source of truth
- `docs/developer-guide.md` is the contributor source of truth
- `docs/ai-metadata.json` exists for machine consumption and should be updated when the human-facing structure materially changes
- logging, tests, documentation, and GitHub Actions are treated as part of the feature surface and should be adjusted together when needed

## Notes for future contributors

- market-data behavior now includes bounded stale fallback semantics
- backend request and service flow now emits structured JSON logs through `log/slog`
- frontend quality gates now include Playwright coverage for core dashboard journeys
- release-ready Kubernetes output should use immutable release tags, not rely on `latest`
- protected API routes depend on demo-session bearer tokens
