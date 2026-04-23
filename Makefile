.PHONY: help fmt lint fix check install-hooks

.DEFAULT_GOAL := help

help:
	@echo "Available commands:"
	@echo "  make fmt           - Format all code (gofmt)"
	@echo "  make lint          - Run linter on the whole project"
	@echo "  make fix           - Auto-fix fmt and lint issues (Global)"
	@echo "  make check         - CI mode: Check without fixing"
	@echo "  make install-hooks - Setup the git pre-commit hook"
	@echo "  make test          - Run tests with race detection"

test:
	go test -v -race ./... | grep -E "PASS|FAIL|ok"

fmt:
	gofmt -s -w .

lint:
	golangci-lint run ./...

fix:
	gofmt -s -w .
	golangci-lint run --fix

check:
	@echo "Checking formatting..."
	@if [ -n "$$(gofmt -s -l .)" ]; then \
		echo "❌ Code is not formatted. Run 'make fmt' or 'make fix'"; \
		exit 1; \
	fi
	@echo "Running linter..."
	golangci-lint run ./...

install-hooks:
	cp scripts/pre-commit.sh .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
	@echo "✅ Hook installed!"