# System Operations

## Purpose

This document explains how TradeLab is operated as a running system: which components exist, how they are configured, how releases move through the system, and what an operator should check first when something goes wrong.

For audience-specific first setup and smoke validation, see [getting-started.md](getting-started.md) and [installation-validation.md](installation-validation.md).

## Runtime topology

TradeLab currently consists of these runtime components:

- `frontend`: Next.js web application served as the user-facing interface
- `backend`: Go HTTP API that owns session handling, portfolio logic, order execution, and market-data access
- `strategy engine`: in-process scheduler inside the backend that evaluates active strategy bundles
- `postgres`: PostgreSQL database for persistent application state
- `traefik`: managed ingress controller that routes `/` to the frontend and `/api` to the backend
- `metallb`: assigns LAN LoadBalancer IPs to cluster entry services such as Traefik
- `argocd`: GitOps controller for platform and application synchronization
- `migration initContainer`: applies schema migrations before the backend starts
- `GitHub Actions + GHCR`: build, verify, publish, and package release artifacts

## Environments

### Development

- Kubernetes namespace: `tradelab-dev`
- image tags: `latest`
- Argo CD source revision: `master`
- database secret source: generated bootstrap secret unless `tradelab-database` already exists
- ingress class: `traefik`
- default development ingress entrypoint: `http://192.168.2.200/tradelab-dev`
- prepared production ingress entrypoint: `http://192.168.2.200/tradelab`
- intended for fast local or shared-dev iteration

### Production

- Kubernetes namespace: `tradelab`
- image tags in release manifests: immutable release tag such as `v0.1.<run-number>`
- Argo CD source revision: latest promoted official release tag
- database secret source: generated bootstrap secret by default, or external secrets through the optional production-external-secrets overlay
- intended for controlled deployment via Argo CD promotion after a manual release

## Configuration and secrets

### Backend

Important backend configuration includes:

- `HTTP_ADDRESS`
- `DATABASE_URL`
- `MARKET_DATA_BASE_URL`
- `TRADESLAB_CLERK_ISSUER_URL`
- `TRADESLAB_CLERK_JWKS_URL`
- `TRADESLAB_AUTH_MOCK_MODE`
- `STRATEGY_ENGINE_ENABLED`
- `STRATEGY_ENGINE_TICK`

TradeLab now treats external authentication as a separate trust boundary from application trading sessions. See [authentication-model.md](authentication-model.md) and [clerk-architecture.md](clerk-architecture.md).

### Frontend

Important frontend configuration includes:

- `NEXT_PUBLIC_API_BASE_URL`
- `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY`
- `NEXT_PUBLIC_AUTH_MOCK_MODE`
- any same-origin or proxy alignment needed for local/runtime routing

### Secrets

Secret-bearing values should not be committed directly into base manifests.

- development secrets come from the generated bootstrap secret unless `tradelab-database` already exists
- generated Kubernetes secrets are used as the first-run fallback when `tradelab-database` is absent
- production can later switch to the configured external secret store through `deploy/kubernetes/overlays/production-external-secrets`
- database credentials and `DATABASE_URL` are the primary required runtime secrets today
- Clerk publishable keys are frontend-safe, but Clerk secret material must remain server-side only
- mock-auth flags are development and CI-only controls and must not be treated as production runtime settings
- strategy-engine settings are operational configuration, not secrets

## Deployment flow

### Development deploy

1. Bootstrap the cluster platform through [infrastructure-bootstrap.md](infrastructure-bootstrap.md) if Traefik, MetalLB, and Argo CD are not already present.
2. Render or apply the development overlay.
3. Verify the generated or pre-created `tradelab-database` secret.
4. Verify ingress, backend health, and frontend availability.

### Production deploy

1. Decide whether production should use generated first-run credentials or the external-secret overlay.
2. Trigger the manual release workflow from `master`.
3. Confirm the GitHub Release exists and immutable images were published.
4. Trigger the `Promote Production` workflow to move `tradelab-prod` to the selected official release tag.
5. Verify backend health, frontend reachability, and database connectivity.

## Release model

TradeLab uses a PR-first delivery model:

1. Feature work lands through pull requests.
2. CI validates backend tests, frontend tests, frontend build, container builds, and Kubernetes rendering.
3. An operator manually triggers the release workflow from `master`.
4. The release workflow:
   - verifies backend and frontend again
   - builds release artifacts
   - publishes GHCR images
   - renders immutable Kubernetes release manifests
   - creates a GitHub Release

## GitHub Actions sequence

TradeLab currently uses four GitHub Actions workflows:

1. `CI`
2. `Auto Merge PR`
3. `Release`
4. `Promote Production`

### CI workflow

The CI workflow runs on pull requests and on pushes to `master`.

Its jobs run in parallel:

- `Backend tests`
- `Frontend unit tests`
- `Frontend build`
- `Frontend E2E tests`
- `Backend container build`
- `Frontend container build`
- `Kubernetes manifests`
- `Metadata validation`

### Auto-merge workflow

The auto-merge workflow listens for a completed `CI` workflow run.

If the finished run:

- succeeded
- belongs to a pull request
- targets `master`

then the workflow performs a `squash` merge into `master`.

### Release workflow

The release workflow runs only through manual workflow dispatch.

Its execution order is:

1. `Release metadata`
2. in parallel:
   - `Verify backend`
   - `Verify frontend`
3. after verification succeeds, in parallel:
   - `Build backend binaries`
   - `Build frontend artifact`
   - `Publish backend image`
   - `Publish frontend image`
4. `Package Kubernetes manifests`
5. `Create GitHub release`

### Production promotion workflow

The production promotion workflow also runs only through manual workflow dispatch.

It:

1. resolves the selected or latest GitHub release tag
2. updates [tradelab-prod.yaml](../deploy/infrastructure/applications/tradelab-prod.yaml)
3. ensures the production application is part of the Argo CD root kustomization
4. commits the promotion change back to `master`

This means the effective repository delivery path is:

`pull request -> CI -> auto-merge into master -> manual release verification/build/publish/package -> GitHub release -> manual production promotion`

For a release-focused view of published artifacts and release meaning, see [release-process.md](release-process.md).

## Health and failure surfaces

### First places to check

- backend `/healthz` for application readiness
- GitHub Actions logs for CI, release, or packaging failures
- Kubernetes pod status for migration, backend, frontend, and database pods
- ingress routing if the UI is up but API requests fail

### Common failure classes

- `database startup or migration failure`
  Usually visible first in the migration initContainer or backend pod startup.
- `secret/configuration failure`
  Usually visible as failed pod startup, database auth errors, or unreachable upstreams.
- `market-data upstream degradation`
  Backend may fall back to stale market data for a bounded period; after that, market-dependent actions can fail explicitly.
- `strategy engine failure`
  Usually visible first in backend logs as failed strategy evaluations or missing automated trades for active bundles.
- `identity configuration failure`
  Usually visible as missing registered-account bootstrap, rejected Clerk bearer tokens, or an absent social-login surface when auth was expected.
- `session security failure`
  Usually visible as missing or cleared registered app-session cookies, repeated logout loops, or guest-session refresh after an unauthorized response.
- `release packaging failure`
  Usually visible in GitHub Actions during image publication or manifest rendering.

## Rollback expectations

- application rollback should be driven by reapplying a previously released immutable manifest or deploying an older release artifact
- database rollback is not automatic and must be evaluated separately from application rollback
- if a release introduces runtime regressions without schema incompatibility, the preferred rollback path is a prior release manifest and image set

## Operator checklist

Before deploy:

- secret sources are available
- release artifacts or render inputs are correct
- ingress host values are correct for the target environment

After deploy:

- backend `/healthz` responds
- frontend loads and can reach `/api`
- database-backed routes respond successfully
- guest session creation works
- registered-account bootstrap works when Clerk is configured
- protected API access works for the intended identity mode
- active strategies evaluate on the expected tick and produce activity when thresholds are met

When triaging incidents:

- identify whether the failure is build/release, deployment/runtime, or upstream market-data related
- verify the most recent successful release and manifest tag
- confirm whether stale market-data fallback is masking or delaying a harder upstream failure

## GitHub-facing operator notes

The public repository surface is part of the operating model for TradeLab. Keep these aligned with the current product state:

- README positioning and roadmap links
- support, contributing, and security guidance
- release-process documentation
- GitHub repository settings captured in [github-rollout.md](github-rollout.md)
- infrastructure bootstrap guidance in [infrastructure-bootstrap.md](infrastructure-bootstrap.md)
