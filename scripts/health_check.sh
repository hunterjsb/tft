#!/bin/bash

# Health check for TFT Discord bot
echo "Checking TFT service health..."

# Check if process is running
if pgrep -f "tft-bot" > /dev/null; then
    echo "TFT service is running"
    exit 0
else
    echo "TFT service is not running"
    exit 1
fi