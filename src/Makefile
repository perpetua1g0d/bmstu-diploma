run:
	./scripts/deploy.sh
teardown:
	bash scripts/teardown.sh

pods:
	kubectl get pods -A
interface:
	kubectl port-forward -n admin-panel svc/auth-ui 8080:80

a-check-db:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Error: Pod name is required. Usage: make $@ <pod-name>"; \
		exit 1; \
	fi
	kubectl exec -it $(word 2, $(MAKECMDGOALS)) -n postgres-a -c postgres -- psql -U admin -d appdb -c "SELECT * FROM log"
	@# Игнорируем аргумент, чтобы make не пытался его обработать как цель
	@: $(wordlist 2, $(words $(MAKECMDGOALS)), $(MAKECMDGOALS))
b-check-db:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Error: Pod name is required. Usage: make $@ <pod-name>"; \
		exit 1; \
	fi
	kubectl exec -it $(word 2, $(MAKECMDGOALS)) -n postgres-b -c postgres -- psql -U admin -d appdb -c "SELECT * FROM log"
	@# Игнорируем аргумент, чтобы make не пытался его обработать как цель
	@: $(wordlist 2, $(words $(MAKECMDGOALS)), $(MAKECMDGOALS))

a-sidecar-logs:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Error: Pod name is required. Usage: make $@ <pod-name>"; \
		exit 1; \
	fi
	kubectl logs -n postgres-a -c sidecar $(word 2, $(MAKECMDGOALS))
	@# Игнорируем аргумент, чтобы make не пытался его обработать как цель
	@: $(wordlist 2, $(words $(MAKECMDGOALS)), $(MAKECMDGOALS))

b-sidecar-logs:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Error: Pod name is required. Usage: make $@ <pod-name>"; \
		exit 1; \
	fi
	kubectl logs -n postgres-b -c sidecar $(word 2, $(MAKECMDGOALS))
	@# Игнорируем аргумент, чтобы make не пытался его обработать как цель
	@: $(wordlist 2, $(words $(MAKECMDGOALS)), $(MAKECMDGOALS))

# a-sidecar-logs ($1=postgres-a-5498bf5dd-6xv7t):
# kubectl logs -n postgres-a -c sidecar $1

# a-check-db ($1=postgres-a-5498bf5dd-6xv7t):
# kubectl exec -it postgres-a-5498bf5dd-6xv7t -n postgres-a -c postgres -- psql -U admin -d appdb -c "SELECT * FROM log"

# todo: learn also how to restart pods with new changes w/o restarting whole cluster

.PHONY: run teardown pods interface a-check-db b-check-db a-sidecar-logs b-sidecar-logs
