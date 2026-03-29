# Developer Guide

## Purpose

This guide explains how to work in the TradeLab repository safely and efficiently as a contributor.

## Core references

- Product intent:
  [PRD.md](PRD.md)
- Public roadmap:
  [roadmap.md](roadmap.md)
- Authentication model:
  [authentication-model.md](authentication-model.md)
- Data model:
  [data-model.md](data-model.md)
- Runtime and operations:
  [system-operations.md](system-operations.md)
- Deployment details:
  [deployment.md](deployment.md)
- Infrastructure bootstrap:
  [infrastructure-bootstrap.md](infrastructure-bootstrap.md)
- Release process:
  [release-process.md](release-process.md)
- GitHub rollout:
  [github-rollout.md](github-rollout.md)
- First run and onboarding:
  [getting-started.md](getting-started.md)
- Installation validation:
  [installation-validation.md](installation-validation.md)
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

### Configuration checkpoints

Before starting local services, review these only if you are not using the defaults:

- `DATABASE_URL` before backend startup or migrations
- `HTTP_ADDRESS` before backend startup
- `MARKET_DATA_BASE_URL` before backend startup
- `TRADESLAB_CLERK_ISSUER_URL` before backend startup when testing real Clerk-backed registered accounts
- `TRADESLAB_CLERK_JWKS_URL` before backend startup when testing real Clerk-backed registered accounts
- `TRADESLAB_AUTH_MOCK_MODE` before backend startup when local or CI auth mocking should replace live Clerk verification
- `STRATEGY_ENGINE_ENABLED` before backend startup when local strategy automation should be disabled explicitly
- `STRATEGY_ENGINE_TICK` before backend startup when the strategy loop should use a non-default interval
- `TRADESLAB_API_PROXY_TARGET` before frontend startup
- `NEXT_PUBLIC_API_BASE_URL` before frontend startup if you want direct browser-to-API calls
- `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` before frontend startup when using real Clerk UI
- `NEXT_PUBLIC_AUTH_MOCK_MODE` before frontend startup when local or CI auth mocking should replace live Clerk UI

The full environment and deployment parameter reference lives in [deployment.md](deployment.md).

### Start dependencies

```bash
docker compose up -d postgres
```

### Run migrations

```bash
cd backend
go run ./cmd/migrate up
```

### Run backend

```bash
cd backend
go run ./cmd/api
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
npm ci
npm run test
npm run build
npm run test:e2e
npm run docs:screenshots
```

The frontend uses [frontend/.npmrc](../frontend/.npmrc) to keep npm resolution aligned across local runs, GitHub Actions, and Docker builds while the current Clerk dependency chain still requires `legacy-peer-deps`.

### Optional deployment validation

```bash
kubectl kustomize deploy/infrastructure/bootstrap/argocd-install
kubectl kustomize deploy/infrastructure/bootstrap/root-application
kubectl kustomize deploy/infrastructure/applications
kubectl kustomize deploy/infrastructure/platform/namespaces
kubectl kustomize deploy/infrastructure/platform/metallb/config
kubectl kustomize deploy/kubernetes/overlays/development
kubectl kustomize deploy/kubernetes/overlays/production
kubectl kustomize deploy/kubernetes/overlays/production-external-secrets
```

## Architecture map

### Backend

- `backend/cmd/api`: API entrypoint
- `backend/cmd/migrate`: migration entrypoint
- `backend/internal/domain`: domain types
- `backend/internal/service`: business logic
- `backend/internal/service/strategy`: strategy bundle lifecycle and in-process automation engine
- `backend/internal/store`: repository interfaces
- `backend/internal/store/postgres`: PostgreSQL implementations
- `backend/internal/http`: routing and API response shaping

### Frontend

- `frontend/app`: app entrypoints
- `frontend/components`: UI components, including the overview dashboard, focused market detail screen, and automation card
- `frontend/lib`: API client code and shared helpers
- `frontend/lib/tradelab-auth.tsx`: auth provider boundary for guest, mock, and Clerk-backed registered modes
- `frontend/lib/use-account-session.ts`: guest-plus-registered account orchestration plus guest session refresh logic
- `frontend/__tests__`: component and workflow tests
- `frontend/scripts`: utility scripts, including documentation screenshot generation

### Delivery and deployment

- `.github/workflows`: CI, auto-merge, and release automation
- `deploy/infrastructure`: Argo CD bootstrap, platform apps, and MetalLB configuration
- `deploy/kubernetes`: base manifests, overlays, and release rendering helper

## Contribution model

TradeLab currently follows a PR-first workflow:

- changes are grouped into focused branches and pull requests
- CI must pass before merge
- merges are expected to remain reviewable and reasonably scoped
- `master` is the integration branch for development and release preparation
- if `master` moves while a PR is open, the PR branch must be rebased or otherwise updated to the current `master` state before final merge or auto-merge
- official releases are triggered manually from GitHub Actions
- production promotion is a separate manual workflow that moves Argo CD to an official release tag

## GitHub Actions flow

TradeLab uses five workflows:

1. `CI`
2. `Auto Merge PR`
3. `Publish Master Images`
4. `Release`
5. `Promote Production`

### CI

The CI workflow runs these jobs in parallel:

- `Backend tests`
- `Frontend unit tests`
- `Frontend build`
- `Frontend E2E tests`
- `Backend container build`
- `Frontend container build`
- `Kubernetes manifests`
- `Metadata validation`

### Auto Merge PR

When `CI` finishes successfully for a pull request targeting `master`, the auto-merge workflow performs a `squash` merge.

If `master` advanced after the PR branch was created, refresh the PR branch first so the final merge happens against the current integration state instead of an outdated head.

### Release

The release workflow is started manually from `master` and runs in this order:

1. `Release metadata`
2. `Verify backend` and `Verify frontend`
3. `Build backend binaries`, `Build frontend artifact`, `Publish backend image`, and `Publish frontend image`
4. `Package Kubernetes manifests`
5. `Create GitHub release`

### Promote Production

The production promotion workflow is also manual.

It resolves the requested or latest official GitHub release and updates the production Argo CD application to that release tag.

### Publish Master Images

Every merge to `master` builds and publishes backend and frontend development images with both:

- `master`
- `master-<shortsha>`

It then commits the exact deployed development revision back into [tradelab-dev.yaml](../deploy/infrastructure/applications/tradelab-dev.yaml), setting:

- the Argo CD `targetRevision` to the source commit SHA
- backend and frontend image overrides to the matching immutable `master-<shortsha>` tags

This is the expected delivery chain for normal feature work:

`feature branch -> pull request -> CI -> auto-merge -> master -> publish master images and update Argo dev -> manual release -> manual production promotion`

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
- `docs/getting-started.md` is the audience-aware entrypoint for setup and first-run success
- `docs/installation-validation.md` defines the required smoke path for local and deployed environments
- `docs/onboarding-requirements.md` captures the intended guest-first product onboarding behavior
- `docs/authentication-model.md`, `docs/clerk-architecture.md`, `docs/auth-flows.md`, and `docs/account-lifecycle.md` define the implemented guest-plus-registered identity surface
- `docs/security-model.md` is the source of truth for session, secret, redaction, and sensitive-data boundaries
- `docs/system-operations.md` is the runtime and operator source of truth
- `docs/infrastructure-bootstrap.md` is the source of truth for the k3s platform bootstrap and Argo CD app-of-apps layout
- application secrets should default to generated bootstrap values when no secret exists, not to committed credentials
- `docs/developer-guide.md` is the contributor source of truth
- `docs/user-guide.md` is the user-facing walkthrough with screenshots
- `docs/roadmap.md` is the public product direction summary
- `docs/release-process.md` describes release artifacts and workflow meaning
- `docs/github-rollout.md` captures manual GitHub repository settings and presentation steps
- `docs/ai-metadata.json` exists for machine consumption, acts as the dependency map for repo-facing follow-up work, and should be updated when the human-facing structure materially changes
- logging, tests, documentation, and GitHub Actions are treated as part of the feature surface and should be adjusted together when needed
- development delivery must keep Git revision, Argo targetRevision, and GHCR image tags aligned; do not reintroduce mutable `latest`-based dev deployment

## AI metadata contract

`docs/ai-metadata.json` now does two jobs:

- it describes the repository and its public engineering surface
- it defines the dependency rules for follow-up work when something changes

When making changes, consult the `artifact_groups` and `change_management` sections to decide which documentation, screenshots, tests, workflow files, or GitHub-facing artifacts must also be reviewed or updated.

## Notes for future contributors

- market-data behavior now includes bounded stale fallback semantics
- backend request and service flow now emits structured JSON logs through `log/slog`
- frontend quality gates now include Playwright coverage for core dashboard journeys
- Phase 4 adds a global accounting-mode UI and a dedicated `/markets/[symbol]` route for the focused trading flow
- Phase 5 adds a strategy bundle per `wallet + market` with in-process evaluation and strategy-specific order/activity visibility
- release-ready Kubernetes output should use immutable release tags, not rely on `latest`
- protected API routes now accept guest bearer tokens or registered HttpOnly app-session cookies, depending on the principal type
