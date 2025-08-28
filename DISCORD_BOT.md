# Discord Bot Documentation

This project includes a Discord bot with OpenAI integration that can be used alongside the TFT functionality.

## Setup

### 1. Install Dependencies

The required dependencies should already be installed, but if needed:

```bash
go mod tidy
```

### 2. Environment Variables

Create a `.env` file in the project root with the following variables:

```env
# Required for Discord Bot
DISCORD_TOKEN=your_discord_bot_token_here
OPENAI_API_KEY=your_openai_api_key_here

# Optional Discord Configuration
GUILD_ID=your_guild_id_here
CHANNEL_ID=your_channel_id_here

# Optional OpenAI Configuration
MAX_TOKENS=150
TEMPERATURE=0.7

# Application Mode
MODE=discord
```

### 3. Discord Bot Setup

1. Go to the [Discord Developer Portal](https://discord.com/developers/applications)
2. Create a new application
3. Go to the "Bot" section and create a bot
4. Copy the bot token and set it as `DISCORD_TOKEN` in your `.env` file
5. Under "Privileged Gateway Intents", enable:
   - Message Content Intent (if you plan to read message content)
   - Server Members Intent (if needed)
6. Go to OAuth2 > URL Generator
7. Select "bot" and "applications.commands" scopes
8. Select the permissions your bot needs (at minimum: Send Messages, Use Slash Commands)
9. Use the generated URL to invite the bot to your server

### 4. OpenAI Setup

1. Go to [OpenAI Platform](https://platform.openai.com/)
2. Create an API key
3. Set it as `OPENAI_API_KEY` in your `.env` file

## Usage

### Running the Discord Bot

Set the mode to discord in your `.env` file:
```env
MODE=discord
```

Then run:
```bash
go run .
```

Or build and run:
```bash
go build -o tft-app .
./tft-app
```

### Bot Commands

The bot supports the following slash commands:

- `/chat <prompt>` - Chat with the AI bot using OpenAI

### Example Usage

1. In Discord, type `/chat Hello, how are you?`
2. The bot will respond using OpenAI's API
3. Long responses are automatically split into multiple messages if they exceed Discord's 2000 character limit

## Configuration Options

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DISCORD_TOKEN` | Your Discord bot token | - | Yes |
| `OPENAI_API_KEY` | Your OpenAI API key | - | Yes |
| `GUILD_ID` | Discord server ID (for faster command registration) | - | No |
| `CHANNEL_ID` | Specific channel ID to respond in | - | No |
| `MAX_TOKENS` | Maximum tokens for OpenAI responses | 150 | No |
| `TEMPERATURE` | OpenAI temperature (0.0-1.0) | 0.7 | No |
| `MODE` | Application mode (`discord` or `tft`) | tft | No |

### Guild ID Configuration

Setting `GUILD_ID` is recommended for development as it makes slash command registration faster. For production bots that should work across multiple servers, leave this empty.

## Architecture

The Discord bot consists of several components:

- **`discord/config.go`** - Configuration loading and validation
- **`discord/discord.go`** - Main bot logic, command handling, and Discord integration
- **`discord/openai.go`** - OpenAI API client wrapper
- **`discord/utils.go`** - Utility functions including graceful shutdown handling

### Key Features

1. **Slash Commands**: Uses Discord's modern slash command interface
2. **OpenAI Integration**: Responds to user prompts using OpenAI's API
3. **Message Chunking**: Automatically splits long responses into multiple messages
4. **Graceful Shutdown**: Properly cleans up commands and connections when stopped
5. **Configurable**: All settings can be adjusted via environment variables

## Troubleshooting

### Common Issues

1. **Bot doesn't respond to commands**
   - Ensure the bot has the "Use Slash Commands" permission
   - Check that `GUILD_ID` is set correctly (if using guild-specific commands)
   - Verify the bot token is correct

2. **OpenAI API errors**
   - Check that your OpenAI API key is valid and has credits
   - Ensure `MAX_TOKENS` is within your API limits

3. **Permission errors**
   - Make sure the bot has "Send Messages" permission in the channel
   - Check that the bot role is high enough in the server hierarchy

### Development Tips

1. Use `GUILD_ID` during development for faster command registration
2. Set `MODE=discord` in your `.env` file to run the bot
3. Monitor the console output for error messages and debugging information
4. Commands are automatically removed when the bot shuts down gracefully

## Switching Between Modes

This application supports two modes:

- **Discord Mode** (`MODE=discord`): Runs the Discord bot
- **TFT Mode** (`MODE=tft`): Runs the TFT demo functionality

You can switch between modes by changing the `MODE` environment variable in your `.env` file.