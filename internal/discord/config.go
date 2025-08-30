package discord

import (
	"fmt"
	"os"
	"strconv"
)

func LoadConfig() (*Config, error) {
	maxTokens := 150
	if maxTokensStr := os.Getenv("MAX_TOKENS"); maxTokensStr != "" {
		if mt, err := strconv.Atoi(maxTokensStr); err == nil {
			maxTokens = mt
		}
	}

	temperature := 0.7
	if tempStr := os.Getenv("TEMPERATURE"); tempStr != "" {
		if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
			temperature = temp
		}
	}

	return &Config{
		DiscordToken: os.Getenv("DISCORD_TOKEN"),
		OpenAIToken:  os.Getenv("OPENAI_API_KEY"),
		GuildID:      os.Getenv("GUILD_ID"),
		ChannelID:    os.Getenv("CHANNEL_ID"),
		MaxTokens:    maxTokens,
		Temperature:  temperature,
	}, nil
}

func (c *Config) Validate() error {
	if c.DiscordToken == "" {
		return fmt.Errorf("DISCORD_TOKEN is required")
	}
	if c.OpenAIToken == "" {
		return fmt.Errorf("OPENAI_API_KEY is required")
	}
	return nil
}
