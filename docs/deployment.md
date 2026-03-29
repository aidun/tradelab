# Kubernetes Deployment

TradeLab ships with a Kubernetes deployment layout under `deploy/kubernetes` and publishes ready-to-run container images to GitHub Container Registry on every successful `master` release.

For the broader runtime model and operator workflow, see [system-operations.md](system-operations.md). For contributor workflow and local setup, see [developer-guide.md](developer-guide.md).

## What gets deployed

- `PostgreSQL` as a `StatefulSet` with a persistent volume claim
- `TradeLab backend` as a `Deployment`
- `TradeLab frontend` as a `Deployment`
- `Ingress` routing `/` to the frontend and `/api` to the backend
- automatic schema migration through a backend `initContainer`

## Container images

Every successful `master` release publishes:

- `ghcr.io/aidun/tradelab-backend:latest`
- `ghcr.io/aidun/tradelab-backend:v0.1.<run-number>`
- `ghcr.io/aidun/tradelab-frontend:latest`
- `ghcr.io/aidun/tradelab-frontend:v0.1.<run-number>`

The release artifact packages Kubernetes manifests with immutable `v0.1.<run-number>` image tags. The
development overlay keeps using `latest` for convenience, while release deployments should use the packaged
manifest artifact or render production manifests through the release helper script.

## Layout

- `deploy/kubernetes/base`: shared manifests
- `deploy/kubernetes/overlays/development`: single-replica development environment
- `deploy/kubernetes/overlays/production`: production defaults with two app replicas and larger database storage

## Prerequisites

- a Kubernetes cluster
- an NGINX-compatible ingress controller
- `kubectl` with `kustomize` support
- access to the published `ghcr.io/aidun/*` images

## Parameter reference

This section explains which parameters exist, when they need to be set, and where they are consumed.

### Local backend parameters

These environment variables are read by the Go backend in [config.go](../backend/internal/config/config.go).

| Parameter | Required | When to set it | Default | Used by |
| --- | --- | --- | --- | --- |
| `HTTP_ADDRESS` | optional | before starting the backend locally or in a custom runtime | `:8080` | backend HTTP listener |
| `DATABASE_URL` | optional for local default, required for non-default DBs | before starting the backend or running migrations | `postgres://tradelab:tradelab@localhost:5432/tradelab?sslmode=disable` | backend database connection and migration runner |
| `MARKET_DATA_BASE_URL` | optional | before starting the backend if you want to use a different upstream market-data provider endpoint | `https://api.binance.com` | backend market-data service |

### Local frontend parameters

These environment variables affect the Next.js frontend runtime and rewrites.

| Parameter | Required | When to set it | Default | Used by |
| --- | --- | --- | --- | --- |
| `TRADESLAB_API_PROXY_TARGET` | optional | before `npm run dev` or `npm run start` if the backend is not available at the default local address | `http://localhost:8080` | frontend rewrite target for `/api/*` requests |
| `NEXT_PUBLIC_API_BASE_URL` | optional | before frontend startup only if you intentionally want browser calls to go directly to another API origin instead of using the rewrite path | empty string | frontend fetch client in the browser |

Recommended local default:

- leave `NEXT_PUBLIC_API_BASE_URL` unset
- leave `TRADESLAB_API_PROXY_TARGET` unset if the backend runs on `http://localhost:8080`
- only set `DATABASE_URL` if your local PostgreSQL credentials or host differ from the default

### Kubernetes development parameters

The development overlay reads its database values from:

- [`.env.database.example`](../deploy/kubernetes/overlays/development/.env.database.example)
- generated local file: `deploy/kubernetes/overlays/development/.env.database`

You must create `deploy/kubernetes/overlays/development/.env.database` before the first apply.

Required keys in that file:

| Parameter | Required | When to set it | Used by |
| --- | --- | --- | --- |
| `POSTGRES_DB` | yes | before `kubectl apply -k deploy/kubernetes/overlays/development` | PostgreSQL StatefulSet |
| `POSTGRES_USER` | yes | before `kubectl apply -k deploy/kubernetes/overlays/development` | PostgreSQL StatefulSet |
| `POSTGRES_PASSWORD` | yes | before `kubectl apply -k deploy/kubernetes/overlays/development` | PostgreSQL StatefulSet |
| `DATABASE_URL` | yes | before `kubectl apply -k deploy/kubernetes/overlays/development` | backend deployment and migration init container |

Additional development values defined in manifests:

| Parameter | Where it is set | Default value | Notes |
| --- | --- | --- | --- |
| `HTTP_ADDRESS` | [backend-configmap.yaml](../deploy/kubernetes/base/backend-configmap.yaml) | `:8080` | normally no change needed |
| `PORT` | [frontend-configmap.yaml](../deploy/kubernetes/base/frontend-configmap.yaml) | `3000` | internal frontend container port |
| `HOSTNAME` | [frontend-configmap.yaml](../deploy/kubernetes/base/frontend-configmap.yaml) | `0.0.0.0` | internal frontend bind host |
| `TRADESLAB_API_PROXY_TARGET` | [frontend-configmap.yaml](../deploy/kubernetes/base/frontend-configmap.yaml) | `http://tradelab-backend:8080` | keeps frontend `/api` rewrite inside the cluster |
| development host | [patch-ingress.yaml](../deploy/kubernetes/overlays/development/patch-ingress.yaml) | `tradelab.localtest.me` | update if you use another local host name |

### Kubernetes production parameters

The production overlay expects secrets from [external-secret.yaml](../deploy/kubernetes/overlays/production/external-secret.yaml).

Required remote secret properties:

| Parameter | Required | When to set it | Used by |
| --- | --- | --- | --- |
| `POSTGRES_DB` | yes | before applying the production overlay | PostgreSQL StatefulSet |
| `POSTGRES_USER` | yes | before applying the production overlay | PostgreSQL StatefulSet |
| `POSTGRES_PASSWORD` | yes | before applying the production overlay | PostgreSQL StatefulSet |
| `DATABASE_URL` | yes | before applying the production overlay | backend deployment and migration init container |

Production values that usually need review before deployment:

| Setting | Where to change it | When to change it |
| --- | --- | --- |
| external secret store name or path | [external-secret.yaml](../deploy/kubernetes/overlays/production/external-secret.yaml) | before first production deploy if your secret backend naming differs |
| ingress host and TLS host | [patch-ingress.yaml](../deploy/kubernetes/overlays/production/patch-ingress.yaml) | before first production deploy and whenever the public domain changes |
| app replica counts | [patch-backend-deployment.yaml](../deploy/kubernetes/overlays/production/patch-backend-deployment.yaml) and [patch-frontend-deployment.yaml](../deploy/kubernetes/overlays/production/patch-frontend-deployment.yaml) | before scaling production capacity |
| PostgreSQL storage size | [patch-postgres-persistent-volume-claim.yaml](../deploy/kubernetes/overlays/production/patch-postgres-persistent-volume-claim.yaml) | before first production deploy or storage expansion |

## Development deployment

The development overlay uses:

- namespace: `tradelab-dev`
- host: `tradelab.localtest.me`
- a local `.env.database` file that is generated outside Git

Create the development secret input once:

```bash
cp deploy/kubernetes/overlays/development/.env.database.example deploy/kubernetes/overlays/development/.env.database
```

Edit `deploy/kubernetes/overlays/development/.env.database` before the first deployment.

Typical development values:

```env
POSTGRES_DB=tradelab
POSTGRES_USER=tradelab
POSTGRES_PASSWORD=tradelab
DATABASE_URL=postgres://tradelab:tradelab@tradelab-postgres:5432/tradelab?sslmode=disable
```

Then apply it:

```bash
kubectl apply -k deploy/kubernetes/overlays/development
```

If you are testing locally with an ingress controller, map `tradelab.localtest.me` to your cluster ingress IP or use a wildcard resolver such as `localtest.me`.

## Production deployment

The production overlay expects an `External Secrets Operator` installation and a `ClusterSecretStore`
named `tradelab-secrets`. The committed manifest maps the Kubernetes secret `tradelab-database` from
the external secret key `tradelab/production/database`.

Before applying the production overlay:

- make sure the external secret store exposes `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, and `DATABASE_URL`
- update `deploy/kubernetes/overlays/production/external-secret.yaml` if your secret paths differ
- replace the ingress host values in `deploy/kubernetes/overlays/production/patch-ingress.yaml`
- review replica counts and storage size patches if the defaults are not suitable for your environment

Then deploy:

```bash
kubectl apply -k deploy/kubernetes/overlays/production
```

For immutable production output tied to a release tag, render the packaged manifest shape with:

```bash
./deploy/kubernetes/render-release-manifests.sh v0.1.123 /tmp/tradelab-kubernetes.yaml
kubectl apply -f /tmp/tradelab-kubernetes.yaml
```

## Operational notes

- The backend waits on database connectivity implicitly through the migration init container.
- Frontend requests can stay same-origin because the ingress sends `/api` traffic to the backend service.
- The frontend also includes a rewrite fallback to `TRADESLAB_API_PROXY_TARGET`, which keeps local standalone runs aligned with the Kubernetes topology.
- If you deploy outside the included ingress setup, re-check both `TRADESLAB_API_PROXY_TARGET` and `NEXT_PUBLIC_API_BASE_URL` so frontend requests still resolve correctly.
