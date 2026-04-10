# Kubernetes sample manifests

**简体中文：** [README_zh.md](./README_zh.md)

Use together with [deploy/README.md](../README.md). Resources are **cluster-scoped** by default (no `namespace`), suitable for quick trials.

## Apply everything

```bash
kubectl apply -k deploy/kubernetes
```

## Use a namespace (recommended for production)

1. Create a namespace, e.g.:

```bash
kubectl create namespace lightningrag
```

2. Apply with `-n` (the Kustomize root does not set `namespace`; use per-file `-n` or an overlay):

```bash
kubectl apply -n lightningrag -f deploy/kubernetes/server/
kubectl apply -n lightningrag -f deploy/kubernetes/web/
```

Or extend `kustomization.yaml` locally:

```yaml
namespace: lightningrag
resources:
  - namespace.yaml   # kind: Namespace metadata.name: lightningrag
  - server/...
```

3. Run **MySQL, Redis**, etc. via StatefulSet, managed service, or Helm; update ConfigMap MySQL/Redis addresses to in-cluster Service DNS.

## TLS

Add `spec.tls` and a cert Secret on `lrag-web-ingress.yaml`, or use cert-manager; set `rules.host` to your real domain.

## Differences vs. Compose

- Front ConfigMap `proxy_pass` targets **`lightningrag-server:8888`** — must match `metadata.name` in `lrag-server-service.yaml`.  
- Backend probe path is **`/health`**; if `system.router-prefix` is non-empty, update Deployment `httpGet.path` accordingly.
