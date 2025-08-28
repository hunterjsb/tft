package riot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func getAPIKey() string {
	return os.Getenv("RIOT_API_KEY")
}

func init() {
	http.DefaultClient.Timeout = time.Second * 10
}

func GetAccountByRiotId(gameName, tagLine string) (*Account, error) {
	url := fmt.Sprintf("%s/riot/account/v1/accounts/by-riot-id/%s/%s?api_key=%s",
		RIOT_AMERICAS_URL, gameName, tagLine, getAPIKey())

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var account Account
	if err := json.Unmarshal(body, &account); err != nil {
		return nil, err
	}

	return &account, nil
}

func GetSummonerByPUUID(puuid string) (*Summoner, error) {
	url := fmt.Sprintf("%s/lol/summoner/v4/summoners/by-puuid/%s?api_key=%s",
		RIOT_NA1_URL, puuid, getAPIKey())

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var summoner Summoner
	if err := json.Unmarshal(body, &summoner); err != nil {
		return nil, err
	}

	return &summoner, nil
}

func GetSummonerByRiotId(gameName, tagLine string) (*Summoner, *Account, error) {
	account, err := GetAccountByRiotId(gameName, tagLine)
	if err != nil {
		return nil, nil, err
	}

	summoner, err := GetSummonerByPUUID(account.PUUID)
	if err != nil {
		return nil, account, err
	}

	return summoner, account, nil
}
