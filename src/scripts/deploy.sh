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
  --kubeconfig-update-default

docker build -t ghcr.io/perpetua1g0d/bmstu-diploma/sidecar:latest ./sidecar
k3d image import ghcr.io/perpetua1g0d/bmstu-diploma/sidecar:latest -c bmstucluster --keep-tools

kubectl apply -f k8s/00-namespaces/

if ! kubectl get secret ghcr-secret -n postgresql >/dev/null 2>&1; then
  echo "Creating ghcr.io secret..."
  kubectl create secret docker-registry ghcr-secret \
    --docker-server=ghcr.io \
    --docker-username=perpetua1g0d \
    --docker-password="$GH_PAT" \
    --namespace=postgresql
else
  echo "ghcr secret already exists, skipped creating secret."
fi

kubectl apply -f k8s/talos/
kubectl apply -f k8s/postgresql/
# kubectl apply -f k8s/kafka/
# kubectl apply -f k8s/redis/
kubectl apply -f k8s/admin-panel/

echo "- Админ-панель: kubectl port-forward -n admin-panel svc/admin-panel 8080:80"
echo "- PostgreSQL: kubectl port-forward -n postgresql svc/postgresql 5434:5434"
echo "- Sidecar: kubectl port-forward -n postgresql svc/postgresql 8080:8080"
