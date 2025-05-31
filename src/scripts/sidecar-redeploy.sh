#!/bin/bash

docker build -t k3d-bmsturegistry:5000/bmstu-diploma/postgres-sidecar:latest ./postgres-sidecar
docker push k3d-bmsturegistry:5000/bmstu-diploma/postgres-sidecar:latest

kubectl rollout restart deployment -n postgres-a
kubectl rollout restart deployment -n postgres-b

# fswatch -o ./postgres-sidecar | xargs -n1 -I{} ./update-sidecar.sh
