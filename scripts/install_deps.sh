#!/bin/bash

# Install Go dependencies
echo "Installing Go dependencies..."

# Use deployment path if available, otherwise current directory
if [ -d "/opt/tft" ]; then
    cd /opt/tft
fi

# Download dependencies
go mod download

echo "Dependencies installed"