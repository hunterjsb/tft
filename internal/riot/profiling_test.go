package riot

import (
	"os"
	"testing"

	"github.com/hunterjsb/tft/internal/dotenv"
)

func TestNewProfileAnalyzer(t *testing.T) {
	analyzer := NewProfileAnalyzer()
	if analyzer.MaxGamesToAnalyze != 20 {
		t.Errorf("Expected MaxGamesToAnalyze to be 20, got %d", analyzer.MaxGamesToAnalyze)
	}
	if analyzer.MinGamesRequired != 5 {
		t.Errorf("Expected MinGamesRequired to be 5, got %d", analyzer.MinGamesRequired)
	}
}

func TestAnalyzePlayer_Success(t *testing.T) {
	// Try to load environment variables from .env file (ignore errors for CI/CD)
	_ = dotenv.LoadDefault()
	_ = dotenv.Load("../../.env")

	if os.Getenv("RIOT_API_KEY") == "" {
		t.Skip("RIOT_API_KEY not set")
	}

	analyzer := NewProfileAnalyzer()

	// Get a test account
	account, err := GetAccountByRiotId("mubs", "NA1")
	if err != nil {
		t.Fatalf("Failed to get test account: %v", err)
	}

	profile, err := analyzer.AnalyzePlayer(account.PUUID)
	if err != nil {
		t.Fatalf("Failed to analyze player: %v", err)
	}

	if profile.PUUID != account.PUUID {
		t.Errorf("Expected PUUID %s, got %s", account.PUUID, profile.PUUID)
	}
	if profile.AnalyzedGames < analyzer.MinGamesRequired {
		t.Errorf("Expected at least %d analyzed games, got %d", analyzer.MinGamesRequired, profile.AnalyzedGames)
	}
	if profile.LastUpdated.IsZero() {
		t.Error("Expected LastUpdated to be set")
	}

	// Test PlayStyle analysis
	if profile.PlayStyle.AveragePlacement < 1 || profile.PlayStyle.AveragePlacement > 8 {
		t.Errorf("Average placement should be between 1-8, got %.2f", profile.PlayStyle.AveragePlacement)
	}
	if profile.PlayStyle.TopFourRate < 0 || profile.PlayStyle.TopFourRate > 1 {
		t.Errorf("Top four rate should be between 0-1, got %.2f", profile.PlayStyle.TopFourRate)
	}

	// Test CompPreference analysis
	if len(profile.CompPreference.FavoriteTraits) == 0 {
		t.Error("Expected some favorite traits to be identified")
	}
	if len(profile.CompPreference.FavoriteUnits) == 0 {
		t.Error("Expected some favorite units to be identified")
	}

	// Test Performance analysis
	if len(profile.Performance.RecentForm) == 0 {
		t.Error("Expected recent form data")
	}
	if profile.Performance.ClimbingTrend == "" {
		t.Error("Expected climbing trend to be set")
	}

	t.Logf("Profile analysis complete:")
	t.Logf("- Analyzed %d games", profile.AnalyzedGames)
	t.Logf("- Average placement: %.2f", profile.PlayStyle.AveragePlacement)
	t.Logf("- Top 4 rate: %.2f%%", profile.PlayStyle.TopFourRate*100)
	t.Logf("- Consistency score: %.2f", profile.Performance.ConsistencyScore)
	t.Logf("- Climbing trend: %s", profile.Performance.ClimbingTrend)
	t.Logf("- Top 3 favorite traits: %v", getTopThreeTraits(profile.CompPreference.FavoriteTraits))
}

func TestAnalyzePlayer_InsufficientGames(t *testing.T) {
	// Try to load environment variables from .env file (ignore errors for CI/CD)
	_ = dotenv.LoadDefault()
	_ = dotenv.Load("../../.env")

	if os.Getenv("RIOT_API_KEY") == "" {
		t.Skip("RIOT_API_KEY not set")
	}

	analyzer := NewProfileAnalyzer()
	analyzer.MinGamesRequired = 100 // Set unreasonably high requirement

	account, err := GetAccountByRiotId("mubs", "NA1")
	if err != nil {
		t.Fatalf("Failed to get test account: %v", err)
	}

	_, err = analyzer.AnalyzePlayer(account.PUUID)
	if err == nil {
		t.Error("Expected error for insufficient games")
	}
}

func TestAnalyzePlayer_InvalidPUUID(t *testing.T) {
	// Try to load environment variables from .env file (ignore errors for CI/CD)
	_ = dotenv.LoadDefault()
	_ = dotenv.Load("../../.env")

	if os.Getenv("RIOT_API_KEY") == "" {
		t.Skip("RIOT_API_KEY not set")
	}

	analyzer := NewProfileAnalyzer()
	_, err := analyzer.AnalyzePlayer("invalid-puuid-12345")
	if err == nil {
		t.Error("Expected error for invalid PUUID")
	}
}

func TestCalculateConsistencyScore(t *testing.T) {
	analyzer := NewProfileAnalyzer()

	// Test perfect consistency (all same placements)
	consistentPlacements := []int{4, 4, 4, 4, 4}
	score := analyzer.calculateConsistencyScore(consistentPlacements)
	if score < 0.8 { // Should be very high for consistent placements
		t.Errorf("Expected high consistency score for identical placements, got %.2f", score)
	}

	// Test inconsistent placements
	inconsistentPlacements := []int{1, 8, 2, 7, 1, 8}
	score = analyzer.calculateConsistencyScore(inconsistentPlacements)
	if score > 0.5 { // Should be lower for inconsistent placements
		t.Errorf("Expected low consistency score for inconsistent placements, got %.2f", score)
	}

	// Test edge case - single game
	singleGame := []int{4}
	score = analyzer.calculateConsistencyScore(singleGame)
	if score != 0.0 {
		t.Errorf("Expected 0 consistency score for single game, got %.2f", score)
	}

	// Test empty slice
	emptyGames := []int{}
	score = analyzer.calculateConsistencyScore(emptyGames)
	if score != 0.0 {
		t.Errorf("Expected 0 consistency score for empty games, got %.2f", score)
	}
}

func TestDetermineClimbingTrend(t *testing.T) {
	analyzer := NewProfileAnalyzer()

	// Test climbing trend (getting better placements)
	climbingForm := []int{6, 5, 4, 3, 2, 1}
	trend := analyzer.determineClimbingTrend(climbingForm)
	if trend != "climbing" {
		t.Errorf("Expected 'climbing' trend, got '%s'", trend)
	}

	// Test declining trend (getting worse placements)
	decliningForm := []int{2, 3, 4, 5, 6, 7}
	trend = analyzer.determineClimbingTrend(decliningForm)
	if trend != "declining" {
		t.Errorf("Expected 'declining' trend, got '%s'", trend)
	}

	// Test stable trend
	stableForm := []int{4, 3, 4, 5, 4, 3}
	trend = analyzer.determineClimbingTrend(stableForm)
	if trend != "stable" {
		t.Errorf("Expected 'stable' trend, got '%s'", trend)
	}

	// Test insufficient data
	shortForm := []int{4, 3}
	trend = analyzer.determineClimbingTrend(shortForm)
	if trend != "unknown" {
		t.Errorf("Expected 'unknown' trend for insufficient data, got '%s'", trend)
	}
}

func TestAnalyzePlayStyle(t *testing.T) {
	analyzer := NewProfileAnalyzer()

	// Create mock participant data
	playerData := []ParticipantDto{
		{Placement: 4, PUUID: "test"},
		{Placement: 2, PUUID: "test"},
		{Placement: 6, PUUID: "test"},
		{Placement: 3, PUUID: "test"},
		{Placement: 1, PUUID: "test"},
	}

	playStyle := analyzer.analyzePlayStyle(playerData)

	expectedAvg := (4.0 + 2.0 + 6.0 + 3.0 + 1.0) / 5.0
	if playStyle.AveragePlacement != expectedAvg {
		t.Errorf("Expected average placement %.2f, got %.2f", expectedAvg, playStyle.AveragePlacement)
	}

	expectedTopFour := 4.0 / 5.0 // 4 out of 5 games are top 4
	if playStyle.TopFourRate != expectedTopFour {
		t.Errorf("Expected top four rate %.2f, got %.2f", expectedTopFour, playStyle.TopFourRate)
	}

	// Test empty data
	emptyPlayStyle := analyzer.analyzePlayStyle([]ParticipantDto{})
	if emptyPlayStyle.AveragePlacement != 0 || emptyPlayStyle.TopFourRate != 0 {
		t.Error("Expected zero values for empty player data")
	}
}

func TestAnalyzeCompPreference(t *testing.T) {
	analyzer := NewProfileAnalyzer()

	// Create mock participant data with traits and units
	playerData := []ParticipantDto{
		{
			PUUID: "test",
			Traits: []TraitDto{
				{Name: "Assassin", TierCurrent: 2},
				{Name: "Challenger", TierCurrent: 3},
			},
			Units: []UnitDto{
				{CharacterID: "TFT9_Akali"},
				{CharacterID: "TFT9_Yasuo"},
			},
		},
		{
			PUUID: "test",
			Traits: []TraitDto{
				{Name: "Assassin", TierCurrent: 4},
				{Name: "Ionia", TierCurrent: 2},
			},
			Units: []UnitDto{
				{CharacterID: "TFT9_Akali"},
				{CharacterID: "TFT9_Irelia"},
			},
		},
	}

	compPref := analyzer.analyzeCompPreference(playerData)

	// Check that traits are tracked
	if len(compPref.FavoriteTraits) == 0 {
		t.Error("Expected some favorite traits")
	}

	// Check that units are tracked
	if len(compPref.FavoriteUnits) == 0 {
		t.Error("Expected some favorite units")
	}

	// Akali should be most frequent (appears in both games)
	akaliFound := false
	for _, unit := range compPref.FavoriteUnits {
		if unit.CharacterID == "TFT9_Akali" && unit.Frequency == 1.0 {
			akaliFound = true
			break
		}
	}
	if !akaliFound {
		t.Error("Expected Akali to have 100% frequency")
	}

	// Assassin should be most frequent trait (appears in both games)
	assassinFound := false
	for _, trait := range compPref.FavoriteTraits {
		if trait.Name == "Assassin" && trait.Frequency == 1.0 {
			assassinFound = true
			break
		}
	}
	if !assassinFound {
		t.Error("Expected Assassin trait to have 100% frequency")
	}
}

func TestAnalyzeItemPreference(t *testing.T) {
	analyzer := NewProfileAnalyzer()

	// Create mock participant data with items
	playerData := []ParticipantDto{
		{
			PUUID: "test",
			Units: []UnitDto{
				{CharacterID: "TFT9_Akali", Items: []int{1, 2, 3}},
				{CharacterID: "TFT9_Yasuo", Items: []int{1, 4}},
			},
		},
		{
			PUUID: "test",
			Units: []UnitDto{
				{CharacterID: "TFT9_Akali", Items: []int{1, 2}},
				{CharacterID: "TFT9_Irelia", Items: []int{5}},
			},
		},
	}

	itemPref := analyzer.analyzeItemPreference(playerData)

	if len(itemPref.FavoriteItems) == 0 {
		t.Error("Expected some favorite items")
	}

	// Item 1 should be most frequent (appears 3 times out of 8 total items)
	item1Found := false
	for _, item := range itemPref.FavoriteItems {
		if item.ItemID == 1 {
			expectedFreq := 3.0 / 8.0
			if item.Frequency != expectedFreq {
				t.Errorf("Expected item 1 frequency %.3f, got %.3f", expectedFreq, item.Frequency)
			}
			item1Found = true
			break
		}
	}
	if !item1Found {
		t.Error("Expected to find item 1 in favorite items")
	}
}

// Helper function for logging
func getTopThreeTraits(traits []TraitFrequency) []string {
	var topThree []string
	for i, trait := range traits {
		if i >= 3 {
			break
		}
		topThree = append(topThree, trait.Name)
	}
	return topThree
}

func BenchmarkAnalyzePlayer(b *testing.B) {
	// Try to load environment variables from .env file (ignore errors for CI/CD)
	_ = dotenv.LoadDefault()
	_ = dotenv.Load("../../.env")

	if os.Getenv("RIOT_API_KEY") == "" {
		b.Skip("RIOT_API_KEY not set")
	}

	account, err := GetAccountByRiotId("mubs", "NA1")
	if err != nil {
		b.Fatalf("Failed to get test account: %v", err)
	}

	analyzer := NewProfileAnalyzer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzePlayer(account.PUUID)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
