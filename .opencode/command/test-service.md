---
description: Run tests for a specific service or the whole repo
agent: build
---

Run tests for the specified service. If `$ARGUMENTS` is a service name (e.g. `auth`, `question`), run `go test ./services/$ARGUMENTS/...`. If `$ARGUMENTS` is empty or `all`, run `go test ./...`. Report pass/fail summary.
