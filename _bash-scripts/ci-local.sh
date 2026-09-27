#!/usr/bin/env bash
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$here/.." && pwd)"
cd "$repo_root"

echo "CI local script - running: web build, go build, golangci-lint, tests"

fail() { echo "ERROR: $*" >&2; exit 1; }

need() {
  command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

run_web_build() {
  if [ -d webresources ]; then
    echo "==> Installing webresources dependencies"
    (cd webresources && npm ci)
    echo "==> Building webresources"
    (cd webresources && npm run build)
  else
    echo "==> webresources directory not found; skipping web build"
  fi
}

build_go_binaries() {
  echo "==> Building Go binaries (./cmd/...)"
  go build -v ./cmd/...
}

run_golangci_lint() {
  echo "==> Running golangci-lint"
  # Prefer local binary
  if command -v golangci-lint >/dev/null 2>&1; then
    golangci-lint run --config .golangci.yml ./...
    return
  fi

  # Try to install via `go install`
  if command -v go >/dev/null 2>&1; then
    echo "golangci-lint not found; attempting 'go install'..."
    GOBIN="$(go env GOPATH 2>/dev/null)/bin"
    mkdir -p "$GOBIN"
    GOPATH_BIN="$GOBIN"
    export PATH="$GOPATH_BIN:$PATH"
    if go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.0; then
      echo "installed golangci-lint to $GOPATH_BIN"
      golangci-lint run --config .golangci.yml ./...
      return
    else
      echo "go install failed; will try Docker fallback"
    fi
  fi

  # Docker fallback
  if command -v docker >/dev/null 2>&1; then
    echo "Running golangci-lint in Docker (fallback)"
    docker run --rm -v "$PWD":/app -w /app golangci/golangci-lint:v1.59.0 golangci-lint run --config .golangci.yml ./...
    return
  fi

  fail "golangci-lint not available (install locally or install Docker)"
}

run_tests() {
  echo "==> Running go tests"
  go test ./internal/*** -cover
}

main() {
  # Ensure basic tools exist for steps we will run
  need go
  need npm || true

  run_web_build
  build_go_binaries
  run_golangci_lint
  run_tests
}

main "$@"
