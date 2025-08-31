package riot

import (
	"os"
	"testing"

	"github.com/hunterjsb/tft/internal/dotenv"
)

func TestMain(m *testing.M) {
	// Try to load environment variables from .env file (ignore errors for CI/CD)
	// In GitHub Actions, the RIOT_API_KEY will be set as an environment variable
	_ = dotenv.LoadDefault()
	_ = dotenv.Load("../../.env")

	if os.Getenv("RIOT_API_KEY") == "" {
		println("RIOT_API_KEY not set, skipping integration tests")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestGetAPIKey(t *testing.T) {
	apiKey := GetAPIKey()
	if apiKey == "" {
		t.Fatal("API key should not be empty when RIOT_API_KEY is set")
	}
}

func TestGetAccountByRiotId_Success(t *testing.T) {
	account, err := GetAccountByRiotId("mubs", "NA1")
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	if account.PUUID == "" {
		t.Error("Account PUUID should not be empty")
	}
	if account.GameName != "mubs" {
		t.Errorf("Expected game name mubs, got %s", account.GameName)
	}
	if account.TagLine != "NA1" {
		t.Errorf("Expected tag line NA1, got %s", account.TagLine)
	}
}

func TestGetAccountByRiotId_NotFound(t *testing.T) {
	_, err := GetAccountByRiotId("ThisPlayerDoesNotExist123456", "NA1")
	if err == nil {
		t.Error("Expected error for non-existent account")
	}
}

func BenchmarkGetAccountByRiotId(b *testing.B) {
	if os.Getenv("RIOT_API_KEY") == "" {
		b.Skip("RIOT_API_KEY not set")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetAccountByRiotId("mubs", "NA1")
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
