#!/usr/bin/env bash
# Generate Go code from proto files using buf.
# Requires: buf, protoc-gen-go, protoc-gen-go-grpc on PATH.
set -euo pipefail

cd "$(dirname "$0")/../proto"

if ! command -v buf >/dev/null 2>&1; then
  echo "buf not found. Install with: brew install buf" >&2
  exit 1
fi

echo "Generating Go code from proto files..."
buf generate

echo "Done. Generated files in proto/<service>/v1/."
