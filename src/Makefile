run:
	./scripts/deploy.sh
teardown:
	bash scripts/teardown.sh

pods:
	kubectl get pods -A
interface:
	kubectl port-forward -n admin-panel svc/auth-ui 8080:80

# a-sidecar-logs ($1=postgres-a-5498bf5dd-6xv7t):
# kubectl logs -n postgres-a -c sidecar $1

# a-check-db ($1=postgres-a-5498bf5dd-6xv7t):
# kubectl exec -it postgres-a-5498bf5dd-6xv7t -n postgres-a -c postgres -- psql -U admin -d appdb -c "SELECT * FROM log"

# todo: learn also how to restart pods with new changes w/o restarting whole cluster

.PHONY: run teardown pods interface
