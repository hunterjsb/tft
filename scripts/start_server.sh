#!/bin/bash

# Start the TFT Discord bot
echo "Starting TFT service..."

cd /opt/tft

# Debug: Log environment and file status
echo "=== DEBUG INFO $(date) ===" >> /tmp/tft-debug.log
echo "Current directory: $(pwd)" >> /tmp/tft-debug.log
echo "Bot executable exists: $(test -f ./tft-bot && echo 'YES' || echo 'NO')" >> /tmp/tft-debug.log
echo "Bot executable permissions: $(ls -la ./tft-bot)" >> /tmp/tft-debug.log
echo ".env file exists: $(test -f ./.env && echo 'YES' || echo 'NO')" >> /tmp/tft-debug.log
echo ".env contents:" >> /tmp/tft-debug.log
cat .env >> /tmp/tft-debug.log 2>/dev/null || echo "Failed to read .env" >> /tmp/tft-debug.log
echo "=== END DEBUG INFO ===" >> /tmp/tft-debug.log

# Start the bot in background with nohup
nohup ./tft-bot > /var/log/tft.log 2>&1 &

# Save PID
echo $! > /var/run/tft.pid

# Wait a moment and check if process is still running
sleep 2
if pgrep -f "tft-bot" > /dev/null; then
    echo "TFT service started successfully"
else
    echo "TFT service failed to start - check logs" >> /tmp/tft-debug.log
    echo "Last 10 lines of log:" >> /tmp/tft-debug.log
    tail -10 /var/log/tft.log >> /tmp/tft-debug.log 2>/dev/null || echo "No log file found" >> /tmp/tft-debug.log
fi

echo "TFT service startup complete"