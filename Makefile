SHELL := /bin/bash
SERVICE ?=

.PHONY: proto-gen build test lint docker-compose docker-build migrate-db seed clean

# --- Code generation ---
proto-gen:
	cd proto && buf generate
	@echo "Proto code generated."

# --- Build ---
build:
	@echo "Building all services..."
	go build ./...
	@echo "Build complete."

build-service:
	@if [ -z "$(SERVICE)" ]; then echo "SERVICE is required (e.g. make build-service SERVICE=auth)"; exit 1; fi
	go build ./services/$(SERVICE)/...

# --- Test ---
test:
	go test ./... -count=1

test-service:
	@if [ -z "$(SERVICE)" ]; then echo "SERVICE is required (e.g. make test-service SERVICE=auth)"; exit 1; fi
	go test ./services/$(SERVICE)/... -count=1

# --- Lint ---
lint:
	golangci-lint run ./...

lint-service:
	@if [ -z "$(SERVICE)" ]; then echo "SERVICE is required (e.g. make lint-service SERVICE=auth)"; exit 1; fi
	golangci-lint run ./services/$(SERVICE)/...

# --- Local infrastructure ---
infra-up:
	docker-compose -f deploy/docker/docker-compose.yml up -d

infra-down:
	docker-compose -f deploy/docker/docker-compose.yml down

docker-build:
	@echo "Building gateway..."
	docker build -f deploy/docker/Dockerfile.gateway -t doubt-resolver/gateway .
	@for s in auth user question matching solution chat notification knowledgebase analytics admin; do \
		echo "Building $$s service..."; \
		docker build -f deploy/docker/Dockerfile.service --build-arg SERVICE=$$s -t doubt-resolver/$$s .; \
	done

# --- Database ---
migrate-db:
	@echo "Run migrations per service (requires golang-migrate)."
	@echo "Example:"
	@echo "  migrate -path services/auth/migrations -database 'postgres://postgres:postgres@localhost:5432/doubt_resolver?sslmode=disable' up"

# --- Misc ---
seed:
	@echo "Seeding data... (not implemented)"

clean:
	@rm -rf proto/gen
	@echo "Cleaned generated code (if any)."
