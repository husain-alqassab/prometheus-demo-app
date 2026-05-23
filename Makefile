.PHONY: help build run test docker-build docker-run clean deploy

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Go application
	go build -o prometheus-demo-app main.go

run: ## Run the application locally
	go run main.go

test: ## Run tests
	go test -v ./...

docker-build: ## Build Docker image
	docker build -t prometheus-demo-app:latest .

docker-run: ## Run Docker container
	docker run -p 8080:8080 prometheus-demo-app:latest

clean: ## Clean build artifacts
	rm -f prometheus-demo-app
	go clean

deploy: ## Deploy to OpenShift (requires oc login)
	oc new-build --binary --name=prometheus-demo-app -l app=prometheus-demo-app || true
	oc start-build prometheus-demo-app --from-dir=. --follow
	oc apply -f deployment.yaml
	oc apply -f service.yaml
	oc apply -f route.yaml

undeploy: ## Remove from OpenShift
	oc delete route prometheus-demo-app || true
	oc delete service prometheus-demo-app || true
	oc delete deployment prometheus-demo-app || true
	oc delete buildconfig prometheus-demo-app || true
	oc delete imagestream prometheus-demo-app || true

logs: ## Show application logs
	oc logs -f deployment/prometheus-demo-app

status: ## Show deployment status
	oc get pods,svc,route -l app=prometheus-demo-app