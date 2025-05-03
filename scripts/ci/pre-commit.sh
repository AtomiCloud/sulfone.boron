#!/usr/bin/env bash

set -eou pipefail

echo "🔨 Running go mod tidy..."
go mod tidy

echo "🔨 Running pre-commit..."
pre-commit run --all
