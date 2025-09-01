#!/bin/bash

# Build the TFT Discord bot
echo "Building TFT server..."

# Use deployment path if available, otherwise current directory
if [ -d "/opt/tft" ]; then
    cd /opt/tft
fi

# Set Go PATH
export PATH=$PATH:/usr/local/go/bin

# Build the application
go build -o tft-bot ./cmd/discordbot

# Make executable
chmod +x tft-bot

echo "TFT server built successfully"