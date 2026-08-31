---
description: Generate Go code from proto files
agent: build
mode: subagent
---

Run `make proto-gen` to regenerate Go code from all proto files. If a specific proto was targeted, run protoc for just that file. Verify the generated `.pb.go` files exist and report any errors.
