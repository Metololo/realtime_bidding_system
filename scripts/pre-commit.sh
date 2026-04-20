#!/bin/sh

echo "Running pre-commit..."
make fix
git add .

make check || exit 1