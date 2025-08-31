package discord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/hunterjsb/tft/internal/riot"
)

// PlayerParams holds parsed player information from Discord command options
type PlayerParams struct {
	GameName string
	TagLine  string
	Region   string // Optional, empty string means auto-detect
}

// PlayerLookupResult contains the result of looking up a player
type PlayerLookupResult struct {
	Account  *riot.Account
	Summoner *riot.Summoner // Optional, may be nil
	Params   PlayerParams
}

// ParsePlayerParams extracts player information from Discord command options.
// Expects options in order: [gamename, tagline, region (optional)]
func ParsePlayerParams(options []*discordgo.ApplicationCommandInteractionDataOption) PlayerParams {
	params := PlayerParams{}

	if len(options) > 0 {
		params.GameName = options[0].StringValue()
	}
	if len(options) > 1 {
		params.TagLine = options[1].StringValue()
	}
	if len(options) > 2 && options[2].StringValue() != "" {
		params.Region = options[2].StringValue()
	}

	return params
}

// LookupPlayer performs account and summoner lookup for the given player parameters.
// Returns nil error on success, or sends appropriate error message to Discord on failure.
func (b *DiscordBot) LookupPlayer(s *discordgo.Session, i *discordgo.InteractionCreate, params PlayerParams) (*PlayerLookupResult, error) {
	// Validate required parameters
	if params.GameName == "" {
		b.sendError(s, i, "Invalid Input", "Player name is required")
		return nil, fmt.Errorf("missing game name")
	}
	if params.TagLine == "" {
		b.sendError(s, i, "Invalid Input", "Tagline is required")
		return nil, fmt.Errorf("missing tagline")
	}

	// Look up the account
	account, err := riot.GetAccountByRiotId(params.GameName, params.TagLine)
	if err != nil {
		b.sendError(s, i, "Player Not Found", fmt.Sprintf("Could not find player `%s#%s`", params.GameName, params.TagLine))
		return nil, fmt.Errorf("account lookup failed: %w", err)
	}

	// Try to get summoner info (optional, non-fatal if it fails)
	summoner, _ := riot.GetSummonerByPUUID(account.PUUID)

	return &PlayerLookupResult{
		Account:  account,
		Summoner: summoner,
		Params:   params,
	}, nil
}

// GetActiveGame fetches the active TFT game for a player, respecting the region parameter.
// Returns nil error on success, or sends appropriate error message to Discord on failure.
func (b *DiscordBot) GetActiveGame(s *discordgo.Session, i *discordgo.InteractionCreate, result *PlayerLookupResult) (*riot.CurrentGameInfo, error) {
	gameInfo, err := riot.GetActiveTFTGameByPUUIDWithRegionOrDefault(result.Account.PUUID, result.Params.Region)
	if err != nil {
		if strings.Contains(err.Error(), "status 404") {
			errorMsg := fmt.Sprintf("`%s#%s` is not currently in a TFT game.", result.Account.GameName, result.Account.TagLine)
			if result.Params.Region != "" {
				errorMsg += fmt.Sprintf("\n\n*Searched in region: %s*", result.Params.Region)
			} else {
				errorMsg += "\n\n*Tip: If this player should be in-game, try specifying their server region (e.g., `BR1`, `EUW1`, `KR`)*"
			}
			b.sendError(s, i, "No Active Game", errorMsg)
		} else {
			b.sendError(s, i, "API Error", "Error fetching active game from Riot API")
		}
		return nil, fmt.Errorf("active game lookup failed: %w", err)
	}

	return gameInfo, nil
}

// GetProfileIconURL returns the profile icon URL for a player, using summoner info if available
func (result *PlayerLookupResult) GetProfileIconURL() string {
	iconID := 0
	if result.Summoner != nil {
		iconID = result.Summoner.ProfileIconID
	}
	if iconID > 0 {
		return fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/15.17.1/img/profileicon/%d.png", iconID)
	}
	return ""
}

// GetDisplayName returns a formatted display name for the player
func (result *PlayerLookupResult) GetDisplayName() string {
	return fmt.Sprintf("%s#%s", result.Account.GameName, result.Account.TagLine)
}

// GetSummonerLevel returns the summoner level, or 0 if not available
func (result *PlayerLookupResult) GetSummonerLevel() int {
	if result.Summoner != nil {
		return result.Summoner.SummonerLevel
	}
	return 0
}
