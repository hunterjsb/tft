package riot

import "fmt"

func GetSummonerByPUUID(puuid string) (*Summoner, error) {
	endpoint := fmt.Sprintf("/lol/summoner/v4/summoners/by-puuid/%s", puuid)
	url := buildNA1URL(endpoint)

	var summoner Summoner
	if err := makeAPIRequest(url, &summoner); err != nil {
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
