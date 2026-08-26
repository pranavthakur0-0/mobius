#!/usr/bin/env bash
# ==============================================================================
# Mobius Local Setup & Verification Script
# ==============================================================================

set -euo pipefail

echo "==> Setting up Mobius development environment..."

# 1. Check Go installation
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.24+ (https://go.dev/dl/)"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo "✅ Found Go: $GO_VERSION"

# 2. Check for .env file
if [ ! -f ".env" ]; then
    echo "⚠️  No .env file found. Creating from .env.example..."
    cp .env.example .env
    echo "ℹ️  Please edit .env with your API keys."
else
    echo "✅ Found .env configuration file."
fi

# 3. Create required runtime directories
mkdir -p .mobius/artifacts .mobius/events bin

# 4. Tidy dependencies
echo "==> Tidying dependencies..."
go mod tidy

# 5. Run tests
echo "==> Running tests..."
go test -v ./...

# 6. Build binary
echo "==> Building binary..."
go build -o bin/mobius ./cmd/mobius

echo ""
echo "🎉 Mobius is ready! Run './bin/mobius' to start the interactive REPL."
