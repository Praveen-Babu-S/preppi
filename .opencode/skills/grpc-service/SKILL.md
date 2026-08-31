---
name: grpc-service
description: Use when creating, scaffolding, or adding a new microservice in the doubt-resolver platform. Trigger on requests like "create a new service", "scaffold a microservice", "add a service for <domain>". Generates the full service structure (main.go, handler, repository, service, migrations, proto, Dockerfile) following project conventions.
---

# gRPC Microservice Scaffolding

## Purpose

Create a new microservice following the doubt-resolver project conventions. The service follows
the layered architecture: gRPC handler → service (business logic) → repository (DB access).

## When to Use

Use this skill when the user asks to:
- Create a new microservice (e.g., "add a chat service", "create a reports service")
- Scaffold a service skeleton
- Generate a new service's structure

## Service Structure

A service must be created at `services/<name>/` with these directories:

```
services/<name>/
├── main.go              # Service entry point: gRPC server + DB + NATS connect
├── handler/             # gRPC handlers (validates request, calls service)
│   └── <name>_handler.go
├── repository/          # Data access (interface + GORM implementation)
│   ├── <name>_repository.go     # Interface
│   ├── <name>_repository_impl.go # GORM implementation
│   └── mock_<name>_repository.go # GoMock mock
├── service/             # Business logic
│   └── <name>_service.go
└── migrations/          # SQL migrations
    └── <timestamp>_init.up.sql
    └── <timestamp>_init.down.sql
```

## Steps

1. **Determine service name** from the user's request (snake_case, e.g. `chat`, `notification`).
2. **Create directories** under `services/<name>/`.
3. **Create the proto file** at `proto/<name>/v1/<name>.proto`. Follow the proto conventions:
   - Package: `<name>.v1`
   - Go package: `<name>/v1`
   - Use `google.protobuf.Timestamp` for dates
   - Typed response messages (never bare errors)
4. **Create `main.go`** with:
   - Viper config loading (port, DB, NATS, Redis from env/YAML)
   - GORM database connection via `pkg/database`
   - gRPC server registration
   - NATS JetStream subscription setup
5. **Create `handler/`** — implements the gRPC service interface. Handlers validate input and delegate to the service layer. No business logic here.
6. **Create `service/`** — contains business logic. All errors wrapped: `fmt.Errorf("operation_name: %w", err)`.
7. **Create `repository/`** — define an interface, implement with GORM. Return errors wrapped.
8. **Create `migrations/`** — initial schema if the service needs tables.
9. **Create Dockerfile** at `deploy/docker/Dockerfile.service` (or generate per-service).

## main.go Template

```go
package main

import (
    "fmt"
    "log"
    "net"

    "google.golang.org/grpc"
    "github.com/rs/zerolog"

    "github.com/<org>/doubt-resolver/pkg/config"
    "github.com/<org>/doubt-resolver/pkg/database"
    "github.com/<org>/doubt-resolver/pkg/logger"
    pb "github.com/<org>/doubt-resolver/proto/<name>/v1"
    "<org>/doubt-resolver/services/<name>/handler"
    "<org>/doubt-resolver/services/<name>/repository"
    "<org>/doubt-resolver/services/<name>/service"
)

func main() {
    cfg := config.Load("<name>")
    log := logger.New(cfg)

    db, err := database.Connect(cfg)
    if err != nil {
        log.Fatal().Err(err).Msg("failed to connect to database")
    }

    repo := repository.New<Name>Repository(db)
    svc := service.New<Name>Service(repo)
    h := handler.New<Name>Handler(svc)

    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
    if err != nil {
        log.Fatal().Err(err).Msg("failed to listen")
    }

    s := grpc.NewServer()
    pb.Register<Name>ServiceServer(s, h)

    log.Info().Int("port", cfg.Port).Msg("<name> service starting")
    if err := s.Serve(lis); err != nil {
        log.Fatal().Err(err).Msg("failed to serve")
    }
}
```

## Rules

- Never put business logic in handlers.
- Always define repository interfaces so they can be mocked for tests.
- All services embed `gorm.Model` in their models.
- Use `pkg/`, `pkg/config`, `pkg/logger`, `pkg/database` shared helpers.
- Structured logging with zerolog: `log.Info().Str("key", "val").Msg("...")`.
