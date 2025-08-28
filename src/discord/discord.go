package discord

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hunterjsb/tft/src/riot"
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
				Description: "Player's tagline (e.g., 'NA1')",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "count",
				Description: "Number of games to show (1-10, default: 5)",
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

// handleTFTRecentCommand handles the /tftrecent command
func (b *DiscordBot) handleTFTRecentCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Acknowledge the interaction immediately
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	options := i.ApplicationCommandData().Options

	// Parse command options
	gameName := options[0].StringValue()
	tagLine := options[1].StringValue()

	count := 5 // default
	if len(options) > 2 {
		count = int(options[2].IntValue())
		if count < 1 || count > 10 {
			count = 5
		}
	}

	// Get account information
	account, err := riot.GetAccountByRiotId(gameName, tagLine)
	if err != nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: strPtr(fmt.Sprintf("❌ Could not find player `%s#%s`", gameName, tagLine)),
		})
		return
	}

	// Get recent TFT match IDs
	matchIDs, err := riot.GetTFTMatchIDsByPUUID(account.PUUID, 0, count, nil, nil)
	if err != nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: strPtr("❌ Error fetching match history"),
		})
		return
	}

	if len(matchIDs) == 0 {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: strPtr(fmt.Sprintf("🔍 No TFT games found for `%s#%s`", gameName, tagLine)),
		})
		return
	}

	// Format the response
	response := formatTFTMatches(account, matchIDs)

	// Send the response
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: strPtr(response),
	})
}

// formatTFTMatches formats TFT match data into a nice Discord message
func formatTFTMatches(account *riot.Account, matchIDs []string) string {
	var response strings.Builder

	response.WriteString(fmt.Sprintf("🎮 **Recent TFT Games for %s#%s**\n\n", account.GameName, account.TagLine))

	for i, matchID := range matchIDs {
		// Get detailed match data
		match, err := riot.GetTFTMatchByID(matchID)
		if err != nil {
			response.WriteString(fmt.Sprintf("❌ **Game %d**: Error loading match data\n", i+1))
			continue
		}

		// Find the player's data in the match
		var player *riot.ParticipantDto
		for _, participant := range match.Info.Participants {
			if participant.PUUID == account.PUUID {
				player = &participant
				break
			}
		}

		if player == nil {
			response.WriteString(fmt.Sprintf("❌ **Game %d**: Player not found in match\n", i+1))
			continue
		}

		// Format match info
		gameTime := time.Unix(match.Info.GameCreation/1000, 0)
		duration := int(match.Info.GameLength)
		minutes := duration / 60
		seconds := duration % 60

		// Get placement emoji
		placementEmoji := getPlacementEmoji(player.Placement)

		// Format active traits
		activeTraits := formatActiveTraits(player.Traits)

		response.WriteString(fmt.Sprintf(
			"**%s Game %d** • Set %d\n"+
				"└ %s **#%d** • Level %d • %dm %ds\n"+
				"└ 🏆 %d damage • 💰 %d gold left\n"+
				"└ %s\n"+
				"└ *%s*\n\n",
			placementEmoji, i+1, match.Info.TftSetNumber,
			placementEmoji, player.Placement, player.Level, minutes, seconds,
			player.TotalDamageToPlayers, player.GoldLeft,
			activeTraits,
			gameTime.Format("Jan 2, 3:04 PM"),
		))
	}

	return response.String()
}

// getPlacementEmoji returns an emoji based on placement
func getPlacementEmoji(placement int) string {
	switch placement {
	case 1:
		return "🥇"
	case 2:
		return "🥈"
	case 3:
		return "🥉"
	case 4:
		return "🟢"
	default:
		return "🔴"
	}
}

// formatActiveTraits formats the player's active traits
func formatActiveTraits(traits []riot.TraitDto) string {
	if len(traits) == 0 {
		return "No active traits"
	}

	var activeTraits []string
	for _, trait := range traits {
		if trait.TierCurrent > 0 {
			// Format trait with current tier
			traitStr := fmt.Sprintf("%s %d", trait.Name, trait.TierCurrent)
			if trait.Style >= 3 { // Gold or higher
				traitStr = "⭐ " + traitStr
			}
			activeTraits = append(activeTraits, traitStr)
		}
	}

	if len(activeTraits) == 0 {
		return "No active traits"
	}

	// Limit to first 5 traits to keep message readable
	if len(activeTraits) > 5 {
		activeTraits = activeTraits[:5]
		return "🎯 " + strings.Join(activeTraits, " • ") + " ..."
	}

	return "🎯 " + strings.Join(activeTraits, " • ")
}

// handleChatCommand handles the /chat command
func (b *DiscordBot) handleChatCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Acknowledge the interaction immediately
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	// Get the prompt option
	options := i.ApplicationCommandData().Options
	promptOption := options[0].StringValue()

	// Generate response from OpenAI
	response, err := b.OpenAI.GenerateResponse(context.Background(), promptOption)
	if err != nil {
		fmt.Printf("Error generating response: %v\n", err)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: strPtr("Sorry, I couldn't process your request."),
		})
		return
	}

	// Check if response is too long for Discord (max 2000 chars)
	if len(response) > 2000 {
		// Split into multiple messages
		chunks := chunkString(response, 2000)

		// Edit the initial response with the first chunk
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: strPtr(chunks[0]),
		})

		// Send remaining chunks as follow-up messages
		for _, chunk := range chunks[1:] {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: chunk,
			})
		}
	} else {
		// Edit the deferred response with the OpenAI response
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: strPtr(response),
		})
	}
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}

// Helper function to chunk a string into smaller parts
func chunkString(s string, chunkSize int) []string {
	if len(s) <= chunkSize {
		return []string{s}
	}

	var chunks []string

	// Try to split at line breaks when possible
	lines := strings.Split(s, "\n")
	currentChunk := ""

	for _, line := range lines {
		// If adding this line would exceed the chunk size, start a new chunk
		if len(currentChunk)+len(line)+1 > chunkSize {
			// If current chunk is not empty, add it to chunks
			if currentChunk != "" {
				chunks = append(chunks, currentChunk)
				currentChunk = ""
			}

			// If the line itself is too long, split it by characters
			if len(line) > chunkSize {
				for len(line) > 0 {
					if len(line) <= chunkSize {
						currentChunk = line
						break
					}

					chunks = append(chunks, line[:chunkSize])
					line = line[chunkSize:]
				}
			} else {
				currentChunk = line
			}
		} else {
			// Add line to current chunk
			if currentChunk == "" {
				currentChunk = line
			} else {
				currentChunk += "\n" + line
			}
		}
	}

	// Add the last chunk if it's not empty
	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}
