# Хук для автоматического создания секрета при создании namespace
kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: ghcr-secret-creator
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["create"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ghcr-secret-creator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ghcr-secret-creator
subjects:
- kind: ServiceAccount
  name: default
  namespace: kube-system
EOF
