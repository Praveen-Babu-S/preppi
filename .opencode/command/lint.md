---
description: Run golangci-lint on the whole repo or a package
agent: build
---

Run `golangci-lint run ./...` from the project root. If `$ARGUMENTS` specifies a package path, lint only that. Report any lint errors and fix them if they are straightforward.
