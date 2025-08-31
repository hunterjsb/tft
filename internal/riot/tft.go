package riot

import "fmt"

// GetTFTMatchByID gets a TFT match by match ID
func GetTFTMatchByID(matchID string) (*MatchDto, error) {
	endpoint := fmt.Sprintf("/tft/match/v1/matches/%s", matchID)
	url := buildAmericasURL(endpoint)

	var match MatchDto
	if err := makeAPIRequest(url, &match); err != nil {
		return nil, err
	}

	return &match, nil
}

// GetTFTMatchIDsByPUUID gets a list of TFT match IDs by PUUID
// start: defaults to 0, start index
// count: defaults to 20, number of match IDs to return
// startTime/endTime: optional epoch timestamps in seconds
func GetTFTMatchIDsByPUUID(puuid string, start, count int, startTime, endTime *int64) ([]string, error) {
	endpoint := fmt.Sprintf("/tft/match/v1/matches/by-puuid/%s/ids", puuid)

	// Build query parameters
	query := ""
	if start > 0 {
		query += fmt.Sprintf("&start=%d", start)
	}
	if count > 0 && count != 20 { // 20 is the default
		query += fmt.Sprintf("&count=%d", count)
	}
	if startTime != nil {
		query += fmt.Sprintf("&startTime=%d", *startTime)
	}
	if endTime != nil {
		query += fmt.Sprintf("&endTime=%d", *endTime)
	}

	// Remove leading & if present
	if len(query) > 0 && query[0] == '&' {
		query = query[1:]
	}

	url := buildAmericasURL(endpoint)
	if len(query) > 0 {
		url += "&" + query
	}

	var matchIDs []string
	if err := makeAPIRequest(url, &matchIDs); err != nil {
		return nil, err
	}

	return matchIDs, nil
}

// GetTFTMatchIDsByPUUIDSimple gets a list of TFT match IDs by PUUID with default parameters
func GetTFTMatchIDsByPUUIDSimple(puuid string) ([]string, error) {
	return GetTFTMatchIDsByPUUID(puuid, 0, 20, nil, nil)
}

// GetActiveTFTGameByPUUID returns current game information for the given PUUID on NA1.
func GetActiveTFTGameByPUUID(puuid string) (*CurrentGameInfo, error) {
	endpoint := fmt.Sprintf("/lol/spectator/tft/v5/active-games/by-puuid/%s", puuid)
	url := buildNA1URL(endpoint)

	var info CurrentGameInfo
	if err := makeAPIRequest(url, &info); err != nil {
		return nil, err
	}
	return &info, nil
}
