#!/bin/sh

STAGED_GO_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep ".go$")

if [ -z "$STAGED_GO_FILES" ]; then
    exit 0
fi

echo "--- 🛠️  Pre-commit: Formatting & Staging ---"
echo "$STAGED_GO_FILES" | xargs -r gofmt -s -w
echo "$STAGED_GO_FILES" | xargs -r git add

echo "--- 🔍 Pre-commit: Linting (New Changes) ---"
golangci-lint run ./... --new-from-rev=HEAD
if [ $? -ne 0 ]; then
    echo "❌ Linting failed. Fix the errors above and try again."
    exit 1
fi

echo "--- 🏎️  Pre-commit: Running Tests (-race) ---"
go test -race ./...
if [ $? -ne 0 ]; then
    echo "❌ Tests failed or Race Condition detected! Commit aborted."
    exit 1
fi

echo "✅ All checks passed! Committing..."