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
// It tries multiple common regions if no region is specified.
func GetActiveTFTGameByPUUID(puuid string) (*CurrentGameInfo, error) {
	// Try common regions in order of popularity
	regions := []string{"NA1", "EUW1", "KR", "BR1", "LAS", "LAN", "EUNE", "OC1", "JP1", "TR1", "RU", "PH2", "SG2", "TH2", "TW2", "VN2"}

	var lastErr error
	var notFoundErr error
	for _, region := range regions {
		info, err := GetActiveTFTGameByPUUIDWithRegion(puuid, region)
		if err == nil {
			return info, nil
		}

		msg := err.Error()
		// Treat 404 as "not in game" and keep trying other regions
		if strings.Contains(msg, "status 404") {
			notFoundErr = err
			continue
		}

		// Ignore DNS/NXDOMAIN and other transient network errors and try next region
		if strings.Contains(msg, "no such host") ||
			strings.Contains(msg, "temporary failure in name resolution") ||
			strings.Contains(msg, "server misbehaving") ||
			strings.Contains(msg, "i/o timeout") ||
			strings.Contains(msg, "context deadline exceeded") ||
			strings.Contains(msg, "dial tcp") {
			lastErr = err
			continue
		}

		// For other errors (e.g., 401/403), stop iterating
		lastErr = err
		break
	}

	if notFoundErr != nil {
		return nil, notFoundErr
	}

	return nil, lastErr
}

// GetActiveTFTGameByPUUIDWithRegion returns current game information for the given PUUID in a specific region.
func GetActiveTFTGameByPUUIDWithRegion(puuid, region string) (*CurrentGameInfo, error) {
	endpoint := fmt.Sprintf("/tft/spectator/v5/active-games/by-puuid/%s", puuid)
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
