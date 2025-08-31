package riot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func GetAPIKey() string {
	return os.Getenv("RIOT_API_KEY")
}

func init() {
	http.DefaultClient.Timeout = time.Second * 10
}

// buildURL constructs a Riot API URL with the given base URL and endpoint
func buildURL(baseURL, endpoint string) string {
	return fmt.Sprintf("%s%s?api_key=%s", baseURL, endpoint, GetAPIKey())
}

// buildAmericasURL constructs a URL for the Americas region
func buildAmericasURL(endpoint string) string {
	return buildURL(RIOT_AMERICAS_URL, endpoint)
}

// buildNA1URL constructs a URL for the NA1 region
func buildNA1URL(endpoint string) string {
	return buildURL(RIOT_NA1_URL, endpoint)
}

// buildRegionalURL constructs a URL for any region based on region code (for spectator API)
func buildRegionalURL(region, endpoint string) string {
	baseURL := GetRegionalURL(region)
	return buildURL(baseURL, endpoint)
}

// buildRegionalRoutingURL constructs a URL for regional routing (for match history API)
func buildRegionalRoutingURL(region, endpoint string) string {
	baseURL := GetRegionalRoutingURL(region)
	return buildURL(baseURL, endpoint)
}

// makeAPIRequest is a generic function that handles HTTP boilerplate for Riot API requests
func makeAPIRequest(url string, result interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, result)
}

func GetAccountByRiotId(gameName, tagLine string) (*Account, error) {
	endpoint := fmt.Sprintf("/riot/account/v1/accounts/by-riot-id/%s/%s", gameName, tagLine)
	url := buildAmericasURL(endpoint)

	var account Account
	if err := makeAPIRequest(url, &account); err != nil {
		return nil, err
	}

	return &account, nil
}
