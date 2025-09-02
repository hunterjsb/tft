package riot

import (
	"fmt"
	"strings"
)

// GetTFTMatchByID gets a TFT match by match ID
func GetTFTMatchByID(matchID string) (*MatchDto, error) {
	endpoint := fmt.Sprintf("/tft/match/v1/matches/%s", matchID)

	// Extract region from match ID to determine routing
	region := extractRegionFromMatchID(matchID)
	url := buildRegionalRoutingURL(region, endpoint)

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
// region: platform region to determine routing (e.g., "KR", "NA1", "EUW1")
func GetTFTMatchIDsByPUUID(puuid string, start, count int, startTime, endTime *int64) ([]string, error) {
	return GetTFTMatchIDsByPUUIDWithRegion(puuid, "NA1", start, count, startTime, endTime)
}

// GetTFTMatchIDsByPUUIDWithRegion gets match IDs with explicit region for routing
func GetTFTMatchIDsByPUUIDWithRegion(puuid, region string, start, count int, startTime, endTime *int64) ([]string, error) {
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

	url := buildRegionalRoutingURL(region, endpoint)
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

// GetActiveTFTGameByPUUID returns current game information for the given PUUID.
// It first probes match history to infer the player's platform, then queries spectator on that platform.
// Falls back to the previous multi-region scan if platform detection fails.
func GetActiveTFTGameByPUUID(puuid string) (*CurrentGameInfo, error) {
	// Probe across regional routing representatives to infer platform via match ID prefix.
	// NA1 -> AMERICAS routing, EUW1 -> EUROPE routing, KR -> ASIA routing (covers SEA).
	probes := []string{"NA1", "EUW1", "KR"}
	for _, probe := range probes {
		ids, err := GetTFTMatchIDsByPUUIDWithRegion(puuid, probe, 0, 1, nil, nil)
		if err == nil {
			platform := probe
			if len(ids) > 0 {
				platform = extractRegionFromMatchID(ids[0]) // e.g., "NA1_..." -> "NA1"
			}
			return GetActiveTFTGameByPUUIDWithRegion(puuid, platform)
		}
		// On auth errors, no point in continuing the probe loop.
		msg := err.Error()
		if strings.Contains(msg, "status 401") || strings.Contains(msg, "status 403") {
			break
		}
	}

	// Fallback: try common platforms in order of popularity
	regions := []string{"NA1", "EUW1", "KR", "BR1", "LAS", "LAN", "EUNE", "OC1", "JP1", "TR1", "RU", "PH2", "SG2", "TH2", "TW2", "VN2"}
	var lastErr error
	for _, region := range regions {
		info, err := GetActiveTFTGameByPUUIDWithRegion(puuid, region)
		if err == nil {
			return info, nil
		}
		lastErr = err
		// Only continue on 404 (not in game), stop on other errors like 403 (forbidden)
		if !strings.Contains(err.Error(), "status 404") {
			break
		}
	}

	return nil, lastErr
}

// GetActiveTFTGameByPUUIDWithRegion returns current game information for the given PUUID in a specific region.
func GetActiveTFTGameByPUUIDWithRegion(puuid, region string) (*CurrentGameInfo, error) {
	endpoint := fmt.Sprintf("/lol/spectator/tft/v5/active-games/by-puuid/%s", puuid)
	url := buildRegionalURL(region, endpoint)

	var info CurrentGameInfo
	if err := makeAPIRequest(url, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// GetActiveTFTGameByPUUIDWithRegionOrDefault returns current game information, defaulting to NA1 if no region specified.
func GetActiveTFTGameByPUUIDWithRegionOrDefault(puuid, region string) (*CurrentGameInfo, error) {
	if region == "" {
		return GetActiveTFTGameByPUUID(puuid) // Multi-region fallback
	}
	return GetActiveTFTGameByPUUIDWithRegion(puuid, region)
}

// extractRegionFromMatchID extracts the region from a match ID (e.g., "NA1_1234567890" -> "NA1")
func extractRegionFromMatchID(matchID string) string {
	parts := strings.Split(matchID, "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return "NA1" // default fallback
}
