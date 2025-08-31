package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/hunterjsb/tft/internal/riot"
	"github.com/sashabaranov/go-openai"
)

// DiscordBot represents a Discord bot
type DiscordBot struct {
	Session         *discordgo.Session
	Config          *Config
	OpenAI          *OpenAIClient
	BotUserID       string
	GuildID         string
	Commands        []*discordgo.ApplicationCommand
	CommandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

// Config holds Discord bot configuration
type Config struct {
	DiscordToken string
	OpenAIToken  string
	GuildID      string
	ChannelID    string
	MaxTokens    int
	Temperature  float64
}

// OpenAIClient wraps the OpenAI API client
type OpenAIClient struct {
	client      *openai.Client
	maxTokens   int
	temperature float32
}

// GameData holds TFT match data for AI analysis
type GameData struct {
	Placement int
	Level     int
	Traits    []riot.TraitDto
	Units     []riot.UnitDto
}
