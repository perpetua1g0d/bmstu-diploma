#!/bin/bash

# Создаем кластер
k3d cluster create bmstucluster \
  --api-port 6443 \
  --servers-memory 4G \
  --agents-memory 4G \
  --k3s-arg "--kubelet-arg=eviction-hard=memory.available<500Mi@server:*" \
  --k3s-arg "--kubelet-arg=eviction-hard=memory.available<500Mi@agent:*" \
  --k3s-arg "--kubelet-arg=image-gc-high-threshold=90@server:*" \
  --k3s-arg "--kubelet-arg=image-gc-low-threshold=80@server:*" \
  --k3s-arg "--kubelet-arg=fail-swap-on=false@server:*"
  --kubeconfig-update-default

# Собираем и загружаем образ
docker build -t ghcr.io/perpetua1g0d/bmstu-diploma/sidecar:latest ./sidecar
k3d image import ghcr.io/perpetua1g0d/bmstu-diploma/sidecar:latest -c bmstucluster --keep-tools

# Применяем манифесты
kubectl apply -f k8s/00-namespaces/
kubectl apply -f k8s/01-talos/
kubectl apply -f k8s/02-postgresql/
# kubectl apply -f k8s/03-kafka/
# kubectl apply -f k8s/04-redis/
# kubectl apply -f k8s/05-sidecars/kafka-sidecar.yaml
kubectl apply -f k8s/05-sidecars/postgresql-sidecar.yaml
kubectl apply -f k8s/06-admin-panel/

# Проверка
# echo "Состояние подов:"
# kubectl get pods -A

echo "Система развернута. Доступные сервисы:"
echo "- Админ-панель: kubectl port-forward -n admin-panel svc/admin-panel 8080:80"
echo "- PostgreSQL: kubectl port-forward -n postgresql svc/postgresql 5434:5432"
echo "- Kafka: kubectl port-forward -n kafka svc/kafka 9092:9092"
