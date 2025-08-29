package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hunterjsb/tft/src/riot"
)

// TFTHandler handles TFT-related Discord commands
type TFTHandler struct{}

// NewTFTHandler creates a new TFT handler
func NewTFTHandler() *TFTHandler {
	return &TFTHandler{}
}

// HandleRecentCommand handles the /tftrecent command
func (h *TFTHandler) HandleRecentCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	// Get account information
	account, err := riot.GetAccountByRiotId(gameName, tagLine)
	if err != nil {
		errorEmbed := h.createErrorEmbed("Player Not Found",
			fmt.Sprintf("Could not find player `%s#%s`", gameName, tagLine))
		if _, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &errorEmbed,
		}); editErr != nil {
			fmt.Printf("Error editing interaction response: %v\n", editErr)
		}
		return
	}

	// Get recent TFT match IDs
	matchIDs, err := riot.GetTFTMatchIDsByPUUID(account.PUUID, 0, count, nil, nil)
	if err != nil {
		errorEmbed := h.createErrorEmbed("API Error", "Error fetching match history from Riot API")
		if _, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &errorEmbed,
		}); editErr != nil {
			fmt.Printf("Error editing interaction response: %v\n", editErr)
		}
		return
	}

	if len(matchIDs) == 0 {
		noGamesEmbed := h.createWarningEmbed("No Games Found",
			fmt.Sprintf("No TFT games found for `%s#%s`", gameName, tagLine))
		if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &noGamesEmbed,
		}); err != nil {
			fmt.Printf("Error editing interaction response: %v\n", err)
		}
		return
	}

	// Format and send the response
	embeds := h.formatMatches(account, matchIDs)
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &embeds,
	}); err != nil {
		fmt.Printf("Error editing interaction response: %v\n", err)
	}
}

// formatMatches formats TFT match data using Discord embeds
func (h *TFTHandler) formatMatches(account *riot.Account, matchIDs []string) []*discordgo.MessageEmbed {
	var embeds []*discordgo.MessageEmbed

	// Header embed with player info
	headerEmbed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🎮 Recent TFT Games for %s#%s", account.GameName, account.TagLine),
		Description: fmt.Sprintf("Showing %d recent games", len(matchIDs)),
		Color:       0x0099ff,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
	embeds = append(embeds, headerEmbed)

	for i, matchID := range matchIDs {
		// Get detailed match data
		match, err := riot.GetTFTMatchByID(matchID)
		if err != nil {
			embeds = append(embeds, h.createMatchErrorEmbed(i+1, "Error loading match data"))
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
			embeds = append(embeds, h.createMatchErrorEmbed(i+1, "Player not found in match"))
			continue
		}

		// Create match embed
		matchEmbed := h.createMatchEmbed(i+1, match, player)
		embeds = append(embeds, matchEmbed)
	}

	return embeds
}

// createMatchEmbed creates an embed for a single match
func (h *TFTHandler) createMatchEmbed(gameNum int, match *riot.MatchDto, player *riot.ParticipantDto) *discordgo.MessageEmbed {
	// Format match info
	gameTime := time.Unix(match.Info.GameCreation/1000, 0)
	duration := int(match.Info.GameLength)
	minutes := duration / 60
	seconds := duration % 60

	// Get placement emoji and color
	placementEmoji := h.getPlacementEmoji(player.Placement)
	embedColor := h.getPlacementColor(player.Placement)

	// Format active traits
	activeTraits := h.formatActiveTraits(player.Traits)

	return &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s Game %d - Set %d", placementEmoji, gameNum, match.Info.TftSetNumber),
		Color: embedColor,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📊 Performance",
				Value:  fmt.Sprintf("**Placement:** #%d\n**Level:** %d\n**Duration:** %dm %ds", player.Placement, player.Level, minutes, seconds),
				Inline: true,
			},
			{
				Name:   "⚔️ Combat Stats",
				Value:  fmt.Sprintf("**Damage:** %d\n**Gold Left:** %d", player.TotalDamageToPlayers, player.GoldLeft),
				Inline: true,
			},
			{
				Name:   "🎯 Active Traits",
				Value:  activeTraits,
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: gameTime.Format("Jan 2, 3:04 PM"),
		},
	}
}

// createMatchErrorEmbed creates an error embed for a failed match load
func (h *TFTHandler) createMatchErrorEmbed(gameNum int, errorMsg string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("❌ Game %d - Error", gameNum),
		Description: errorMsg,
		Color:       0xff0000,
	}
}

// createErrorEmbed creates error message embed
func (h *TFTHandler) createErrorEmbed(title, description string) []*discordgo.MessageEmbed {
	return []*discordgo.MessageEmbed{
		{
			Title:       fmt.Sprintf("❌ %s", title),
			Description: description,
			Color:       0xff0000,
			Timestamp:   time.Now().Format(time.RFC3339),
		},
	}
}

// createWarningEmbed creates warning message embed
func (h *TFTHandler) createWarningEmbed(title, description string) []*discordgo.MessageEmbed {
	return []*discordgo.MessageEmbed{
		{
			Title:       fmt.Sprintf("🔍 %s", title),
			Description: description,
			Color:       0xffaa00,
			Timestamp:   time.Now().Format(time.RFC3339),
		},
	}
}

// getPlacementEmoji returns an emoji based on placement
func (h *TFTHandler) getPlacementEmoji(placement int) string {
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

// getPlacementColor returns a color based on placement for Discord components
func (h *TFTHandler) getPlacementColor(placement int) int {
	switch placement {
	case 1:
		return 0xffd700 // Gold
	case 2:
		return 0xc0c0c0 // Silver
	case 3:
		return 0xcd7f32 // Bronze
	case 4:
		return 0x00ff00 // Green
	default:
		return 0xff0000 // Red
	}
}

// formatActiveTraits formats the player's active traits
func (h *TFTHandler) formatActiveTraits(traits []riot.TraitDto) string {
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
