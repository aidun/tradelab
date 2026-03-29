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
- database password: `tradelab`

Apply it with:

```bash
kubectl apply -k deploy/kubernetes/overlays/development
```

If you are testing locally with an ingress controller, map `tradelab.localtest.me` to your cluster ingress IP or use a wildcard resolver such as `localtest.me`.

## Production deployment

Before applying the production overlay, replace the placeholder database password and ingress host values:

- `deploy/kubernetes/base/database-secret.yaml`
- `deploy/kubernetes/overlays/production/patch-ingress.yaml`

Then deploy:

```bash
kubectl apply -k deploy/kubernetes/overlays/production
```

## Operational notes

- The backend waits on database connectivity implicitly through the migration init container.
- Frontend requests can stay same-origin because the ingress sends `/api` traffic to the backend service.
- The frontend also includes a rewrite fallback to `TRADESLAB_API_PROXY_TARGET`, which keeps local standalone runs aligned with the Kubernetes topology.
