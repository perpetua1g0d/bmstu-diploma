#!/bin/bash

k3d cluster create bmstucluster \
  --api-port 6443 \
  --servers-memory 4G \
  --agents-memory 4G \
  --k3s-arg "--kubelet-arg=eviction-hard=memory.available<500Mi@server:*" \
  --k3s-arg "--kubelet-arg=eviction-hard=memory.available<500Mi@agent:*" \
  --k3s-arg "--kubelet-arg=image-gc-high-threshold=90@server:*" \
  --k3s-arg "--kubelet-arg=image-gc-low-threshold=80@server:*" \
  --k3s-arg "--kubelet-arg=fail-swap-on=false@server:*" \
  --kubeconfig-update-default \
  --k3s-arg "--kube-apiserver-arg=service-account-jwks-uri=https://kubernetes.default.svc/openid/v1/jwks@server:*" \
  --k3s-arg "--kube-apiserver-arg=service-account-issuer=https://kubernetes.default.svc@server:*"

## The --keep-tools argument with k3d prevents the deletion of the Docker container used to run K3s
# (the lightweight Kubernetes distribution) when the cluster is removed.
# This is useful for keeping the Docker tools and images associated with the cluster for later use.

# services
docker build -t ghcr.io/perpetua1g0d/bmstu-diploma/business-service:latest ./business-service
docker push ghcr.io/perpetua1g0d/bmstu-diploma/business-service:latest
k3d image import ghcr.io/perpetua1g0d/bmstu-diploma/business-service:latest -c bmstucluster --keep-tools

# idp
docker build -t ghcr.io/perpetua1g0d/bmstu-diploma/idp:latest ./idp
docker push ghcr.io/perpetua1g0d/bmstu-diploma/idp:latest
k3d image import ghcr.io/perpetua1g0d/bmstu-diploma/idp:latest -c bmstucluster --keep-tools

# main infra servies with sidecars:
docker build -t ghcr.io/perpetua1g0d/bmstu-diploma/postgres-sidecar:latest ./postgres-sidecar
docker push ghcr.io/perpetua1g0d/bmstu-diploma/postgres-sidecar:latest
k3d image import ghcr.io/perpetua1g0d/bmstu-diploma/postgres-sidecar:latest -c bmstucluster --keep-tools

docker build -t ghcr.io/perpetua1g0d/bmstu-diploma/auth-ui:latest ./auth-ui
docker push ghcr.io/perpetua1g0d/bmstu-diploma/auth-ui:latest
k3d image import ghcr.io/perpetua1g0d/bmstu-diploma/auth-ui:latest -c bmstucluster --keep-tools

kubectl apply -f k8s/namespaces/

kubectl apply -f k8s/monitoring/grafana-dashboards.yaml

namespaces=("postgres-a" "postgres-b" "idp" "admin-panel" "monitoring" "service-a" "service-b")
for ns in "${namespaces[@]}"; do
  if ! kubectl get secret ghcr-secret -n "$ns" >/dev/null 2>&1; then
    kubectl create secret docker-registry ghcr-secret \
      --docker-server=ghcr.io \
      --docker-username=perpetua1g0d \
      --docker-password="$GH_PAT" \
      --namespace="$ns"
    echo "Secret GHCR created in namespace: $ns"
  else
    echo "Secret already exists in namespace: $ns"
  fi
done

kubectl create configmap grafana-dashboards \
  -n monitoring \
  --from-file=cluster_service_metrics.json=k8s/monitoring/dashboards/cluster_service_metrics.json \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl label configmap -n monitoring grafana-dashboards grafana_dashboard=1 --overwrite

# predownload
IMAGES=(
  "quay.io/prometheus-operator/prometheus-operator:v0.68.0"
  "quay.io/prometheus-operator/prometheus-config-reloader:v0.68.0"
  "quay.io/prometheus/prometheus:v2.42.0"
  "docker.io/grafana/grafana:9.5.3"
  "quay.io/kiwigrid/k8s-sidecar:1.30.0"
)

for image in "${IMAGES[@]}"; do
  docker pull $image
  k3d image import $image -c bmstucluster
done

kubectl apply -f k8s/monitoring/rbac.yaml

helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm upgrade --install monitoring prometheus-community/kube-prometheus-stack \
  -n monitoring \
  -f k8s/monitoring/prom-stack-values.yaml \
  --wait \
  --timeout 10m

kubectl apply -f k8s/idp/
kubectl apply -f k8s/postgresql/postgres-a/
kubectl apply -f k8s/postgresql/postgres-b/
kubectl apply -f k8s/service-a/
kubectl apply -f k8s/service-b/
kubectl apply -f k8s/admin-panel/
# kubectl apply -f k8s/kafka/
# kubectl apply -f k8s/redis/

# kubectl annotate pods -n postgres-a -l app=postgres-a \
#   prometheus.io/scrape=true \
#   prometheus.io/port=9187

echo "- Админ-панель: kubectl port-forward -n admin-panel svc/admin-panel 8080:80"
echo "- PostgreSQL: kubectl port-forward -n postgresql svc/postgresql 5434:5434"
echo "- Sidecar: kubectl port-forward -n postgresql svc/postgresql 8080:8080"
echo "Grafana: http://localhost:30091"
echo "Prometheus: http://localhost:30090"
