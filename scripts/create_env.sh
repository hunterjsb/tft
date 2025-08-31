#!/bin/bash

# Create .env file on EC2 instance
echo "Creating .env file..."

# Use deployment path if available, otherwise current directory
if [ -d "/opt/tft" ]; then
    cd /opt/tft
fi

# Create .env file template - values need to be manually set on EC2
cat > .env << 'EOF'
# Discord Bot Configuration
DISCORD_TOKEN=YOUR_DISCORD_TOKEN_HERE
OPENAI_API_KEY=YOUR_OPENAI_API_KEY_HERE

# Discord Server Configuration (optional)
GUILD_ID=
CHANNEL_ID=YOUR_CHANNEL_ID_HERE

# OpenAI Configuration (optional)
# MAX_TOKENS=150
# TEMPERATURE=0.7

# Riot API Configuration
RIOT_API_KEY=YOUR_RIOT_API_KEY_HERE
EOF

echo ".env file created with placeholder values"
echo "IMPORTANT: Update /opt/tft/.env on EC2 with actual values!"