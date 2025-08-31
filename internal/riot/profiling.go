package riot

import (
	"fmt"
	"sort"
	"time"
)

// PlayerProfile represents a player's analyzed gameplay patterns
type PlayerProfile struct {
	PUUID          string                `json:"puuid"`
	GameName       string                `json:"gameName,omitempty"`
	TagLine        string                `json:"tagLine,omitempty"`
	ProfileIconID  int64                 `json:"profileIconId,omitempty"`
	AnalyzedGames  int                   `json:"analyzedGames"`
	LastUpdated    time.Time             `json:"lastUpdated"`
	PlayStyle      PlayStyleProfile      `json:"playStyle"`
	CompPreference CompPreferenceProfile `json:"compPreference"`
	ItemPreference ItemPreferenceProfile `json:"itemPreference"`
	Performance    PerformanceProfile    `json:"performance"`
}

// PlayStyleProfile captures how a player typically plays
type PlayStyleProfile struct {
	AggresionLevel   float64 `json:"aggressionLevel"`  // 0-1, based on early game econ choices
	EconomyStyle     string  `json:"economyStyle"`     // "greedy", "balanced", "aggressive"
	LevelingPattern  string  `json:"levelingPattern"`  // "fast", "slow", "adaptive"
	RerollTendency   float64 `json:"rerollTendency"`   // 0-1, how often they reroll vs save
	ContestRate      float64 `json:"contestRate"`      // 0-1, how often they play contested comps
	TopFourRate      float64 `json:"topFourRate"`      // win rate for top 4 placements
	AveragePlacement float64 `json:"averagePlacement"` // 1-8 average placement
}

// CompPreferenceProfile shows what compositions a player prefers
type CompPreferenceProfile struct {
	FavoriteTraits  []TraitFrequency `json:"favoriteTraits"`
	FavoriteUnits   []UnitFrequency  `json:"favoriteUnits"`
	CompFlexibility float64          `json:"compFlexibility"` // 0-1, how often they pivot
	TraitDiversity  float64          `json:"traitDiversity"`  // 0-1, variety of traits played
	MetaFollower    float64          `json:"metaFollower"`    // 0-1, how closely they follow meta
	PreferredCost   []CostPreference `json:"preferredCost"`   // unit cost preference distribution
}

// ItemPreferenceProfile shows item building patterns
type ItemPreferenceProfile struct {
	FavoriteItems     []ItemFrequency `json:"favoriteItems"`
	ItemEfficiency    float64         `json:"itemEfficiency"`    // items completed vs components held
	EarlyItemPriority []string        `json:"earlyItemPriority"` // most common first items
	FlexibleItemUser  float64         `json:"flexibleItemUser"`  // 0-1, item build diversity
	CarryItemFocus    float64         `json:"carryItemFocus"`    // 0-1, items on carry vs utility
}

// PerformanceProfile tracks performance metrics
type PerformanceProfile struct {
	RecentForm        []int   `json:"recentForm"`        // last 10 game placements
	ConsistencyScore  float64 `json:"consistencyScore"`  // 0-1, placement consistency
	ClimbingTrend     string  `json:"climbingTrend"`     // "climbing", "stable", "declining"
	HighRollGames     int     `json:"highRollGames"`     // games with 1st/2nd place
	LowRollGames      int     `json:"lowRollGames"`      // games with 7th/8th place
	AverageGameLength float64 `json:"averageGameLength"` // seconds, indicates early vs late game
}

// Supporting frequency types
type TraitFrequency struct {
	Name      string  `json:"name"`
	Frequency float64 `json:"frequency"` // 0-1, how often this trait appears
	AvgTier   float64 `json:"avgTier"`   // average tier achieved with this trait
}

type UnitFrequency struct {
	CharacterID string   `json:"characterId"`
	Name        string   `json:"name"`
	Frequency   float64  `json:"frequency"` // 0-1, how often this unit is used
	AvgTier     float64  `json:"avgTier"`   // average tier when used
	Items       []string `json:"items"`     // most common items on this unit
}

type ItemFrequency struct {
	ItemID    int      `json:"itemId"`
	Name      string   `json:"name"`
	Frequency float64  `json:"frequency"` // 0-1, how often this item is built
	Units     []string `json:"units"`     // most common units that get this item
}

type CostPreference struct {
	Cost       int     `json:"cost"`       // 1-5 cost units
	Preference float64 `json:"preference"` // 0-1, normalized preference for this cost
}

// ProfileAnalyzer handles the analysis of player data
type ProfileAnalyzer struct {
	MaxGamesToAnalyze int // default 20
	MinGamesRequired  int // default 5
}

// NewProfileAnalyzer creates a new analyzer with default settings
func NewProfileAnalyzer() *ProfileAnalyzer {
	return &ProfileAnalyzer{
		MaxGamesToAnalyze: 20,
		MinGamesRequired:  5,
	}
}

// AnalyzePlayer creates a comprehensive profile for a player
func (pa *ProfileAnalyzer) AnalyzePlayer(puuid string) (*PlayerProfile, error) {
	// Get recent match IDs
	matchIDs, err := GetTFTMatchIDsByPUUID(puuid, 0, pa.MaxGamesToAnalyze, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get match IDs for %s: %w", puuid, err)
	}

	if len(matchIDs) < pa.MinGamesRequired {
		return nil, fmt.Errorf("insufficient games for analysis: %d (minimum %d)", len(matchIDs), pa.MinGamesRequired)
	}

	// Get detailed match data
	var matches []*MatchDto
	for _, matchID := range matchIDs {
		match, err := GetTFTMatchByID(matchID)
		if err != nil {
			continue // skip failed matches
		}
		matches = append(matches, match)
	}

	if len(matches) < pa.MinGamesRequired {
		return nil, fmt.Errorf("insufficient valid matches for analysis: %d", len(matches))
	}

	// Create profile
	profile := &PlayerProfile{
		PUUID:         puuid,
		AnalyzedGames: len(matches),
		LastUpdated:   time.Now(),
	}

	// Extract player-specific data from matches
	var playerData []ParticipantDto
	for _, match := range matches {
		for _, participant := range match.Info.Participants {
			if participant.PUUID == puuid {
				playerData = append(playerData, participant)
				break
			}
		}
	}

	// Analyze different aspects
	profile.PlayStyle = pa.analyzePlayStyle(playerData)
	profile.CompPreference = pa.analyzeCompPreference(playerData)
	profile.ItemPreference = pa.analyzeItemPreference(playerData)
	profile.Performance = pa.analyzePerformance(playerData)

	return profile, nil
}

// AnalyzeLobby creates profiles for all players in an active game
func (pa *ProfileAnalyzer) AnalyzeLobby(gameInfo *CurrentGameInfo) ([]*PlayerProfile, error) {
	var profiles []*PlayerProfile

	for _, participant := range gameInfo.Participants {
		profile, err := pa.AnalyzePlayer(participant.PUUID)
		if err != nil {
			// Create basic profile with just PUUID if analysis fails
			profile = &PlayerProfile{
				PUUID:         participant.PUUID,
				ProfileIconID: participant.ProfileIconID,
				AnalyzedGames: 0,
				LastUpdated:   time.Now(),
			}
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// Analysis helper methods (placeholder implementations)
func (pa *ProfileAnalyzer) analyzePlayStyle(playerData []ParticipantDto) PlayStyleProfile {
	if len(playerData) == 0 {
		return PlayStyleProfile{}
	}

	// Calculate average placement
	totalPlacement := 0
	topFours := 0
	for _, game := range playerData {
		totalPlacement += game.Placement
		if game.Placement <= 4 {
			topFours++
		}
	}

	avgPlacement := float64(totalPlacement) / float64(len(playerData))
	topFourRate := float64(topFours) / float64(len(playerData))

	return PlayStyleProfile{
		AveragePlacement: avgPlacement,
		TopFourRate:      topFourRate,
		EconomyStyle:     pa.determineEconomyStyle(playerData),
		LevelingPattern:  pa.determineLevelingPattern(playerData),
	}
}

func (pa *ProfileAnalyzer) analyzeCompPreference(playerData []ParticipantDto) CompPreferenceProfile {
	traitMap := make(map[string]int)
	unitMap := make(map[string]int)

	for _, game := range playerData {
		// Count trait usage
		for _, trait := range game.Traits {
			if trait.TierCurrent > 0 {
				traitMap[trait.Name]++
			}
		}

		// Count unit usage
		for _, unit := range game.Units {
			unitMap[unit.CharacterID]++
		}
	}

	// Convert to frequency arrays
	var favoriteTraits []TraitFrequency
	for name, count := range traitMap {
		frequency := float64(count) / float64(len(playerData))
		favoriteTraits = append(favoriteTraits, TraitFrequency{
			Name:      name,
			Frequency: frequency,
		})
	}

	var favoriteUnits []UnitFrequency
	for characterID, count := range unitMap {
		frequency := float64(count) / float64(len(playerData))
		favoriteUnits = append(favoriteUnits, UnitFrequency{
			CharacterID: characterID,
			Frequency:   frequency,
		})
	}

	// Sort by frequency
	sort.Slice(favoriteTraits, func(i, j int) bool {
		return favoriteTraits[i].Frequency > favoriteTraits[j].Frequency
	})
	sort.Slice(favoriteUnits, func(i, j int) bool {
		return favoriteUnits[i].Frequency > favoriteUnits[j].Frequency
	})

	return CompPreferenceProfile{
		FavoriteTraits: favoriteTraits,
		FavoriteUnits:  favoriteUnits,
	}
}

func (pa *ProfileAnalyzer) analyzeItemPreference(playerData []ParticipantDto) ItemPreferenceProfile {
	itemMap := make(map[int]int)

	for _, game := range playerData {
		for _, unit := range game.Units {
			for _, itemID := range unit.Items {
				itemMap[itemID]++
			}
		}
	}

	var favoriteItems []ItemFrequency
	totalItems := 0
	for itemID, count := range itemMap {
		totalItems += count
		favoriteItems = append(favoriteItems, ItemFrequency{
			ItemID:    itemID,
			Frequency: float64(count),
		})
	}

	// Normalize frequencies
	for i := range favoriteItems {
		favoriteItems[i].Frequency = favoriteItems[i].Frequency / float64(totalItems)
	}

	sort.Slice(favoriteItems, func(i, j int) bool {
		return favoriteItems[i].Frequency > favoriteItems[j].Frequency
	})

	return ItemPreferenceProfile{
		FavoriteItems: favoriteItems,
	}
}

func (pa *ProfileAnalyzer) analyzePerformance(playerData []ParticipantDto) PerformanceProfile {
	var recentForm []int
	highRolls := 0
	lowRolls := 0
	totalGameTime := 0.0

	for _, game := range playerData {
		recentForm = append(recentForm, game.Placement)
		if game.Placement <= 2 {
			highRolls++
		}
		if game.Placement >= 7 {
			lowRolls++
		}
		totalGameTime += game.TimeEliminated
	}

	// Keep only last 10 games for recent form
	if len(recentForm) > 10 {
		recentForm = recentForm[len(recentForm)-10:]
	}

	avgGameLength := totalGameTime / float64(len(playerData))
	consistencyScore := pa.calculateConsistencyScore(recentForm)

	return PerformanceProfile{
		RecentForm:        recentForm,
		ConsistencyScore:  consistencyScore,
		HighRollGames:     highRolls,
		LowRollGames:      lowRolls,
		AverageGameLength: avgGameLength,
		ClimbingTrend:     pa.determineClimbingTrend(recentForm),
	}
}

// Helper methods for analysis
func (pa *ProfileAnalyzer) determineEconomyStyle(playerData []ParticipantDto) string {
	if len(playerData) == 0 {
		return "unknown"
	}

	totalGoldLeft := 0
	highGoldGames := 0

	for _, game := range playerData {
		totalGoldLeft += game.GoldLeft
		if game.GoldLeft >= 20 {
			highGoldGames++
		}
	}

	avgGoldLeft := float64(totalGoldLeft) / float64(len(playerData))
	highGoldRate := float64(highGoldGames) / float64(len(playerData))

	// Classify based on gold management patterns
	if avgGoldLeft >= 15 || highGoldRate >= 0.6 {
		return "greedy"
	} else if avgGoldLeft <= 5 && highGoldRate <= 0.2 {
		return "aggressive"
	}
	return "balanced"
}

func (pa *ProfileAnalyzer) determineLevelingPattern(playerData []ParticipantDto) string {
	if len(playerData) == 0 {
		return "unknown"
	}

	totalLevel := 0
	highLevelGames := 0

	for _, game := range playerData {
		totalLevel += game.Level
		if game.Level >= 9 {
			highLevelGames++
		}
	}

	avgLevel := float64(totalLevel) / float64(len(playerData))
	highLevelRate := float64(highLevelGames) / float64(len(playerData))

	// Classify based on final levels achieved
	if avgLevel >= 8.5 || highLevelRate >= 0.4 {
		return "fast"
	} else if avgLevel <= 7.0 && highLevelRate <= 0.1 {
		return "slow"
	}
	return "adaptive"
}

func (pa *ProfileAnalyzer) calculateConsistencyScore(placements []int) float64 {
	if len(placements) < 2 {
		return 0.0
	}

	// Calculate standard deviation of placements
	sum := 0
	for _, placement := range placements {
		sum += placement
	}
	mean := float64(sum) / float64(len(placements))

	variance := 0.0
	for _, placement := range placements {
		variance += (float64(placement) - mean) * (float64(placement) - mean)
	}
	variance /= float64(len(placements))

	// Lower standard deviation = higher consistency
	// Convert to 0-1 scale where 1 is most consistent
	stdDev := variance
	return 1.0 / (1.0 + stdDev/4.0) // normalize roughly
}

func (pa *ProfileAnalyzer) determineClimbingTrend(recentForm []int) string {
	if len(recentForm) < 5 {
		return "unknown"
	}

	// Simple trend analysis - compare first half to second half
	mid := len(recentForm) / 2
	firstHalf := recentForm[:mid]
	secondHalf := recentForm[mid:]

	firstAvg := 0.0
	for _, placement := range firstHalf {
		firstAvg += float64(placement)
	}
	firstAvg /= float64(len(firstHalf))

	secondAvg := 0.0
	for _, placement := range secondHalf {
		secondAvg += float64(placement)
	}
	secondAvg /= float64(len(secondHalf))

	// Lower placement numbers are better
	if secondAvg < firstAvg-0.5 {
		return "climbing"
	} else if secondAvg > firstAvg+0.5 {
		return "declining"
	}
	return "stable"
}

// LoadSampleActiveGame loads the sample active game data for testing
func LoadSampleActiveGame() (*CurrentGameInfo, error) {
	// This would load from samples/active_game_sample.json
	// For now, we'll return nil to indicate this should be implemented
	// when we need to test without a live game
	return nil, fmt.Errorf("sample loading not yet implemented - use live game data")
}

// AnalyzeLobbyFromSample analyzes a lobby using sample data for testing
func (pa *ProfileAnalyzer) AnalyzeLobbyFromSample() ([]*PlayerProfile, error) {
	sampleGame, err := LoadSampleActiveGame()
	if err != nil {
		return nil, fmt.Errorf("failed to load sample game: %w", err)
	}
	return pa.AnalyzeLobby(sampleGame)
}

// GetSpectatorInfo returns information about spectating a game
// The encryption key from CurrentGameInfo.Observers.EncryptionKey is used to:
// - Decrypt live spectator data streams from Riot's servers
// - Build live game viewers and analysis tools
// - Access real-time game state for third-party applications
func GetSpectatorInfo(gameInfo *CurrentGameInfo) map[string]interface{} {
	return map[string]interface{}{
		"canSpectate":      gameInfo.Observers.EncryptionKey != "",
		"encryptionKey":    gameInfo.Observers.EncryptionKey,
		"gameId":           gameInfo.GameID,
		"platform":         gameInfo.PlatformID,
		"participantCount": len(gameInfo.Participants),
		"spectatorUrl":     fmt.Sprintf("spectator %s %s %d %s", gameInfo.PlatformID, gameInfo.Observers.EncryptionKey, gameInfo.GameID, gameInfo.PlatformID),
	}
}

// LoadSampleActiveGameFromFile loads sample active game data from JSON file
func LoadSampleActiveGameFromFile(filename string) (*CurrentGameInfo, error) {
	// This would read and unmarshal the JSON file
	// Implementation would go here when needed for testing
	return nil, fmt.Errorf("sample file loading not yet implemented")
}
