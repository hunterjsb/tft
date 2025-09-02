#!/bin/bash

# Health check for TFT Discord bot
echo "Checking TFT service health..."

# Wait a moment for bot to fully initialize
sleep 3

# Check if process is running
if pgrep -f "tft-bot" > /dev/null; then
    echo "TFT service is running"
    exit 0
else
    echo "TFT service is not running"
    exit 1
fi