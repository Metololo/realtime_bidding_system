.PHONY: help fmt lint fix check

.DEFAULT_GOAL := help

help:
	@echo "Available commands:"
	@echo ""
	@echo "  make fmt           - Format code using gofmt"
	@echo "  make lint          - Run linter (golangci-lint)"
	@echo "  make fix           - Auto-fix formatting and lint issues"
	@echo "  make check         - Run checks without modifying files (CI mode)"
	@echo "  make install-hooks - Install pre-commit hooks"
	@echo ""

fmt:
	gofmt -s -w .

lint:
	golangci-lint run

fix: fmt
	golangci-lint run --fix

check:
	@echo "Checking formatting..."
	@if [ -n "$$(gofmt -s -l .)" ]; then \
		echo "❌ Code is not formatted. Run 'make fmt'"; \
		gofmt -s -l .; \
		exit 1; \
	fi

	@echo "Running linter..."
	golangci-lint run

install-hooks:
	cp scripts/pre-commit.sh .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit