COMPOSE ?= podman-compose
GOCACHE ?= /tmp/go-build

.PHONY: build
build:
	GOCACHE=$(GOCACHE) go build ./...

.PHONY: test
test:
	GOCACHE=$(GOCACHE) go test ./...

.PHONY: run-api
run-api:
	GOCACHE=$(GOCACHE) go run ./cmd/api

.PHONY: run-worker
run-worker:
	GOCACHE=$(GOCACHE) go run ./cmd/worker

.PHONY: db-up
db-up:
	$(COMPOSE) -f compose.dev.yaml up -d

.PHONY: db-down
db-down:
	$(COMPOSE) -f compose.dev.yaml down

.PHONY: migrate-up
migrate-up:
	./scripts/migrate-up.sh
