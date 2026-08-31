---
description: Scaffold a new microservice
agent: build
mode: subagent
---

Use the grpc-service skill to scaffold a new microservice named from `$ARGUMENTS`. Create the full directory structure (`main.go`, `handler/`, `repository/`, `service/`, `migrations/`), the proto file at `proto/<name>/v1/<name>.proto`, and a Dockerfile. Follow all project conventions and report the created file list.
