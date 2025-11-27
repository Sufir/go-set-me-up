#!/usr/bin/env bash
set -euo pipefail

# Format code and simplify syntax
echo "[1/5] gofmt (fix)"
gofmt -s -w .

# Rearrange imports and add missing ones
echo "[2/5] goimports (fix)"
goimports -w .

# Static analysis
echo "[3/5] go vet"
go vet ./...

# Lint with automatic fixes where possible
echo "[4/5] golangci-lint (fix)"
# Note: enable fixer plugins and apply available fixes
golangci-lint run --fix ./...

# Run tests
echo "[5/5] go test"
go test ./...
