# Kubernetes Deployment

TradeLab ships with a Kubernetes deployment layout under `deploy/kubernetes`, a GitOps infrastructure bootstrap under `deploy/infrastructure`, and publishes ready-to-run container images to GitHub Container Registry for both `master` development builds and official releases.

For the broader runtime model and operator workflow, see [system-operations.md](system-operations.md). For contributor workflow and local setup, see [developer-guide.md](developer-guide.md). For the GitOps platform bootstrap, see [infrastructure-bootstrap.md](infrastructure-bootstrap.md). For a first successful run by audience, start with [getting-started.md](getting-started.md).

## What gets deployed

- `PostgreSQL` as a `StatefulSet` with a persistent volume claim
- `TradeLab backend` as a `Deployment`
- `TradeLab frontend` as a `Deployment`
- `Ingress` routing `/` to the frontend and `/api` to the backend
- automatic schema migration through a backend `initContainer`

## Container images

Every successful `master` push publishes development images:

- `ghcr.io/aidun/tradelab-backend:master`
- `ghcr.io/aidun/tradelab-backend:master-<shortsha>`
- `ghcr.io/aidun/tradelab-frontend:master`
- `ghcr.io/aidun/tradelab-frontend:master-<shortsha>`

Every successful official release publishes:

- `ghcr.io/aidun/tradelab-backend:latest`
- `ghcr.io/aidun/tradelab-backend:v0.1.<run-number>`
- `ghcr.io/aidun/tradelab-frontend:latest`
- `ghcr.io/aidun/tradelab-frontend:v0.1.<run-number>`

Argo CD development now deploys immutable `master-<shortsha>` image tags that are committed back into the development application manifest after each successful `master` image publish. Release deployments should use the packaged manifest artifact or the production Argo CD application pinned to an official release tag.

## Layout

- `deploy/infrastructure`: Argo CD bootstrap plus platform applications for MetalLB and Traefik
- `deploy/kubernetes/base`: shared manifests
- `deploy/kubernetes/overlays/development`: single-replica development environment
- `deploy/kubernetes/overlays/production`: production defaults with two app replicas and larger database storage

## Prerequisites

- a Kubernetes cluster
- the infrastructure bootstrap from [infrastructure-bootstrap.md](infrastructure-bootstrap.md), or an equivalent `Traefik + MetalLB + Argo CD` foundation
- `kubectl` with `kustomize` support
- access to the published `ghcr.io/aidun/*` images

## Audience-specific use

Use this document differently depending on your role:

- developers should use it as the parameter reference after reading [getting-started.md](getting-started.md)
- operators should use it as the deployment execution guide together with [system-operations.md](system-operations.md)
- first-time product reviewers usually do not need this document and should start with [user-guide.md](user-guide.md)

## Parameter reference

This section explains which parameters exist, when they need to be set, and where they are consumed.

### Local backend parameters

These environment variables are read by the Go backend in [config.go](../backend/internal/config/config.go).

| Parameter | Required | When to set it | Default | Used by |
| --- | --- | --- | --- | --- |
| `HTTP_ADDRESS` | optional | before starting the backend locally or in a custom runtime | `:8080` | backend HTTP listener |
| `DATABASE_URL` | optional for local default, required for non-default DBs | before starting the backend or running migrations | `postgres://tradelab:tradelab@localhost:5432/tradelab?sslmode=disable` | backend database connection and migration runner |
| `MARKET_DATA_BASE_URL` | optional | before starting the backend if you want to use a different upstream market-data provider endpoint | `https://api.binance.com` | backend market-data service |
| `TRADESLAB_CLERK_ISSUER_URL` | optional | before starting the backend when real Clerk-backed registered accounts should be enabled | empty | backend Clerk JWT issuer validation |
| `TRADESLAB_CLERK_JWKS_URL` | optional | before starting the backend when real Clerk-backed registered accounts should be enabled | empty | backend Clerk JWKS lookup |
| `TRADESLAB_AUTH_MOCK_MODE` | optional | before starting the backend in local or CI runs when mock registered-account tokens should be accepted | `false` | backend auth verifier |

### Local frontend parameters

These environment variables affect the Next.js frontend runtime and rewrites.

| Parameter | Required | When to set it | Default | Used by |
| --- | --- | --- | --- | --- |
| `TRADESLAB_API_PROXY_TARGET` | optional | before `npm run dev` or `npm run start` if the backend is not available at the default local address | `http://localhost:8080` | frontend rewrite target for `/api/*` requests |
| `NEXT_PUBLIC_API_BASE_URL` | optional | before frontend startup only if you intentionally want browser calls to go directly to another API origin instead of using the rewrite path | empty string | frontend fetch client in the browser |
| `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` | optional | before frontend startup when you want the real Clerk UI and token flow in the browser | empty | frontend Clerk provider |
| `NEXT_PUBLIC_AUTH_MOCK_MODE` | optional | before frontend startup in local or CI runs when mock Google/Apple auth should replace live Clerk configuration | `false` | frontend auth provider wrapper |

Security notes:

- `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` is frontend-safe
- Clerk secret material must stay server-side and must not be committed or exposed through browser env vars
- `TRADESLAB_AUTH_MOCK_MODE` and `NEXT_PUBLIC_AUTH_MOCK_MODE` are development and CI switches only and must not be enabled in production manifests

Recommended local default:

- leave `NEXT_PUBLIC_API_BASE_URL` unset
- leave `TRADESLAB_API_PROXY_TARGET` unset if the backend runs on `http://localhost:8080`
- only set `DATABASE_URL` if your local PostgreSQL credentials or host differ from the default

First local run order:

1. start PostgreSQL
2. apply migrations
3. start the backend
4. start the frontend

### Kubernetes development parameters

The development overlay no longer requires a repo-local `.env.database` file.

If the secret `tradelab-database` does not already exist, the committed bootstrap job creates it automatically with:

- `POSTGRES_DB=tradelab`
- `POSTGRES_USER=tradelab`
- a generated `POSTGRES_PASSWORD`
- a generated `DATABASE_URL`

If you want fixed development credentials instead, create `tradelab-database` manually before the first sync or apply. The bootstrap job exits without changing anything when the secret already exists.

Additional development values defined in manifests:

| Parameter | Where it is set | Default value | Notes |
| --- | --- | --- | --- |
| `HTTP_ADDRESS` | [backend-configmap.yaml](../deploy/kubernetes/base/backend-configmap.yaml) | `:8080` | normally no change needed |
| `PORT` | [frontend-configmap.yaml](../deploy/kubernetes/base/frontend-configmap.yaml) | `3000` | internal frontend container port |
| `HOSTNAME` | [frontend-configmap.yaml](../deploy/kubernetes/base/frontend-configmap.yaml) | `0.0.0.0` | internal frontend bind host |
| `TRADESLAB_API_PROXY_TARGET` | [frontend-configmap.yaml](../deploy/kubernetes/base/frontend-configmap.yaml) | `http://tradelab-backend:8080` | keeps frontend `/api` rewrite inside the cluster |
| development entrypoints | [patch-ingress.yaml](../deploy/kubernetes/overlays/development/patch-ingress.yaml) | `http://192.168.2.200/` and `http://192.168.2.200/tradelab-dev` | hostless Traefik access on the MetalLB IP |

Development image policy:

- the overlay defaults to the mutable `master` tag for direct manual rendering
- the Argo CD development application overrides both images to immutable `master-<shortsha>` tags
- the same Argo application pins `targetRevision` to the exact source commit that produced those images
- if the master-image workflow was introduced after the current `master` state landed or a run was missed, operators can manually dispatch `Publish Master Images` with `source_ref=master` to backfill the immutable development images and Argo target update

### Kubernetes production parameters

The default production overlay is now self-contained and follows the same bootstrap rule as development:

- if `tradelab-database` does not exist yet, the bootstrap job creates it with generated credentials
- if `tradelab-database` already exists, the bootstrap job does nothing

This makes first bring-up possible without a separate secret controller.

For a later operator-managed secret flow, TradeLab also ships:

- [production-external-secrets](../deploy/kubernetes/overlays/production-external-secrets/kustomization.yaml)

That overlay adds an `ExternalSecret` on top of the normal production deployment and explicitly removes the bootstrap job so the external secret store stays authoritative.

Required remote secret properties for the external-secret variant:

| Parameter | Required | When to set it | Used by |
| --- | --- | --- | --- |
| `POSTGRES_DB` | yes | before applying the production overlay | PostgreSQL StatefulSet |
| `POSTGRES_USER` | yes | before applying the production overlay | PostgreSQL StatefulSet |
| `POSTGRES_PASSWORD` | yes | before applying the production overlay | PostgreSQL StatefulSet |
| `DATABASE_URL` | yes | before applying the production overlay | backend deployment and migration init container |

Production values that usually need review before deployment:

| Setting | Where to change it | When to change it |
| --- | --- | --- |
| external secret store name or path | [external-secret.yaml](../deploy/kubernetes/overlays/production-external-secrets/external-secret.yaml) | before first production deploy if your secret backend naming differs |
| ingress host and TLS host | [patch-ingress.yaml](../deploy/kubernetes/overlays/production/patch-ingress.yaml) | before first production deploy and whenever the public domain changes |
| app replica counts | [patch-backend-deployment.yaml](../deploy/kubernetes/overlays/production/patch-backend-deployment.yaml) and [patch-frontend-deployment.yaml](../deploy/kubernetes/overlays/production/patch-frontend-deployment.yaml) | before scaling production capacity |
| PostgreSQL storage size | [patch-postgres-persistent-volume-claim.yaml](../deploy/kubernetes/overlays/production/patch-postgres-persistent-volume-claim.yaml) | before first production deploy or storage expansion |

## Development deployment

The development overlay uses:

- namespace: `tradelab-dev`
- primary entrypoint: `http://192.168.2.200/`
- alternate entrypoint: `http://192.168.2.200/tradelab-dev`
- generated database credentials if no secret already exists

Then apply it:

```bash
kubectl apply -k deploy/kubernetes/overlays/development
```

After apply, validate:

- the `tradelab-database` secret exists
- PostgreSQL pod is healthy
- migration init container completed
- backend `/healthz` responds
- frontend loads through ingress
- a guest demo session can be created

If you are using the committed platform bootstrap, Traefik is reachable directly on the reserved MetalLB IP. The development ingress is intentionally hostless so the app does not depend on local DNS. The `/tradelab-dev` entrypoint is implemented through a Traefik strip-prefix middleware and lands on the same canonical application as `/`.

## Production deployment

Before applying the production overlay:

- review the hostless path entrypoints in `deploy/kubernetes/overlays/production/patch-ingress.yaml`
- review replica counts and storage size patches if the defaults are not suitable for your environment

The production overlay is prepared for these IP-based entrypoints:

- primary entrypoint: `http://<traefik-ip>/`
- alternate entrypoint: `http://<traefik-ip>/tradelab`

As with development, the `/tradelab` production entrypoint is implemented through a Traefik strip-prefix middleware and lands on the same canonical application as `/`.

If you want generated first-run credentials, apply the default production overlay:

```bash
kubectl apply -k deploy/kubernetes/overlays/production
```

If you want operator-managed credentials from External Secrets, make sure the external secret store exposes `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, and `DATABASE_URL`, then apply:

```bash
kubectl apply -k deploy/kubernetes/overlays/production-external-secrets
```

This variant is the recommended long-term production shape because:

- generated first-run credentials are useful for bootstrap and review environments
- the external-secret variant keeps secret ownership in the secret store
- Argo CD can manage both overlays without the bootstrap job racing the external secret controller

For immutable production output tied to a release tag, render the packaged manifest shape with:

```bash
./deploy/kubernetes/render-release-manifests.sh v0.1.123 /tmp/tradelab-kubernetes.yaml
kubectl apply -f /tmp/tradelab-kubernetes.yaml
```

After apply, validate the first protected product flow with [installation-validation.md](installation-validation.md).

Production image policy:

- the production Argo CD application always uses release-tagged backend and frontend images
- the `Promote Production` workflow updates both the Argo source revision and the production image tags together
- production should never depend on mutable `master` or `latest` tags

## Operational notes

- The backend waits on database connectivity implicitly through the migration init container.
- Frontend requests can stay same-origin because the ingress sends `/api` traffic to the backend service.
- The frontend also includes a rewrite fallback to `TRADESLAB_API_PROXY_TARGET`, which keeps local standalone runs aligned with the Kubernetes topology.
- If you deploy outside the included ingress setup, re-check both `TRADESLAB_API_PROXY_TARGET` and `NEXT_PUBLIC_API_BASE_URL` so frontend requests still resolve correctly.
- The committed application overlays now assume the `traefik` ingress class.
- The committed overlays now generate initial database credentials automatically when `tradelab-database` is missing.
- The recommended cluster entrypoint is the GitOps platform bootstrap in [infrastructure-bootstrap.md](infrastructure-bootstrap.md).
- First-time install and smoke-test expectations live in [getting-started.md](getting-started.md) and [installation-validation.md](installation-validation.md).
