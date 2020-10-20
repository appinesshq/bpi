SHELL := /bin/bash

# ==============================================================================
# Running from within docker compose

run: up

up:
	docker-compose -f compose.yaml up --detach --remove-orphans

down:
	docker-compose -f compose.yaml down --remove-orphans

logs:
	docker-compose -f compose.yaml logs -f

# ==============================================================================
# Administration

schema:
	go run app/admin/main.go schema

seed:
	go run app/admin/main.go seed

keys:
	go run app/admin/main.go keygen

token:
	go run app/admin/main.go gentoken $(ARGS)

# Modules support

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

tidy:
	go mod tidy
	go mod vendor

deps-upgrade:
	go get -u -t -d -v ./...
	go mod tidy
	go mod vendor

deps-cleancache:
	go clean -modcache


# ==============================================================================
# Running tests within the local computer

test:
	go test ./... -count=1 -v
	staticcheck ./...