# BRYNN developer commands.
# Host requirements: make and Docker. Nothing else is installed on your machine.
.DEFAULT_GOAL := help

# If this tag does not exist yet, use the previous release: make check GO_IMAGE=golang:1.26
GO_IMAGE ?= golang:1.27

# Go tooling runs in a pinned container so every machine (and CI) agrees.
DOCKER_GO = docker run --rm -e CGO_ENABLED=1 -e GOCACHE=/cache/build -e GOMODCACHE=/cache/mod \
	-v "$(CURDIR)":/work -v brynn-go-cache:/cache -w /work $(GO_IMAGE)

.PHONY: help check fmt up down logs smoke

help: ## List available commands
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-8s %s\n", $$1, $$2}'

check: ## Format check, go vet and race-enabled tests (in a Go container)
	$(DOCKER_GO) sh scripts/check-go.sh

fmt: ## Rewrite Go files with gofmt (run this if check reports unformatted files)
	docker run --rm -u "$$(id -u):$$(id -g)" -v "$(CURDIR)":/work -w /work/apps/api $(GO_IMAGE) gofmt -w .

up: ## Build and start the API on http://127.0.0.1:8080 (waits until healthy)
	docker compose up --build -d --wait

down: ## Stop and remove the containers
	docker compose down

logs: ## Follow the API logs
	docker compose logs -f api

smoke: ## Probe the running API
	curl -fsS http://127.0.0.1:8080/healthz && echo
	curl -fsS http://127.0.0.1:8080/readyz && echo
	curl -fsS http://127.0.0.1:8080/version && echo
