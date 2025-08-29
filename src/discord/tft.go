package discord

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hunterjsb/tft/src/riot"
)

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
	tagLine := DEFAULT_REGION
	if len(options) > 1 && options[1].StringValue() != "" {
		tagLine = options[1].StringValue()
	}

	count := 5 // default
	for i, option := range options {
		if option.Name == "count" && i >= 1 {
			count = int(option.IntValue())
			if count < 1 || count > 10 {
				count = 5
			}
			break
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

// handleLastGameCommand handles the /lastgame command
func (b *DiscordBot) handleLastGameCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
	tagLine := DEFAULT_REGION
	if len(options) > 1 && options[1].StringValue() != "" {
		tagLine = options[1].StringValue()
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

	// Get most recent TFT match
	matchIDs, err := riot.GetTFTMatchIDsByPUUID(account.PUUID, 0, 1, nil, nil)
	if err != nil {
		b.sendError(s, i, "API Error", "Error fetching match history from Riot API")
		return
	}

	if len(matchIDs) == 0 {
		b.sendError(s, i, "No Games Found", fmt.Sprintf("No TFT games found for `%s#%s`", gameName, tagLine))
		return
	}

	// Format and send the detailed response
	embed := b.formatLastGame(account, summoner, matchIDs[0])
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	}); err != nil {
		fmt.Printf("Error editing interaction response: %v\n", err)
	}
}

// formatLastGame formats detailed info for a single TFT match
func (b *DiscordBot) formatLastGame(account *riot.Account, summoner *riot.Summoner, matchID string) *discordgo.MessageEmbed {
	// Get detailed match data
	match, err := riot.GetTFTMatchByID(matchID)
	if err != nil {
		return &discordgo.MessageEmbed{
			Title:       "Error",
			Description: "Could not load match data",
			Color:       0xff0000,
		}
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
		return &discordgo.MessageEmbed{
			Title:       "Error",
			Description: "Player not found in match",
			Color:       0xff0000,
		}
	}

	// Calculate game time and duration
	gameTime := time.Unix(match.Info.GameCreation/1000, 0)
	duration := int(match.Info.GameLength)
	minutes := duration / 60
	seconds := duration % 60

	// Get placement color and emoji
	placementEmoji := b.getPlacementEmoji(player.Placement)
	embedColor := b.getColorByPerformance(float64(player.Placement))

	// Generate AI analysis
	analysis := b.generateGameAnalysis(player)

	// Find the main carry (highest damage dealer or best unit)
	mainCarryIcon := b.getMainCarryIcon(player.Units)

	// Format key champions simply
	keyChampions := b.formatKeyChampions(player.Units)

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s #%d • L%d • %dm %ds", placementEmoji, player.Placement, player.Level, minutes, seconds),
		Color: embedColor,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    fmt.Sprintf("%s#%s", account.GameName, account.TagLine),
			IconURL: fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/15.17.1/img/profileicon/%d.png", summoner.ProfileIconID),
		},
		Description: analysis,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: mainCarryIcon,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Key Champions",
				Value:  keyChampions,
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("%s • %d dmg • %d gold", gameTime.Format("Jan 2 3:04PM"), player.TotalDamageToPlayers, player.GoldLeft),
		},
	}

	return embed
}

// getPlacementEmoji returns emoji based on placement
func (b *DiscordBot) getPlacementEmoji(placement int) string {
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

// generateGameAnalysis creates a 2-sentence AI analysis of the game
func (b *DiscordBot) generateGameAnalysis(player *riot.ParticipantDto) string {
	// Build descriptive analysis prompt
	var prompt strings.Builder
	prompt.WriteString("Describe this TFT Set 15 game in 1-2 short sentences. Bold key units/items. Focus on build and lobby performance:\n\n")

	// Add game result
	prompt.WriteString(fmt.Sprintf("#%d/8, Lvl %d, %d dmg\n", player.Placement, player.Level, player.TotalDamageToPlayers))

	// Add traits (prioritize gold)
	var traitStrs []string
	for _, trait := range player.Traits {
		if trait.TierCurrent > 0 {
			cleanName := b.cleanTraitName(trait.Name)
			if trait.Style >= 3 { // Gold traits first
				traitStrs = append([]string{fmt.Sprintf("%s%d", cleanName, trait.TierCurrent)}, traitStrs...)
			} else {
				traitStrs = append(traitStrs, fmt.Sprintf("%s%d", cleanName, trait.TierCurrent))
			}
		}
	}
	if len(traitStrs) > 0 {
		prompt.WriteString(fmt.Sprintf("Traits: %s\n", strings.Join(traitStrs, ", ")))
	}

	// Add key champions with items
	var unitStrs []string
	for _, unit := range player.Units {
		if unit.Tier >= 2 || len(unit.Items) >= 2 {
			cleanName := b.cleanChampionName(unit.CharacterID)
			unitStr := fmt.Sprintf("%s★%d", cleanName, unit.Tier)

			// Add items for important units
			if len(unit.Items) > 0 {
				var items []string
				for _, itemID := range unit.Items {
					if itemID > 0 {
						items = append(items, fmt.Sprintf("%d", itemID))
					}
				}
				if len(items) > 0 {
					unitStr += fmt.Sprintf("(%s)", strings.Join(items, ","))
				}
			}
			unitStrs = append(unitStrs, unitStr)
		}
	}
	if len(unitStrs) > 0 {
		if len(unitStrs) > 4 {
			unitStrs = unitStrs[:4]
		}
		prompt.WriteString(fmt.Sprintf("Key Units: %s\n", strings.Join(unitStrs, ", ")))
	}

	// Note: Augments not available in current API response

	prompt.WriteString("\nBe concise. Bold important units/items using **bold**.")

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	analysis, err := b.OpenAI.GenerateResponse(ctx, prompt.String())
	if err != nil {
		fmt.Printf("AI Analysis error: %v\n", err)
		return "Analysis failed"
	}

	// Clean up response
	analysis = strings.TrimSpace(analysis)
	if len(analysis) > 180 {
		analysis = analysis[:177] + "..."
	}

	return analysis
}

// getMainCarryIcon finds the main carry champion (highest damage or best unit)
func (b *DiscordBot) getMainCarryIcon(units []riot.UnitDto) string {
	var mainCarry riot.UnitDto

	// First priority: 3-star units with items (likely main carry)
	for _, unit := range units {
		if unit.Tier >= 3 && len(unit.Items) >= 2 {
			if len(mainCarry.Items) < len(unit.Items) || mainCarry.Tier < unit.Tier {
				mainCarry = unit
			}
		}
	}

	// Second priority: any 3-star unit
	if mainCarry.CharacterID == "" {
		for _, unit := range units {
			if unit.Tier >= 3 {
				mainCarry = unit
				break
			}
		}
	}

	// Fallback: highest tier unit with most items
	if mainCarry.CharacterID == "" {
		for _, unit := range units {
			if unit.Tier > mainCarry.Tier || (unit.Tier == mainCarry.Tier && len(unit.Items) > len(mainCarry.Items)) {
				mainCarry = unit
			}
		}
	}

	if mainCarry.CharacterID != "" {
		cleanName := b.cleanChampionName(mainCarry.CharacterID)
		return fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/15.17.1/img/champion/%s.png", cleanName)
	}

	return ""
}

// formatKeyChampions formats champions in a simple, readable way
func (b *DiscordBot) formatKeyChampions(units []riot.UnitDto) string {
	var champLines []string

	// Get top champions sorted by star level
	for tier := 3; tier >= 2; tier-- {
		for _, unit := range units {
			if unit.Tier == tier {
				cleanName := b.cleanChampionName(unit.CharacterID)

				// Format items simply as numbers
				var items []string
				for _, itemID := range unit.Items {
					if itemID > 0 {
						items = append(items, fmt.Sprintf("%d", itemID))
					}
				}

				itemsText := ""
				if len(items) > 0 {
					itemsText = fmt.Sprintf(" [%s]", strings.Join(items, ", "))
				}

				champText := fmt.Sprintf("**%s**★%d%s", cleanName, unit.Tier, itemsText)
				champLines = append(champLines, champText)

				if len(champLines) >= 4 {
					break
				}
			}
		}
		if len(champLines) >= 4 {
			break
		}
	}

	if len(champLines) == 0 {
		return "No key champions"
	}

	return strings.Join(champLines, " • ")
}
