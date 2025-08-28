#!/bin/bash

set -e

# Change to script directory
cd "$(dirname "$0")"

echo "🎯 TFT Riot API Test Suite"
echo "========================="

# Parse command line arguments
LINT_ONLY=false
SKIP_TESTS=false
VERBOSE=false

while [[ $# -gt 0 ]]; do
  case $1 in
    --lint-only)
      LINT_ONLY=true
      shift
      ;;
    --skip-tests)
      SKIP_TESTS=true
      shift
      ;;
    --verbose)
      VERBOSE=true
      shift
      ;;
    -h|--help)
      echo "Usage: $0 [OPTIONS]"
      echo "Options:"
      echo "  --lint-only    Run only linting, skip tests"
      echo "  --skip-tests   Run only linting and build, skip tests"
      echo "  --verbose      Enable verbose output"
      echo "  -h, --help     Show this help message"
      exit 0
      ;;
    *)
      echo "Unknown option $1"
      exit 1
      ;;
  esac
done

# Step 1: Build check
echo "🔨 Building packages..."
if ! go build -v ./src/riot/...; then
    echo "❌ Build failed"
    exit 1
fi
echo "✅ Build successful"

# Step 2: Vet check
echo "🔍 Running go vet..."
if ! go vet ./src/riot/...; then
    echo "❌ Go vet failed"
    exit 1
fi
echo "✅ Go vet passed"

# Step 3: Lint check (if golangci-lint is available)
echo "🧹 Running linter..."
if command -v golangci-lint &> /dev/null; then
    if ! golangci-lint run ./src/riot/...; then
        echo "❌ Linting failed"
        exit 1
    fi
    echo "✅ Linting passed"
else
    echo "⚠️  golangci-lint not installed, skipping lint check"
    echo "   Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin"
fi

# Exit early if lint-only mode
if [ "$LINT_ONLY" = true ]; then
    echo "🎉 Lint checks completed successfully!"
    exit 0
fi

# Exit early if skip-tests mode
if [ "$SKIP_TESTS" = true ]; then
    echo "🎉 Build and lint completed successfully!"
    exit 0
fi

# Step 4: Load API key from .env if not already set
if [ -z "$RIOT_API_KEY" ] && [ -f ".env" ]; then
    export $(grep -v '^#' .env | xargs)
    echo "📄 Loaded API key from .env file"
fi

# Check if RIOT_API_KEY is set
if [ -z "$RIOT_API_KEY" ]; then
    echo "❌ RIOT_API_KEY not found"
    echo "   Set RIOT_API_KEY environment variable or add to .env file"
    echo "   You can still run: $0 --lint-only"
    exit 1
fi

# Step 5: Run tests
echo "🧪 Running integration tests..."
TEST_FLAGS="-race -timeout=5m"
if [ "$VERBOSE" = true ]; then
    TEST_FLAGS="$TEST_FLAGS -v"
fi

if go test $TEST_FLAGS ./src/riot/...; then
    echo "✅ All tests passed"
else
    echo "❌ Some tests failed"
    exit 1
fi

echo ""
echo "🎉 All checks passed successfully!"
echo "   Build ✅ Vet ✅ Lint ✅ Tests ✅"
