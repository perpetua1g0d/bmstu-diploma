#!/bin/bash

# Создаем кластер с ограничением ресурсов
k3d cluster create bmstucluster \
  --api-port 6443 \
  --servers-memory 1G \
  --agents-memory 1G \
  --k3s-arg "--kubelet-arg=eviction-hard=memory.available<100Mi@server:*" \
  --k3s-arg "--kubelet-arg=eviction-hard=memory.available<100Mi@agent:*"

# Собираем и загружаем образ сайдкара
docker build -t bmstu-sidecar:latest ./sidecar
k3d image import bmstu-sidecar:latest -c bmstucluster

# Применяем манифесты в правильном порядке
kubectl apply -f k8s/00-namespaces/
kubectl apply -f k8s/01-talos/
kubectl apply -f k8s/02-postgresql/
kubectl apply -f k8s/03-kafka/
kubectl apply -f k8s/04-redis/
kubectl apply -f k8s/05-sidecars/
kubectl apply -f k8s/06-admin-panel/

echo "Система развернута. Доступные сервисы:"
echo "- Админ-панель: kubectl port-forward -n admin-panel svc/admin-panel 8080:80"
echo "- PostgreSQL: kubectl port-forward -n postgresql svc/postgresql 5434:5432"
echo "- Kafka: kubectl port-forward -n kafka svc/kafka 9092:9092"
