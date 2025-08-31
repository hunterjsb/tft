#!/bin/bash

# Create .env file on EC2 instance from AWS Parameter Store
echo "Creating .env file from AWS Parameter Store..."

# Use deployment path if available, otherwise current directory
if [ -d "/opt/tft" ]; then
    cd /opt/tft
fi

# Function to get parameter from AWS SSM
get_param() {
    aws ssm get-parameter --name "/tft/$1" --with-decryption --query 'Parameter.Value' --output text 2>/dev/null || echo ""
}

# Get values from Parameter Store
DISCORD_TOKEN=$(get_param "DISCORD_TOKEN")
OPENAI_API_KEY=$(get_param "OPENAI_API_KEY")
CHANNEL_ID=$(get_param "CHANNEL_ID")
RIOT_API_KEY=$(get_param "RIOT_API_KEY")
GUILD_ID=$(get_param "GUILD_ID")

# Create .env file
cat > .env << EOF
# Discord Bot Configuration
DISCORD_TOKEN=${DISCORD_TOKEN}
OPENAI_API_KEY=${OPENAI_API_KEY}

# Discord Server Configuration (optional)
GUILD_ID=${GUILD_ID}
CHANNEL_ID=${CHANNEL_ID}

# OpenAI Configuration (optional)
# MAX_TOKENS=150
# TEMPERATURE=0.7

# Riot API Configuration
RIOT_API_KEY=${RIOT_API_KEY}
EOF

echo ".env file created from AWS Parameter Store"