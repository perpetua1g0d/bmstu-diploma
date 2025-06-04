namespaces=("postgres-a" "postgres-b" "idp" "admin-panel" "monitoring" "service-a" "service-b")
for ns in "${namespaces[@]}"; do
    # kubectl delete secret ghcr-secret -n $ns
    kubectl create secret docker-registry ghcr-secret \
      --docker-server=ghcr.io \
      --docker-username=perpetua1g0d \
      --docker-password="$GH_PAT" \
      --namespace="$ns"
    echo "Secret GHCR recreated in namespace: $ns"
done
