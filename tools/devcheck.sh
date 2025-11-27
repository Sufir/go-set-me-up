#!/usr/bin/env bash
set -euo pipefail

go mod tidy

MODULE=$(go list -m)
GOIMPORTS_BIN="$(go env GOPATH)/bin/goimports"
GOLANGCI_LINT_BIN="$(go env GOPATH)/bin/golangci-lint"

echo "[1/6] gofmt (fix)"
gofmt -s -w .
OUT=$(gofmt -s -l .)
if [ -n "$OUT" ]; then
  echo "$OUT"
  exit 1
fi

echo "[2/6] goimports (fix, -local=$MODULE)"
"$GOIMPORTS_BIN" -w -local "$MODULE" ./
OUT=$("$GOIMPORTS_BIN" -l -local "$MODULE" ./)
if [ -n "$OUT" ]; then
  echo "$OUT"
  exit 1
fi

echo "[3/6] golangci-lint (fix)"
"$GOLANGCI_LINT_BIN" run --config tools/.golangci-lint.yml --timeout 3m --fix ./...

echo "[4/6] go vet"
go vet ./...

echo "[5/6] go mod tidy check"
if ! git diff --exit-code -- go.mod go.sum >/dev/null; then
  echo "go.mod/go.sum изменены после tidy. Закоммитьте эти изменения перед пушем."
fi

echo "[6/6] go test -race"
go test -race ./...
