#!/bin/bash


kubectl rollout restart deployment -n postgres-a
kubectl rollout restart deployment -n postgres-b
kubectl rollout restart deployment -n idp

# after changes values.yaml
helm upgrade monitoring -n monitoring -f k8s/monitoring/prom-stack-values.yaml

# removes monitoring release
helm uninstall monitoring -n monitoring
# removes linked resources:
kubectl delete pvc -n monitoring --all
kubectl delete configmap -n monitoring --all
# apply changed config:
helm upgrade --install monitoring prometheus-community/kube-prometheus-stack \
  -n monitoring \
  -f k8s/monitoring/prom-stack-values.yaml \
  --wait \
  --timeout 10m

# fswatch -o ./postgres-sidecar | xargs -n1 -I{} ./update-sidecar.sh
