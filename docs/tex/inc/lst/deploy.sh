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
  --k3s-arg "--kube-apiserver-arg=service-account-jwks-uri= \
        https://kubernetes.default.svc/openid/v1/jwks@server:*" \
  --k3s-arg "--kube-apiserver-arg=service-account-issuer= \
        https://kubernetes.default.svc@server:*"

# talos
docker build -t ghcr.io/perpetua1g0d/bmstu-diploma/talos:latest ./talos
docker push ghcr.io/perpetua1g0d/bmstu-diploma/talos:latest
k3d image import ghcr.io/perpetua1g0d/bmstu-diploma/talos:latest -c bmstucluster --keep-tools

# run sidecar code in sidecar containter:
docker build -t ghcr.io/perpetua1g0d/bmstu-diploma/postgres-sidecar:latest ./postgres-sidecar
docker push ghcr.io/perpetua1g0d/bmstu-diploma/postgres-sidecar:latest
k3d image import ghcr.io/perpetua1g0d/bmstu-diploma/postgres-sidecar:latest -c bmstucluster --keep-tools

kubectl apply -f k8s/namespaces/
# kubectl apply -k k8s/namespaces/

namespaces=("postgres-a" "postgres-b" "talos")
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

kubectl apply -f k8s/talos/
kubectl apply -f k8s/postgresql/postgres-a/
kubectl apply -f k8s/postgresql/postgres-b/
