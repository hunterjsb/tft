package discord

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hunterjsb/tft/internal/riot"
)

// handleLobbyCommand handles the /lobby command
func (b *DiscordBot) handleLobbyCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Acknowledge the interaction immediately
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		fmt.Printf("Error acknowledging interaction: %v\n", err)
		return
	}

	// Parse player parameters
	params := ParsePlayerParams(i.ApplicationCommandData().Options)

	// Look up player account and summoner info
	playerResult, err := b.LookupPlayer(s, i, params)
	if err != nil {
		return // Error already sent to Discord
	}

	// Get active game for the player
	gameInfo, err := b.GetActiveGame(s, i, playerResult)
	if err != nil {
		return // Error already sent to Discord
	}

	// Analyze the entire lobby
	analyzer := riot.NewProfileAnalyzer()
	lobby, err := analyzer.AnalyzeLobbyAggregated(gameInfo)
	if err != nil {
		b.sendError(s, i, "Analysis Error", fmt.Sprintf("Could not analyze lobby: %v", err))
		return
	}

	embed := b.formatLobbyAnalysisEmbed(playerResult, gameInfo, lobby)

	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	}); err != nil {
		fmt.Printf("Error editing interaction response: %v\n", err)
	}
}

// formatLobbyAnalysisEmbed formats a lobby analysis into a Discord embed
func (b *DiscordBot) formatLobbyAnalysisEmbed(playerResult *PlayerLookupResult, game *riot.CurrentGameInfo, lobby *riot.LobbyProfile) *discordgo.MessageEmbed {
	playerCount := len(game.Participants)

	// Build contested traits summary (top 5)
	contested := "None"
	if len(lobby.ContestedTraits) > 0 {
		top := 5
		if len(lobby.ContestedTraits) < top {
			top = len(lobby.ContestedTraits)
		}
		var parts []string
		for i := 0; i < top; i++ {
			name := b.cleanTraitName(lobby.ContestedTraits[i].Name)
			percent := int(lobby.ContestedTraits[i].Frequency * 100)
			if percent <= 0 {
				percent = 1 // minimal indicator
			}
			parts = append(parts, fmt.Sprintf("`%s` %d%%", name, percent))
		}
		contested = strings.Join(parts, " ")
	}

	// Build per-player summaries
	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "📊 Lobby Summary",
			Value:  fmt.Sprintf("Players: **%d** • Avg placement: **#%.2f** • Avg Top4: **%.0f%%**", playerCount, lobby.AvgPlacement, lobby.TopFourRate*100),
			Inline: false,
		},
		{
			Name:   "🔥 Most Contested Traits",
			Value:  contested,
			Inline: false,
		},
	}

	// Map to check "you"
	you := playerResult.Account.PUUID

	// Build individual player lines
	for idx, participant := range game.Participants {
		// Find matching profile
		var profile *riot.PlayerProfile
		for _, p := range lobby.Profiles {
			if p != nil && p.PUUID == participant.PUUID {
				profile = p
				break
			}
		}

		playerLabel := fmt.Sprintf("Player %d", idx+1)
		if participant.PUUID == you {
			playerLabel = "You"
		}

		value := b.formatPlayerSummary(profile)
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("👤 %s", playerLabel),
			Value:  value,
			Inline: true,
		})
	}

	// Balance inline fields in rows of 2 or 3
	// Discord will layout them automatically; we keep Inline true on player fields

	color := b.getColorByPerformance(lobby.AvgPlacement)
	author := &discordgo.MessageEmbedAuthor{
		Name:    playerResult.GetDisplayName(),
		IconURL: playerResult.GetProfileIconURL(),
	}

	title := "TFT Lobby Analysis"
	if game.GameID != 0 {
		title = fmt.Sprintf("TFT Lobby Analysis • Game %d", game.GameID)
	}

	return &discordgo.MessageEmbed{
		Title:     title,
		Color:     color,
		Author:    author,
		Fields:    fields,
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Mode: %s • Queue: %d • Region: %s", game.GameMode, game.GameQueueConfigID, game.PlatformID),
		},
	}
}

// formatPlayerSummary creates a concise summary for a player's profile
func (b *DiscordBot) formatPlayerSummary(p *riot.PlayerProfile) string {
	if p == nil || p.AnalyzedGames == 0 {
		return "_No recent data_"
	}

	avg := p.PlayStyle.AveragePlacement
	top4 := p.PlayStyle.TopFourRate * 100
	econ := capitalizeFirst(p.PlayStyle.EconomyStyle)
	leveling := capitalizeFirst(p.PlayStyle.LevelingPattern)

	// Pick up to 2 favorite traits
	var favTraits []string
	for i, t := range p.CompPreference.FavoriteTraits {
		if i >= 2 {
			break
		}
		if t.Name == "" {
			continue
		}
		favTraits = append(favTraits, b.cleanTraitName(t.Name))
	}
	fav := "Unknown"
	if len(favTraits) > 0 {
		fav = strings.Join(favTraits, ", ")
	}

	// Threat indicator
	threat := ""
	switch {
	case avg <= 3.2 && top4 >= 60:
		threat = "🔥 "
	case avg <= 3.8 && top4 >= 50:
		threat = "💎 "
	}

	return fmt.Sprintf("%s#%.1f • Top4 %.0f%%\nStyle: %s/%s • Fav: %s", threat, avg, top4, econ, leveling, fav)
}
