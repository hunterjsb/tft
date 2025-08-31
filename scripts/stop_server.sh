#!/bin/bash

# Stop the TFT Discord bot service
echo "Stopping TFT service..."

# Kill any running TFT processes
pkill -f "tft" || true

# Wait a moment for processes to stop
sleep 2

echo "TFT service stopped"