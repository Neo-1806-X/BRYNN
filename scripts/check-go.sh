#!/bin/sh
# Quality gate for the Go code in apps/api.
# Runs the same way on a laptop (via `make check`, inside a container) and in CI.
set -eu

cd "$(dirname "$0")/../apps/api"

echo "==> go version"
go version

echo "==> gofmt"
unformatted="$(gofmt -l .)"
if [ -n "$unformatted" ]; then
    echo "These files are not gofmt-formatted (run: make fmt):"
    echo "$unformatted"
    exit 1
fi

echo "==> go vet"
go vet ./...

echo "==> go test -race"
go test -race -count=1 ./...

echo "OK"
