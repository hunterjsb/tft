package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// Command definitions
var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "chat",
		Description: "Chat with the LLM bot",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "prompt",
				Description: "Your message to the AI",
				Required:    true,
			},
		},
	},
	{
		Name:        "tftrecent",
		Description: "Get recent TFT games for a player",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "gamename",
				Description: "Player's Riot ID (e.g., 'mubs')",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "tagline",
				Description: "Player's tagline (default: NA1)",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "count",
				Description: "Number of games to show (1-10, default: 5)",
				Required:    false,
			},
		},
	},
	{
		Name:        "lastgame",
		Description: "Get player's last TFT game with detailed info",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "gamename",
				Description: "Player's Riot ID (e.g., 'mubs')",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "tagline",
				Description: "Player's tagline (default: NA1)",
				Required:    false,
			},
		},
	},
}

// NewDiscordBot creates a new Discord bot with the provided configuration
func NewDiscordBot(config *Config) (*DiscordBot, error) {
	session, err := discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	openAI := NewOpenAIClient(config.OpenAIToken, config.MaxTokens, config.Temperature)

	bot := &DiscordBot{
		Session:         session,
		Config:          config,
		OpenAI:          openAI,
		GuildID:         config.GuildID,
		CommandHandlers: make(map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)),
	}

	// Set up command handlers
	bot.CommandHandlers["chat"] = bot.handleChatCommand
	bot.CommandHandlers["tftrecent"] = bot.handleTFTRecentCommand
	bot.CommandHandlers["lastgame"] = bot.handleLastGameCommand

	return bot, nil
}

// Start starts the Discord bot
func (b *DiscordBot) Start() error {
	// Get bot user ID
	user, err := b.Session.User("@me")
	if err != nil {
		return fmt.Errorf("error getting bot user: %w", err)
	}
	b.BotUserID = user.ID

	// Register interaction handler
	b.Session.AddHandler(b.interactionHandler)

	// Open a websocket connection to Discord
	err = b.Session.Open()
	if err != nil {
		return fmt.Errorf("error opening Discord session: %w", err)
	}

	// Register commands
	registeredCommands, err := b.registerCommands()
	if err != nil {
		return fmt.Errorf("error registering commands: %w", err)
	}
	b.Commands = registeredCommands

	fmt.Println("Bot is now running with slash commands registered.")
	return nil
}

// Stop stops the Discord bot and removes commands if configured to do so
func (b *DiscordBot) Stop() error {
	// Remove commands (you can make this configurable if needed)
	fmt.Println("Removing commands...")
	for _, cmd := range b.Commands {
		err := b.Session.ApplicationCommandDelete(b.Session.State.User.ID, b.GuildID, cmd.ID)
		if err != nil {
			fmt.Printf("Error removing command '%s': %v\n", cmd.Name, err)
		}
	}

	return b.Session.Close()
}

// registerCommands registers the defined slash commands
func (b *DiscordBot) registerCommands() ([]*discordgo.ApplicationCommand, error) {
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))

	for i, cmd := range commands {
		registered, err := b.Session.ApplicationCommandCreate(b.Session.State.User.ID, b.GuildID, cmd)
		if err != nil {
			return nil, fmt.Errorf("error creating command '%s': %w", cmd.Name, err)
		}
		registeredCommands[i] = registered
	}

	return registeredCommands, nil
}

// interactionHandler handles Discord interaction events
func (b *DiscordBot) interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Check if it's a command interaction
	if i.Type == discordgo.InteractionApplicationCommand {
		// Get command name
		commandName := i.ApplicationCommandData().Name

		// Check if there's a handler for this command
		if handler, ok := b.CommandHandlers[commandName]; ok {
			handler(s, i)
		}
	}
}

// sendError sends an error embed
func (b *DiscordBot) sendError(s *discordgo.Session, i *discordgo.InteractionCreate, title, description string) {
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       0xff0000,
	}

	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	}); err != nil {
		fmt.Printf("Error editing error response: %v\n", err)
	}
}
