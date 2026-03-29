# Infrastructure Bootstrap

## Purpose

This document defines the baseline Kubernetes platform for TradeLab:

- `Argo CD` as the GitOps controller
- `MetalLB` for LoadBalancer IP allocation on the LAN
- `Traefik` as the managed ingress controller

The goal is to bootstrap a fresh `k3s` cluster into a state where TradeLab can be deployed through Argo CD without additional ad-hoc platform setup.

## Target topology

TradeLab infrastructure now assumes this cluster shape:

- `argocd` namespace for the GitOps control plane
- `metallb-system` namespace for MetalLB
- `traefik-system` namespace for Traefik
- `tradelab-dev` namespace prepared for the development application deployment
- `tradelab` namespace prepared for the production application deployment

The reserved MetalLB LAN range is:

- `192.168.2.200-192.168.2.220`

Traefik is expected to claim:

- `192.168.2.200`

That IP is used by the current development ingress hostname:

- `tradelab.192.168.2.200.sslip.io`

## Repository layout

The infrastructure bootstrap lives under `deploy/infrastructure/`:

- `bootstrap/argocd-install`: manual first-step Argo CD installation
- `bootstrap/root-application`: manual second-step root application
- `applications/`: Argo CD child applications managed by the root app
- `platform/namespaces`: shared namespace creation
- `platform/metallb/config`: MetalLB IP pool and L2 advertisement

TradeLab application manifests remain under `deploy/kubernetes/` and are not replaced by this structure.

## Bootstrap order

### Step 1. Install Argo CD

```bash
kubectl apply -k deploy/infrastructure/bootstrap/argocd-install
kubectl -n argocd rollout status deployment/argocd-server
```

### Step 2. Apply the root application

```bash
kubectl apply -k deploy/infrastructure/bootstrap/root-application
```

The root application then manages:

1. platform namespaces
2. MetalLB
3. Traefik

## Argo CD model

TradeLab uses an `App of Apps` layout.

- root app: `tradelab-root`
- child apps:
  - `platform-namespaces`
  - `metallb`
  - `traefik`
- prepared but not yet wired into the root app:
  - `tradelab-dev`

Sync behavior for platform applications:

- `automated`
- `prune`
- `selfHeal`

Sync waves are used so the cluster is built in a predictable order:

- namespaces first
- MetalLB second
- Traefik third

## Operational expectations

- MetalLB must advertise only addresses that are actually free on the LAN.
- Traefik is the only ingress class assumed by the committed TradeLab overlays.
- The repo does not yet bootstrap `cert-manager`, `External Secrets Operator`, or TLS automation in this infrastructure layer.
- Production secret handling remains a later step and still belongs to the application deployment flow.

## Validation checklist

After bootstrap, confirm:

- `kubectl get ns` shows `argocd`, `metallb-system`, `traefik-system`, `tradelab-dev`, and `tradelab`
- `kubectl -n metallb-system get pods` shows a healthy controller and speakers
- `kubectl -n traefik-system get svc` shows Traefik with `192.168.2.200`
- `kubectl get ingressclass` includes `traefik`
- `kubectl -n argocd get applications` shows the root app and child apps as `Synced` and `Healthy`

## Next step

Once this platform layer is healthy, the next TradeLab step is to create an Argo CD application for `deploy/kubernetes/overlays/development` and let the application release ride on top of the managed ingress and load-balancer foundation.
