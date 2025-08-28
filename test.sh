#!/bin/bash

set -e

echo "🎯 TFT Riot API Tests"

# Load API key from .env if not already set
if [ -z "$RIOT_API_KEY" ] && [ -f ".env" ]; then
    export $(grep -v '^#' .env | xargs)
fi

# Check if RIOT_API_KEY is set
if [ -z "$RIOT_API_KEY" ]; then
    echo "❌ RIOT_API_KEY not found"
    echo "Set RIOT_API_KEY environment variable or add to .env file"
    exit 1
fi

echo "✅ Running tests..."
cd "$(dirname "$0")"
go test -v ./src/riot
