package riot

import (
	"os"
	"strings"
	"testing"
)

func TestGetTFTMatchByID_Success(t *testing.T) {
	if os.Getenv("RIOT_API_KEY") == "" {
		t.Skip("RIOT_API_KEY not set")
	}

	account, err := GetAccountByRiotId("mubs", "NA1")
	if err != nil {
		t.Fatalf("Failed to get account for test: %v", err)
	}

	matchIDs, err := GetTFTMatchIDsByPUUIDSimple(account.PUUID)
	if err != nil {
		t.Fatalf("Failed to get match IDs: %v", err)
	}

	if len(matchIDs) == 0 {
		t.Skip("No TFT matches found for test account")
	}

	match, err := GetTFTMatchByID(matchIDs[0])
	if err != nil {
		t.Fatalf("Failed to get TFT match: %v", err)
	}

	if match.Metadata.MatchID != matchIDs[0] {
		t.Errorf("Expected match ID %s, got %s", matchIDs[0], match.Metadata.MatchID)
	}
	if match.Info.GameLength <= 0 {
		t.Error("Game length should be positive")
	}
	if match.Info.TftSetNumber <= 0 {
		t.Error("TFT set number should be positive")
	}
}

func TestGetTFTMatchByID_InvalidMatchID(t *testing.T) {
	if os.Getenv("RIOT_API_KEY") == "" {
		t.Skip("RIOT_API_KEY not set")
	}

	_, err := GetTFTMatchByID("INVALID_MATCH_ID")
	if err == nil {
		t.Error("Expected error for invalid match ID")
	}
}

func TestGetTFTMatchIDsByPUUIDSimple_Success(t *testing.T) {
	if os.Getenv("RIOT_API_KEY") == "" {
		t.Skip("RIOT_API_KEY not set")
	}

	account, err := GetAccountByRiotId("mubs", "NA1")
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	matchIDs, err := GetTFTMatchIDsByPUUIDSimple(account.PUUID)
	if err != nil {
		t.Fatalf("Failed to get match IDs: %v", err)
	}

	if len(matchIDs) > 20 {
		t.Errorf("Should return at most 20 matches, got %d", len(matchIDs))
	}
}

func TestGetTFTMatchIDsByPUUID_CustomParameters(t *testing.T) {
	if os.Getenv("RIOT_API_KEY") == "" {
		t.Skip("RIOT_API_KEY not set")
	}

	account, err := GetAccountByRiotId("mubs", "NA1")
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	matchIDs, err := GetTFTMatchIDsByPUUID(account.PUUID, 0, 5, nil, nil)
	if err != nil {
		t.Fatalf("Failed to get match IDs with custom count: %v", err)
	}

	if len(matchIDs) > 5 {
		t.Errorf("Should return at most 5 matches, got %d", len(matchIDs))
	}
}

func TestGetTFTMatchIDsByPUUID_InvalidPUUID(t *testing.T) {
	if os.Getenv("RIOT_API_KEY") == "" {
		t.Skip("RIOT_API_KEY not set")
	}

	_, err := GetTFTMatchIDsByPUUIDSimple("invalid-puuid")
	if err == nil {
		t.Error("Expected error for invalid PUUID")
	}
}

func BenchmarkGetTFTMatchByID(b *testing.B) {
	if os.Getenv("RIOT_API_KEY") == "" {
		b.Skip("RIOT_API_KEY not set")
	}

	account, err := GetAccountByRiotId("mubs", "NA1")
	if err != nil {
		b.Fatalf("Failed to get account: %v", err)
	}

	matchIDs, err := GetTFTMatchIDsByPUUIDSimple(account.PUUID)
	if err != nil {
		b.Fatalf("Failed to get match IDs: %v", err)
	}
	if len(matchIDs) == 0 {
		b.Skip("No matches available for benchmark")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetTFTMatchByID(matchIDs[0])
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkGetTFTMatchIDsByPUUIDSimple(b *testing.B) {
	if os.Getenv("RIOT_API_KEY") == "" {
		b.Skip("RIOT_API_KEY not set")
	}

	account, err := GetAccountByRiotId("mubs", "NA1")
	if err != nil {
		b.Fatalf("Failed to get account: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetTFTMatchIDsByPUUIDSimple(account.PUUID)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkGetTFTMatchIDsByPUUID(b *testing.B) {
	if os.Getenv("RIOT_API_KEY") == "" {
		b.Skip("RIOT_API_KEY not set")
	}

	account, err := GetAccountByRiotId("mubs", "NA1")
	if err != nil {
		b.Fatalf("Failed to get account: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetTFTMatchIDsByPUUID(account.PUUID, 0, 10, nil, nil)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func TestGetActiveTFTGameByPUUID(t *testing.T) {
	if os.Getenv("RIOT_API_KEY") == "" {
		t.Skip("RIOT_API_KEY not set")
	}

	account, err := GetAccountByRiotId("mubs", "NA1")
	if err != nil {
		t.Fatalf("Failed to get account for test: %v", err)
	}

	info, err := GetActiveTFTGameByPUUIDWithRegion(account.PUUID, "NA1")
	if err != nil {
		// 404 indicates the player is not currently in an active game; treat as non-fatal/skip
		if strings.Contains(err.Error(), "status 404") {
			t.Skip("Player is not currently in an active TFT game (404)")
		}
		t.Fatalf("Failed to get active TFT game: %v", err)
	}

	if info.GameID == 0 {
		t.Error("Expected non-zero GameID")
	}
	if len(info.Participants) == 0 {
		t.Error("Expected at least one participant")
	}

	found := false
	for _, p := range info.Participants {
		if p.PUUID == account.PUUID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected requesting PUUID to be among participants")
	}
}

func TestGetActiveTFTGameByPUUID_InvalidPUUID(t *testing.T) {
	if os.Getenv("RIOT_API_KEY") == "" {
		t.Skip("RIOT_API_KEY not set")
	}

	_, err := GetActiveTFTGameByPUUIDWithRegion("invalid-puuid", "NA1")
	if err == nil {
		t.Error("Expected error for invalid PUUID")
	}
}
