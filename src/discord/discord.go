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
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		fmt.Printf("Error acknowledging interaction: %v\n", err)
		return
	}

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

	// Get account and summoner information
	account, err := riot.GetAccountByRiotId(gameName, tagLine)
	if err != nil {
		b.sendError(s, i, "Player Not Found", fmt.Sprintf("Could not find player `%s#%s`", gameName, tagLine))
		return
	}

	summoner, err := riot.GetSummonerByPUUID(account.PUUID)
	if err != nil {
		b.sendError(s, i, "API Error", "Error fetching summoner data")
		return
	}

	// Get recent TFT match IDs
	matchIDs, err := riot.GetTFTMatchIDsByPUUID(account.PUUID, 0, count, nil, nil)
	if err != nil {
		b.sendError(s, i, "API Error", "Error fetching match history from Riot API")
		return
	}

	if len(matchIDs) == 0 {
		b.sendError(s, i, "No Games Found", fmt.Sprintf("No TFT games found for `%s#%s`", gameName, tagLine))
		return
	}

	// Format and send the response
	embed := b.formatTFTMatches(account, summoner, matchIDs)
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	}); err != nil {
		fmt.Printf("Error editing interaction response: %v\n", err)
	}
}

// formatTFTMatches formats TFT match data into a single embed
func (b *DiscordBot) formatTFTMatches(account *riot.Account, summoner *riot.Summoner, matchIDs []string) *discordgo.MessageEmbed {
	var gamesSummary []string
	var gameData []GameData
	avgPlacement := 0.0
	top4Count := 0

	for i, matchID := range matchIDs {
		if i >= 5 { // Limit to 5 games
			break
		}

		// Get detailed match data
		match, err := riot.GetTFTMatchByID(matchID)
		if err != nil {
			gamesSummary = append(gamesSummary, fmt.Sprintf("Game %d: Error loading", i+1))
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
			gamesSummary = append(gamesSummary, fmt.Sprintf("Game %d: Player not found", i+1))
			continue
		}

		avgPlacement += float64(player.Placement)
		if player.Placement <= 4 {
			top4Count++
		}

		// Collect game data for AI analysis
		data := GameData{
			Placement: player.Placement,
			Level:     player.Level,
			Traits:    player.Traits,
			Units:     player.Units,
		}
		gameData = append(gameData, data)

		// Store temporary format for now
		gamesSummary = append(gamesSummary, fmt.Sprintf("#%d L%d Analyzing...",
			player.Placement, player.Level))
	}

	validGames := len(gamesSummary)
	if validGames > 0 {
		avgPlacement /= float64(validGames)
	}

	top4Rate := float64(top4Count) / float64(validGames) * 100

	// Generate AI comp names for all games at once
	if validGames > 0 && len(gameData) > 0 {
		gamesSummary = b.generateAllCompNames(gameData)
	}

	embed := &discordgo.MessageEmbed{
		Title: "Recent TFT Games",
		Color: b.getColorByPerformance(avgPlacement),
		Author: &discordgo.MessageEmbedAuthor{
			Name:    fmt.Sprintf("%s#%s", account.GameName, account.TagLine),
			IconURL: fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/15.17.1/img/profileicon/%d.png", summoner.ProfileIconID),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Performance",
				Value:  fmt.Sprintf("Avg: #%.1f\nTop4: %.0f%%\nLevel: %d", avgPlacement, top4Rate, summoner.SummonerLevel),
				Inline: true,
			},
			{
				Name:   "Recent Games",
				Value:  strings.Join(gamesSummary, "\n"),
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return embed
}

// getColorByPerformance returns color based on average placement
func (b *DiscordBot) getColorByPerformance(avgPlacement float64) int {
	switch {
	case avgPlacement <= 2.5:
		return 0xffd700 // Gold
	case avgPlacement <= 4.0:
		return 0x00ff00 // Green
	case avgPlacement <= 5.5:
		return 0xffaa00 // Orange
	default:
		return 0xff0000 // Red
	}
}

// handleChatCommand handles the /chat command
func (b *DiscordBot) handleChatCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Acknowledge the interaction immediately
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		fmt.Printf("Error acknowledging interaction: %v\n", err)
		return
	}

	// Get the prompt option
	options := i.ApplicationCommandData().Options
	promptOption := options[0].StringValue()

	// Generate response from OpenAI
	response, err := b.OpenAI.GenerateResponse(context.Background(), promptOption)
	if err != nil {
		fmt.Printf("Error generating response: %v\n", err)
		b.sendError(s, i, "AI Error", "Sorry, I couldn't process your request.")
		return
	}

	// Create an embed for the AI response
	embed := &discordgo.MessageEmbed{
		Title:       "AI Response",
		Description: response,
		Color:       0x00ff00,
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Powered by OpenAI",
		},
	}

	// Check if response is too long for Discord embed description (max 4096 chars)
	if len(response) > 4000 {
		embed.Description = response[:4000] + "..."
	}

	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	}); err != nil {
		fmt.Printf("Error editing interaction response: %v\n", err)
	}
}

// sendError sends an error embed
func (b *DiscordBot) sendError(s *discordgo.Session, i *discordgo.InteractionCreate, title, description string) {
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       0xff0000,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	}); err != nil {
		fmt.Printf("Error editing error response: %v\n", err)
	}
}

// cleanTraitName removes API prefixes and makes trait names user-friendly
func (b *DiscordBot) cleanTraitName(traitName string) string {
	// Remove TFT15_ prefix (or any TFTxx_ prefix)
	if strings.HasPrefix(traitName, "TFT") {
		if underscoreIndex := strings.Index(traitName, "_"); underscoreIndex != -1 {
			traitName = traitName[underscoreIndex+1:]
		}
	}

	return traitName
}

// GameData holds data for AI analysis
type GameData struct {
	Placement int
	Level     int
	Traits    []riot.TraitDto
	Units     []riot.UnitDto
}

// generateAllCompNames uses AI to create comp names for all games in one call
func (b *DiscordBot) generateAllCompNames(games []GameData) []string {
	if b.OpenAI == nil || len(games) == 0 {
		// Fallback without AI
		result := make([]string, len(games))
		for i, game := range games {
			mainTrait := "Unknown"
			for _, trait := range game.Traits {
				if trait.TierCurrent > 0 {
					mainTrait = b.cleanTraitName(trait.Name)
					break
				}
			}
			result[i] = fmt.Sprintf("#%d L%d %s", game.Placement, game.Level, mainTrait)
		}
		return result
	}

	// Build detailed prompt with champions and items
	var prompt strings.Builder
	prompt.WriteString("Create short 2-3 word TFT comp names. Include carry champion names when relevant. Respond with just the names, one per line:\n\n")

	for i, game := range games {
		prompt.WriteString(fmt.Sprintf("Game %d:\n", i+1))

		// Add traits
		var activeTraits []string
		for _, trait := range game.Traits {
			if trait.TierCurrent > 0 {
				cleanName := b.cleanTraitName(trait.Name)
				activeTraits = append(activeTraits, fmt.Sprintf("%s%d", cleanName, trait.TierCurrent))
			}
		}
		prompt.WriteString(fmt.Sprintf("Traits: %s\n", strings.Join(activeTraits, ", ")))

		// Add key champions (3-star and 2-star units)
		var keyChamps []string
		for _, unit := range game.Units {
			cleanName := b.cleanChampionName(unit.CharacterID)
			if unit.Tier >= 2 { // 2-star or higher
				keyChamps = append(keyChamps, fmt.Sprintf("%s%d", cleanName, unit.Tier))
			}
		}
		if len(keyChamps) > 0 {
			prompt.WriteString(fmt.Sprintf("Key Champions: %s\n", strings.Join(keyChamps, ", ")))
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString("Examples: 'Jinx Sniper', 'Smolder Carry', 'Reroll Bruiser', 'Fast 9', 'Flex Board'. Focus on main carry champions.")

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	response, err := b.OpenAI.GenerateResponse(ctx, prompt.String())
	if err != nil {
		// Fallback without AI
		result := make([]string, len(games))
		for i, game := range games {
			mainTrait := "Unknown"
			for _, trait := range game.Traits {
				if trait.TierCurrent > 0 {
					mainTrait = b.cleanTraitName(trait.Name)
					break
				}
			}
			result[i] = fmt.Sprintf("#%d L%d %s", game.Placement, game.Level, mainTrait)
		}
		return result
	}

	// Parse AI response
	compNames := strings.Split(strings.TrimSpace(response), "\n")

	// Create final summaries
	result := make([]string, len(games))
	for i, game := range games {
		if i < len(compNames) {
			compName := strings.TrimSpace(compNames[i])
			// Remove "Game X:" prefix if present
			if colonIndex := strings.Index(compName, ":"); colonIndex != -1 {
				compName = strings.TrimSpace(compName[colonIndex+1:])
			}
			// Limit length
			if len(compName) > 15 {
				compName = compName[:15]
			}
			if compName == "" {
				compName = "Unknown"
			}
			result[i] = fmt.Sprintf("#%d L%d %s", game.Placement, game.Level, compName)
		} else {
			result[i] = fmt.Sprintf("#%d L%d Unknown", game.Placement, game.Level)
		}
	}

	return result
}

// cleanChampionName removes API prefixes from champion names
func (b *DiscordBot) cleanChampionName(championID string) string {
	// Remove TFT15_ prefix
	if strings.HasPrefix(championID, "TFT") {
		if underscoreIndex := strings.Index(championID, "_"); underscoreIndex != -1 {
			championID = championID[underscoreIndex+1:]
		}
	}
	return championID
}
