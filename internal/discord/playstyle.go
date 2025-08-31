package discord

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/hunterjsb/tft/internal/riot"
)

// handlePlaystyleCommand handles the /playstyle command
func (b *DiscordBot) handlePlaystyleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	// Analyze the player's playstyle using our profiling system
	analyzer := riot.NewProfileAnalyzer()
	profile, err := analyzer.AnalyzePlayer(playerResult.Account.PUUID)
	if err != nil {
		b.sendError(s, i, "Analysis Error", fmt.Sprintf("Could not analyze playstyle: %v", err))
		return
	}

	// Format and send the response
	embed := b.formatPlaystyleAnalysis(playerResult, profile)
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	}); err != nil {
		fmt.Printf("Error editing interaction response: %v\n", err)
	}
}

// formatPlaystyleAnalysis formats the player profile into a Discord embed
func (b *DiscordBot) formatPlaystyleAnalysis(playerResult *PlayerLookupResult, profile *riot.PlayerProfile) *discordgo.MessageEmbed {
	// Get color based on performance
	embedColor := b.getColorByPerformance(profile.PlayStyle.AveragePlacement)

	// Build performance summary
	performanceSummary := fmt.Sprintf("**Avg Placement:** #%.1f\n**Top 4 Rate:** %.0f%%\n**Consistency:** %s\n**Trend:** %s",
		profile.PlayStyle.AveragePlacement,
		profile.PlayStyle.TopFourRate*100,
		b.getConsistencyDescription(profile.Performance.ConsistencyScore),
		capitalizeFirst(profile.Performance.ClimbingTrend),
	)

	// Build playstyle description
	playstyleDesc := fmt.Sprintf("**Economy:** %s\n**Leveling:** %s\n**High Rolls:** %d games\n**Low Rolls:** %d games",
		capitalizeFirst(profile.PlayStyle.EconomyStyle),
		capitalizeFirst(profile.PlayStyle.LevelingPattern),
		profile.Performance.HighRollGames,
		profile.Performance.LowRollGames,
	)

	// Get favorite traits (top 5)
	favoriteTraits := b.formatFavoriteTraits(profile.CompPreference.FavoriteTraits, 5)

	// Get favorite units (top 5)
	favoriteUnits := b.formatFavoriteUnits(profile.CompPreference.FavoriteUnits, 5)

	// Get recent form
	recentForm := b.formatRecentForm(profile.Performance.RecentForm)

	// Create main embed
	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("🎯 %s's TFT Playstyle", playerResult.Account.GameName),
		Color: embedColor,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    playerResult.GetDisplayName(),
			IconURL: playerResult.GetProfileIconURL(),
		},
		Description: fmt.Sprintf("Analysis based on **%d recent games**", profile.AnalyzedGames),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📊 Performance",
				Value:  performanceSummary,
				Inline: true,
			},
			{
				Name:   "🎮 Playstyle",
				Value:  playstyleDesc,
				Inline: true,
			},
			{
				Name:   "📈 Recent Form",
				Value:  recentForm,
				Inline: true,
			},
			{
				Name:   "⭐ Favorite Traits",
				Value:  favoriteTraits,
				Inline: true,
			},
			{
				Name:   "🏆 Favorite Units",
				Value:  favoriteUnits,
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Level %d • Analyzed %s", playerResult.GetSummonerLevel(), profile.LastUpdated.Format("Jan 2")),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return embed
}

// getConsistencyDescription converts consistency score to readable text
func (b *DiscordBot) getConsistencyDescription(score float64) string {
	switch {
	case score >= 0.8:
		return "Very High"
	case score >= 0.6:
		return "High"
	case score >= 0.4:
		return "Moderate"
	case score >= 0.2:
		return "Low"
	default:
		return "Very Low"
	}
}

// formatFavoriteTraits formats the top favorite traits
func (b *DiscordBot) formatFavoriteTraits(traits []riot.TraitFrequency, limit int) string {
	if len(traits) == 0 {
		return "No data"
	}

	var traitLines []string
	for i, trait := range traits {
		if i >= limit {
			break
		}

		cleanName := b.cleanTraitName(trait.Name)
		percentage := trait.Frequency * 100

		traitLines = append(traitLines, fmt.Sprintf("**%s** %.0f%%", cleanName, percentage))
	}

	if len(traitLines) == 0 {
		return "No traits found"
	}

	return strings.Join(traitLines, "\n")
}

// formatFavoriteUnits formats the top favorite units
func (b *DiscordBot) formatFavoriteUnits(units []riot.UnitFrequency, limit int) string {
	if len(units) == 0 {
		return "No data"
	}

	var unitLines []string
	for i, unit := range units {
		if i >= limit {
			break
		}

		cleanName := b.cleanChampionName(unit.CharacterID)
		percentage := unit.Frequency * 100

		unitLines = append(unitLines, fmt.Sprintf("**%s** %.0f%%", cleanName, percentage))
	}

	if len(unitLines) == 0 {
		return "No units found"
	}

	return strings.Join(unitLines, "\n")
}

// formatRecentForm formats recent game placements with emojis
func (b *DiscordBot) formatRecentForm(recentForm []int) string {
	if len(recentForm) == 0 {
		return "No recent games"
	}

	var formEmojis []string
	for _, placement := range recentForm {
		emoji := b.getPlacementEmoji(placement)
		formEmojis = append(formEmojis, fmt.Sprintf("%s%d", emoji, placement))
	}

	// Show most recent games first, but limit to 8 to fit in embed
	if len(formEmojis) > 8 {
		formEmojis = formEmojis[len(formEmojis)-8:]
	}

	return strings.Join(formEmojis, " ")
}

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
