#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/src/backend"

echo "Running gofmt check..."
unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
    echo "FAILED: the following files are not gofmt-formatted:"
    echo "$unformatted"
    echo "Run 'gofmt -w .' to fix."
    exit 1
fi

echo "Running go vet..."
if ! go vet ./...; then
    echo "FAILED: go vet"
    exit 1
fi

echo "Running golangci-lint..."
if ! golangci-lint run ./...; then
    echo "FAILED: golangci-lint"
    exit 1
fi

echo "Running build verification..."
if ! go build ./...; then
    echo "FAILED: go build"
    exit 1
fi

echo "Running tests..."
if ! go test ./... -v -race; then
    echo "FAILED: go test"
    exit 1
fi

echo ""
echo "All checks passed. Safe to push."
