# Kubernetes 示例清单

**English:** [README.md](./README.md)

与 `deploy/README_zh.md` 中的说明配合使用。默认资源为 **Cluster 级**（未写 `namespace`），适合快速试用。

## 一次性应用

```bash
kubectl apply -k deploy/kubernetes
```

## 指定命名空间（推荐生产）

1. 创建命名空间，例如：

```bash
kubectl create namespace lightningrag
```

2. 应用时加上 `-n`（Kustomize 根目录未内置 namespace，需逐文件或改用 overlay）：

```bash
kubectl apply -n lightningrag -f deploy/kubernetes/server/
kubectl apply -n lightningrag -f deploy/kubernetes/web/
```

或在本地维护 `kustomization.yaml` 增加：

```yaml
namespace: lightningrag
resources:
  - namespace.yaml   # kind: Namespace metadata.name: lightningrag
  - server/...
```

3. 将 **数据库、Redis** 等依赖以 StatefulSet/托管服务或单独 Helm 安装；本仓库 ConfigMap 中 MySQL/Redis 地址需改为集群内 Service DNS。

## TLS

在 `lrag-web-ingress.yaml` 上增加 `spec.tls` 与证书 Secret，或使用 cert-manager 自动签发；同时把 `rules.host` 改为实际域名。

## 与 Compose 的差异

- 前端 ConfigMap 中 `proxy_pass` 指向 **`lightningrag-server:8888`**，须与 `lrag-server-service.yaml` 中 `metadata.name` 一致。  
- 后端探针路径为 **`/health`**；若配置里 `system.router-prefix` 非空，需同步修改 Deployment 中的 `httpGet.path`。
