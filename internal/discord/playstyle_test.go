package discord

import (
	"testing"
	"time"

	"github.com/hunterjsb/tft/internal/riot"
)

func TestFormatPlaystyleAnalysis(t *testing.T) {
	bot := &DiscordBot{}

	// Create mock data
	account := &riot.Account{
		PUUID:    "test-puuid",
		GameName: "TestPlayer",
		TagLine:  "NA1",
	}

	summoner := &riot.Summoner{
		ID:            "test-summoner-id",
		ProfileIconID: 1234,
		SummonerLevel: 50,
	}

	profile := &riot.PlayerProfile{
		PUUID:         "test-puuid",
		AnalyzedGames: 20,
		LastUpdated:   time.Now(),
		PlayStyle: riot.PlayStyleProfile{
			AveragePlacement: 3.5,
			TopFourRate:      0.65,
			EconomyStyle:     "balanced",
			LevelingPattern:  "adaptive",
		},
		CompPreference: riot.CompPreferenceProfile{
			FavoriteTraits: []riot.TraitFrequency{
				{Name: "TFT15_Empyrean", Frequency: 0.8},
				{Name: "TFT15_Protector", Frequency: 0.6},
				{Name: "TFT15_StarGuardian", Frequency: 0.4},
			},
			FavoriteUnits: []riot.UnitFrequency{
				{CharacterID: "TFT15_Jinx", Frequency: 0.7},
				{CharacterID: "TFT15_Akali", Frequency: 0.5},
				{CharacterID: "TFT15_Yasuo", Frequency: 0.3},
			},
		},
		Performance: riot.PerformanceProfile{
			RecentForm:       []int{4, 2, 6, 3, 1, 5, 4, 2},
			ConsistencyScore: 0.6,
			ClimbingTrend:    "stable",
			HighRollGames:    5,
			LowRollGames:     3,
		},
	}

	embed := bot.formatPlaystyleAnalysis(account, summoner, profile)

	// Test embed structure
	if embed.Title != "🎯 TestPlayer's TFT Playstyle" {
		t.Errorf("Expected title '🎯 TestPlayer's TFT Playstyle', got '%s'", embed.Title)
	}

	if embed.Author.Name != "TestPlayer#NA1" {
		t.Errorf("Expected author name 'TestPlayer#NA1', got '%s'", embed.Author.Name)
	}

	if embed.Description != "Analysis based on **20 recent games**" {
		t.Errorf("Expected description 'Analysis based on **20 recent games**', got '%s'", embed.Description)
	}

	// Test that we have all expected fields
	expectedFields := []string{"📊 Performance", "🎮 Playstyle", "📈 Recent Form", "⭐ Favorite Traits", "🏆 Favorite Units"}
	if len(embed.Fields) != len(expectedFields) {
		t.Errorf("Expected %d fields, got %d", len(expectedFields), len(embed.Fields))
	}

	for i, expectedName := range expectedFields {
		if i < len(embed.Fields) && embed.Fields[i].Name != expectedName {
			t.Errorf("Expected field %d name '%s', got '%s'", i, expectedName, embed.Fields[i].Name)
		}
	}

	// Test footer contains level
	if embed.Footer == nil {
		t.Error("Expected footer to be set")
	} else if embed.Footer.Text == "" {
		t.Error("Expected footer text to be set")
	}
}

func TestGetConsistencyDescription(t *testing.T) {
	bot := &DiscordBot{}

	tests := []struct {
		score    float64
		expected string
	}{
		{0.9, "Very High"},
		{0.7, "High"},
		{0.5, "Moderate"},
		{0.3, "Low"},
		{0.1, "Very Low"},
	}

	for _, test := range tests {
		result := bot.getConsistencyDescription(test.score)
		if result != test.expected {
			t.Errorf("For score %.1f, expected '%s', got '%s'", test.score, test.expected, result)
		}
	}
}

func TestFormatFavoriteTraits(t *testing.T) {
	bot := &DiscordBot{}

	traits := []riot.TraitFrequency{
		{Name: "TFT15_Empyrean", Frequency: 0.8},
		{Name: "TFT15_Protector", Frequency: 0.6},
		{Name: "TFT15_StarGuardian", Frequency: 0.4},
	}

	result := bot.formatFavoriteTraits(traits, 2)
	expected := "**Empyrean** 80%\n**Protector** 60%"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test empty traits
	emptyResult := bot.formatFavoriteTraits([]riot.TraitFrequency{}, 5)
	if emptyResult != "No data" {
		t.Errorf("Expected 'No data' for empty traits, got '%s'", emptyResult)
	}
}

func TestFormatFavoriteUnits(t *testing.T) {
	bot := &DiscordBot{}

	units := []riot.UnitFrequency{
		{CharacterID: "TFT15_Jinx", Frequency: 0.7},
		{CharacterID: "TFT15_Akali", Frequency: 0.5},
		{CharacterID: "TFT15_Yasuo", Frequency: 0.3},
	}

	result := bot.formatFavoriteUnits(units, 2)
	expected := "**Jinx** 70%\n**Akali** 50%"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test empty units
	emptyResult := bot.formatFavoriteUnits([]riot.UnitFrequency{}, 5)
	if emptyResult != "No data" {
		t.Errorf("Expected 'No data' for empty units, got '%s'", emptyResult)
	}
}

func TestFormatRecentForm(t *testing.T) {
	bot := &DiscordBot{}

	recentForm := []int{4, 2, 6, 3, 1}
	result := bot.formatRecentForm(recentForm)

	// Should contain placement emojis and numbers
	if !containsString(result, "4") {
		t.Error("Expected result to contain placement '4'")
	}
	if !containsString(result, "🥇1") {
		t.Error("Expected result to contain first place emoji with '1'")
	}

	// Test empty form
	emptyResult := bot.formatRecentForm([]int{})
	if emptyResult != "No recent games" {
		t.Errorf("Expected 'No recent games' for empty form, got '%s'", emptyResult)
	}

	// Test long form (should limit to 8)
	longForm := []int{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3}
	longResult := bot.formatRecentForm(longForm)

	// Count spaces to estimate number of games shown (rough check)
	spaceCount := 0
	for _, char := range longResult {
		if char == ' ' {
			spaceCount++
		}
	}
	// Should show 8 games max (7 spaces between them)
	if spaceCount > 7 {
		t.Errorf("Expected at most 8 games in result, but got more spaces: %d", spaceCount)
	}
}

// Helper function for string containment check
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsString(s[1:], substr)))
}
