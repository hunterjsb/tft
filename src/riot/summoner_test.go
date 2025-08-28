package riot

import (
	"os"
	"testing"
)

func TestGetSummonerByPUUID_Success(t *testing.T) {
	if os.Getenv("RIOT_API_KEY") == "" {
		t.Skip("RIOT_API_KEY not set")
	}

	account, err := GetAccountByRiotId("mubs", "NA1")
	if err != nil {
		t.Fatalf("Failed to get account for test: %v", err)
	}

	summoner, err := GetSummonerByPUUID(account.PUUID)
	if err != nil {
		t.Fatalf("Failed to get summoner: %v", err)
	}

	if summoner.PUUID != account.PUUID {
		t.Errorf("Expected PUUID %s, got %s", account.PUUID, summoner.PUUID)
	}
	if summoner.SummonerLevel <= 0 {
		t.Error("Summoner level should be positive")
	}
}

func TestGetSummonerByPUUID_InvalidPUUID(t *testing.T) {
	if os.Getenv("RIOT_API_KEY") == "" {
		t.Skip("RIOT_API_KEY not set")
	}

	_, err := GetSummonerByPUUID("invalid-puuid-12345")
	if err == nil {
		t.Error("Expected error for invalid PUUID")
	}
}

func TestGetSummonerByRiotId_Success(t *testing.T) {
	if os.Getenv("RIOT_API_KEY") == "" {
		t.Skip("RIOT_API_KEY not set")
	}

	summoner, account, err := GetSummonerByRiotId("mubs", "NA1")
	if err != nil {
		t.Fatalf("Failed to get summoner by Riot ID: %v", err)
	}

	if account.GameName != "mubs" {
		t.Errorf("Expected game name mubs, got %s", account.GameName)
	}
	if summoner.PUUID != account.PUUID {
		t.Errorf("Summoner PUUID should match account PUUID")
	}
	if summoner.SummonerLevel <= 0 {
		t.Error("Summoner level should be positive")
	}
}

func TestGetSummonerByRiotId_InvalidAccount(t *testing.T) {
	if os.Getenv("RIOT_API_KEY") == "" {
		t.Skip("RIOT_API_KEY not set")
	}

	_, _, err := GetSummonerByRiotId("NonExistentPlayer123456", "NA1")
	if err == nil {
		t.Error("Expected error for non-existent account")
	}
}
func BenchmarkGetSummonerByPUUID(b *testing.B) {
	if os.Getenv("RIOT_API_KEY") == "" {
		b.Skip("RIOT_API_KEY not set")
	}

	account, err := GetAccountByRiotId("mubs", "NA1")
	if err != nil {
		b.Fatalf("Failed to get account for benchmark: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetSummonerByPUUID(account.PUUID)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkGetSummonerByRiotId(b *testing.B) {
	if os.Getenv("RIOT_API_KEY") == "" {
		b.Skip("RIOT_API_KEY not set")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := GetSummonerByRiotId("mubs", "NA1")
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
