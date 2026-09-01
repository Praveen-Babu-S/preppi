# Preppi — Agent Instructions

## Project Overview

A scalable microservices platform that connects underprivileged students with volunteer
mentors for free doubt resolution. Students post doubts (text or images), mentors get
smartly routed questions by subject expertise, and can answer or skip.

## Architecture

- **API Gateway** (Gin): Client-facing REST endpoints, JWT auth, role-based access, rate limiting
- **Services** (10): auth, user, question, matching, solution, chat, notification, knowledgebase, analytics, admin
- **Service-to-service**: gRPC + Protobuf
- **Event bus**: NATS JetStream (async/decoupled flows)
- **Database**: PostgreSQL per service (Database-Per-Service pattern), Redis for cache/sessions
- **Storage**: S3-compatible object storage (MinIO locally) for images

## Tech Stack

- Go (1.22+)
- gRPC + Protocol Buffers (proto3)
- Gin framework (REST gateway)
- GORM (ORM)
- NATS JetStream (event bus)
- Redis (cache, sessions, online status)
- PostgreSQL
- Docker + Kubernetes (GKE target)

## Commands

- `make proto-gen` — Generate Go code from proto files
- `make build` — Build all services
- `make test` — Run all tests
- `make lint` — Run golangci-lint
- `make docker-build` — Build Docker images
- `make migrate-db` — Run database migrations
- `docker-compose up -d` — Start local infra (Postgres, Redis, NATS)
- `go test ./...` — Run all Go tests
- `go build ./...` — Compile all packages

## Directory Structure

```
preppi/
├── api-gateway/          # Gin REST API gateway
│   ├── routes/           # Route registry
│   ├── middleware/       # Auth, rate-limit, CORS, logging
│   └── handlers/         # HTTP handlers -> gRPC calls
├── proto/               # Shared protobuf definitions
│   └── <service>/v1/<service>.proto
├── services/            # Microservices
│   └── <name>/
│       ├── main.go       # Service entry point
│       ├── handler/     # gRPC handlers
│       ├── repository/  # DB access (interface + impl)
│       ├── service/     # Business logic
│       └── migrations/  # SQL migrations
├── pkg/                 # Shared libraries
│   ├── database/        # GORM connection helpers
│   ├── auth/            # JWT utils
│   ├── events/          # NATS publisher/subscriber
│   ├── storage/         # S3 client
│   ├── logger/          # zerolog setup
│   ├── config/          # Viper config
│   └── middleware/      # Shared gRPC interceptors
└── deploy/              # Docker + K8s manifests
```

## Code Conventions

- Follow Go standard style; run `gofmt` and `goimports` before committing
- **No business logic in handlers** — delegate to the service layer
- **Repository pattern** for DB access: define an interface, implement with GORM, mock for tests
- All errors wrapped: `fmt.Errorf("operation_name: %w", err)`
- Structured logging with zerolog: `log.Info().Str("user_id", ...).Msg("...")`
- Config via Viper reading from env vars + YAML; never hardcode secrets
- Passwords hashed with bcrypt (never plaintext)
- SQL migrations for schema changes; no GORM AutoMigrate in production

## Proto Conventions

- Package name: `<service>.v1`, e.g. `question.v1`
- Go package: `<service>/v1`
- Use `google.protobuf.Timestamp` for dates, never strings
- Use `google.protobuf.FieldMask` for partial updates
- Pagination: `page_token` + `page_size` in request, `next_page_token` in response
- All RPCs return typed response messages (never bare errors)
- Enums for status fields (question status, user role, etc.)
- Field numbers: 1-15 for frequently used fields (lower wire overhead)

## Database Conventions

- Each service owns its own database; services never reach into another's DB
- Use golang-migrate style migrations (`<timestamp>_<name>.up.sql` / `.down.sql`)
- GORM models embed `gorm.Model` (ID, CreatedAt, UpdatedAt, DeletedAt)
- Explicit foreign keys in migrations
- One database connection pool per service

## Event Conventions

- Event topics: `<domain>.<action>` (e.g. `question.created`, `solution.upvoted`)
- Events are fire-and-forget JSON payloads on NATS JetStream
- Producers publish events after their transaction commits
- Consumers must be idempotent (events may be delivered at-least-once)

## Git Workflow

- Feature branches from `master`
- Conventional commits: `feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`
- Small, focused commits
- PR required for merge to master; CI must pass

## Testing

- Unit tests with Go's `testing` package + GoMock for repository interfaces
- Table-driven tests preferred
- gRPC handlers tested with bufconn (in-process, no network)
- Integration tests via docker-compose against real Postgres/Redis/NATS

## Security

- Never commit secrets; use env vars / K8s secrets
- JWT: RS256 (not HS256) in production, short-lived access + refresh tokens
- Validate all inputs; never trust client-supplied IDs
- Rate-limit public endpoints at the gateway
