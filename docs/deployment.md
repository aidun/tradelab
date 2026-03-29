# Kubernetes Deployment

TradeLab ships with a Kubernetes deployment layout under `deploy/kubernetes` and publishes ready-to-run container images to GitHub Container Registry on every successful `master` release.

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

The Kubernetes manifests default to the `latest` tags so a fresh release can be deployed without patching image references first.

## Layout

- `deploy/kubernetes/base`: shared manifests
- `deploy/kubernetes/overlays/development`: single-replica development environment
- `deploy/kubernetes/overlays/production`: production defaults with two app replicas and larger database storage

## Prerequisites

- a Kubernetes cluster
- an NGINX-compatible ingress controller
- `kubectl` with `kustomize` support
- access to the published `ghcr.io/aidun/*` images

## Development deployment

The development overlay uses:

- namespace: `tradelab-dev`
- host: `tradelab.localtest.me`
- a local `.env.database` file that is generated outside Git

Create the development secret input once:

```bash
cp deploy/kubernetes/overlays/development/.env.database.example deploy/kubernetes/overlays/development/.env.database
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

Then deploy:

```bash
kubectl apply -k deploy/kubernetes/overlays/production
```

## Operational notes

- The backend waits on database connectivity implicitly through the migration init container.
- Frontend requests can stay same-origin because the ingress sends `/api` traffic to the backend service.
- The frontend also includes a rewrite fallback to `TRADESLAB_API_PROXY_TARGET`, which keeps local standalone runs aligned with the Kubernetes topology.
