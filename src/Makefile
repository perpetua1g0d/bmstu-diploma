run:
	./scripts/deploy.sh
teardown:
	bash scripts/teardown.sh

pods:
	kubectl get pods -A

interface:
	kubectl port-forward -n admin-panel svc/auth-ui 8080:80
grafana:
	kubectl port-forward -n monitoring svc/monitoring-grafana 3000:80
prometheus:
	kubectl port-forward -n monitoring svc/prometheus-operated 30090:9090
#	kubectl port-forward -n monitoring prometheus-monitoring-kube-prometheus-prometheus-0 30090:9090

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

.PHONY: run teardown pods interface a-check-db b-check-db a-sidecar-logs b-sidecar-logs grafana prometheus
