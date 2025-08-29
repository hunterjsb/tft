package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hunterjsb/tft/src/riot"
)

// TFTHandler handles TFT-related Discord commands
type TFTHandler struct {
	openAI *OpenAIClient
}

// NewTFTHandler creates a new TFT handler
func NewTFTHandler(openAIToken string, maxTokens int, temperature float64) *TFTHandler {
	return &TFTHandler{
		openAI: NewOpenAIClient(openAIToken, maxTokens, temperature),
	}
}

// HandleRecentCommand handles the /tftrecent command
func (h *TFTHandler) HandleRecentCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		fmt.Printf("Error acknowledging interaction: %v\n", err)
		return
	}

	options := i.ApplicationCommandData().Options
	gameName := options[0].StringValue()
	tagLine := options[1].StringValue()

	count := 5
	if len(options) > 2 {
		count = int(options[2].IntValue())
		if count < 1 || count > 10 {
			count = 5
		}
	}

	account, err := riot.GetAccountByRiotId(gameName, tagLine)
	if err != nil {
		h.sendError(s, i, "Player not found", fmt.Sprintf("`%s#%s`", gameName, tagLine))
		return
	}

	matchIDs, err := riot.GetTFTMatchIDsByPUUID(account.PUUID, 0, count, nil, nil)
	if err != nil {
		h.sendError(s, i, "API Error", "Failed to fetch match history")
		return
	}

	if len(matchIDs) == 0 {
		h.sendWarning(s, i, "No games found", fmt.Sprintf("`%s#%s`", gameName, tagLine))
		return
	}

	embeds, components := h.formatMatches(account, matchIDs)
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &embeds,
		Components: &components,
	}); err != nil {
		fmt.Printf("Error editing interaction response: %v\n", err)
	}
}

// formatMatches formats TFT match data with AI analysis
func (h *TFTHandler) formatMatches(account *riot.Account, matchIDs []string) ([]*discordgo.MessageEmbed, []discordgo.MessageComponent) {
	var matches []*riot.MatchDto
	var players []*riot.ParticipantDto

	// Collect match data
	for _, matchID := range matchIDs {
		match, err := riot.GetTFTMatchByID(matchID)
		if err != nil {
			continue
		}

		for _, participant := range match.Info.Participants {
			if participant.PUUID == account.PUUID {
				matches = append(matches, match)
				players = append(players, &participant)
				break
			}
		}
	}

	if len(matches) == 0 {
		return []*discordgo.MessageEmbed{}, []discordgo.MessageComponent{}
	}

	// Generate AI analysis
	analysis := h.generateBuildAnalysis(matches, players)

	// Create main embed
	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s#%s Recent Games", account.GameName, account.TagLine),
		Color: h.getAverageColor(players),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Performance",
				Value:  h.formatPerformanceSummary(players),
				Inline: true,
			},
			{
				Name:   "Recent Games",
				Value:  h.formatGamesList(players),
				Inline: true,
			},
			{
				Name:   "Build Analysis",
				Value:  analysis,
				Inline: false,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Create interactive components with short custom IDs
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Detailed View",
					Style:    discordgo.PrimaryButton,
					CustomID: "tft_detail",
				},
				discordgo.Button{
					Label:    "Compare Builds",
					Style:    discordgo.SecondaryButton,
					CustomID: "tft_compare",
				},
				discordgo.Button{
					Label:    "Suggestions",
					Style:    discordgo.SuccessButton,
					CustomID: "tft_suggest",
				},
			},
		},
	}

	return []*discordgo.MessageEmbed{embed}, components
}

// generateBuildAnalysis uses OpenAI to analyze builds
func (h *TFTHandler) generateBuildAnalysis(matches []*riot.MatchDto, players []*riot.ParticipantDto) string {
	if h.openAI == nil {
		return "Analysis unavailable"
	}

	prompt := h.buildAnalysisPrompt(matches, players)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	analysis, err := h.openAI.GenerateResponse(ctx, prompt)
	if err != nil {
		return "Analysis failed"
	}

	// Truncate if too long
	if len(analysis) > 800 {
		return analysis[:797] + "..."
	}
	return analysis
}

// buildAnalysisPrompt creates a prompt for OpenAI
func (h *TFTHandler) buildAnalysisPrompt(matches []*riot.MatchDto, players []*riot.ParticipantDto) string {
	var prompt strings.Builder
	prompt.WriteString("Analyze these TFT games briefly (max 150 words). Focus on patterns, strengths, weaknesses:\n\n")

	for i, player := range players {
		match := matches[i]
		traits := h.getActiveTraits(player.Traits)

		prompt.WriteString(fmt.Sprintf("Game %d: #%d place, Level %d, Set %d\n",
			i+1, player.Placement, player.Level, match.Info.TftSetNumber))
		prompt.WriteString(fmt.Sprintf("Traits: %s\n", strings.Join(traits, ", ")))
		prompt.WriteString(fmt.Sprintf("Damage: %d, Gold: %d\n\n",
			player.TotalDamageToPlayers, player.GoldLeft))
	}

	prompt.WriteString("Provide: 1) Overall performance trend 2) Build consistency 3) Key improvement areas")
	return prompt.String()
}

// formatPerformanceSummary creates a concise performance summary
func (h *TFTHandler) formatPerformanceSummary(players []*riot.ParticipantDto) string {
	if len(players) == 0 {
		return "No data"
	}

	totalPlacement := 0
	top4Count := 0
	avgLevel := 0

	for _, player := range players {
		totalPlacement += player.Placement
		if player.Placement <= 4 {
			top4Count++
		}
		avgLevel += player.Level
	}

	avgPlacement := float64(totalPlacement) / float64(len(players))
	top4Rate := float64(top4Count) / float64(len(players)) * 100
	avgLevel = avgLevel / len(players)

	return fmt.Sprintf("Avg: #%.1f | Top4: %.0f%% | Lvl: %d",
		avgPlacement, top4Rate, avgLevel)
}

// formatGamesList creates a compact games list
func (h *TFTHandler) formatGamesList(players []*riot.ParticipantDto) string {
	var lines []string

	for i, player := range players {
		if i >= 5 { // Limit to 5 games for space
			break
		}

		placement := h.getPlacementIcon(player.Placement)
		mainTrait := h.getMainTrait(player.Traits)

		lines = append(lines, fmt.Sprintf("%s L%d %s",
			placement, player.Level, mainTrait))
	}

	return strings.Join(lines, "\n")
}

// getActiveTraits returns active trait names
func (h *TFTHandler) getActiveTraits(traits []riot.TraitDto) []string {
	var active []string
	for _, trait := range traits {
		if trait.TierCurrent > 0 {
			active = append(active, fmt.Sprintf("%s%d", trait.Name, trait.TierCurrent))
		}
	}
	return active
}

// getMainTrait returns the strongest trait
func (h *TFTHandler) getMainTrait(traits []riot.TraitDto) string {
	var main riot.TraitDto
	for _, trait := range traits {
		if trait.TierCurrent > main.TierCurrent {
			main = trait
		}
	}
	if main.Name == "" {
		return "No traits"
	}
	return fmt.Sprintf("%s%d", main.Name, main.TierCurrent)
}

// getPlacementIcon returns a compact placement indicator
func (h *TFTHandler) getPlacementIcon(placement int) string {
	switch placement {
	case 1:
		return "#1"
	case 2:
		return "#2"
	case 3:
		return "#3"
	case 4:
		return "#4"
	default:
		return fmt.Sprintf("#%d", placement)
	}
}

// getAverageColor returns color based on average performance
func (h *TFTHandler) getAverageColor(players []*riot.ParticipantDto) int {
	if len(players) == 0 {
		return 0x808080
	}

	totalPlacement := 0
	for _, player := range players {
		totalPlacement += player.Placement
	}
	avgPlacement := float64(totalPlacement) / float64(len(players))

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

// sendError sends an error embed
func (h *TFTHandler) sendError(s *discordgo.Session, i *discordgo.InteractionCreate, title, desc string) {
	embed := []*discordgo.MessageEmbed{{
		Title:       title,
		Description: desc,
		Color:       0xff0000,
	}}
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &embed})
}

// sendWarning sends a warning embed
func (h *TFTHandler) sendWarning(s *discordgo.Session, i *discordgo.InteractionCreate, title, desc string) {
	embed := []*discordgo.MessageEmbed{{
		Title:       title,
		Description: desc,
		Color:       0xffaa00,
	}}
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &embed})
}
