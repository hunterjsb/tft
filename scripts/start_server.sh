#!/bin/bash

# Start the TFT Discord bot
echo "Starting TFT service..."

cd /opt/tft

# Start the bot in background with nohup
nohup ./tft-bot > /var/log/tft.log 2>&1 &

# Save PID
echo $! > /var/run/tft.pid

echo "TFT service started"