# Riot Games API Client

A clean, well-structured Go client for the Riot Games API with support for Account, League of Legends Summoner, and Teamfight Tactics Match APIs.

## Features

- **Account API** - Get Riot account information by Riot ID
- **Summoner API** - Get League of Legends summoner data
- **TFT Match API** - Get Teamfight Tactics match history and detailed match data
- **Clean Architecture** - Organized into focused modules with shared HTTP boilerplate
- **Type Safety** - Complete struct definitions for all API responses
- **Error Handling** - Proper error handling for API failures and rate limits

## Setup

1. Get a Riot Games API key from [developer.riotgames.com](https://developer.riotgames.com)
2. Set your API key as an environment variable:
   ```bash
   export RIOT_API_KEY="your-api-key-here"
   ```
   Or create a `.env` file:
   ```
   RIOT_API_KEY=your-api-key-here
   ```

## Usage

### Account API

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/hunterjsb/tft/src/riot"
)

func main() {
    // Get account by Riot ID
    account, err := riot.GetAccountByRiotId("mubs", "NA1")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Account: %s#%s (PUUID: %s)\n", 
        account.GameName, account.TagLine, account.PUUID)
}
```

### Summoner API

```go
// Get summoner data by PUUID
summoner, err := riot.GetSummonerByPUUID(account.PUUID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Summoner Level: %d\n", summoner.SummonerLevel)

// Or get both account and summoner in one call
summoner, account, err := riot.GetSummonerByRiotId("mubs", "NA1")
if err != nil {
    log.Fatal(err)
}
```

### TFT Match API

```go
// Get recent TFT match IDs
matchIDs, err := riot.GetTFTMatchIDsByPUUIDSimple(account.PUUID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d recent matches\n", len(matchIDs))

// Get detailed match data
if len(matchIDs) > 0 {
    match, err := riot.GetTFTMatchByID(matchIDs[0])
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Match: Set %d, %d participants, %.1f seconds\n",
        match.Info.TftSetNumber, 
        len(match.Info.Participants), 
        match.Info.GameLength)
}

// Get match IDs with custom parameters
startTime := int64(1640995200) // Unix timestamp
matchIDs, err = riot.GetTFTMatchIDsByPUUID(account.PUUID, 0, 10, &startTime, nil)
```

## Package Structure

- `riot.go` - Common HTTP boilerplate and account functions
- `summoner.go` - League of Legends summoner functions
- `tft.go` - Teamfight Tactics match functions  
- `types.go` - All data structures and API response types

## Testing

Run the integration tests:

```bash
# Using environment variable
RIOT_API_KEY="your-api-key" go test -v ./src/riot

# Using .env file
./test.sh

# Run benchmarks
go test -bench=. ./src/riot
```

**Note:** Tests use real API calls and require a valid API key with appropriate permissions.

## API Regions

- **Americas region** (`americas.api.riotgames.com`) - Used for account and TFT match data
- **NA1 region** (`na1.api.riotgames.com`) - Used for League of Legends summoner data

Different game modes and data types use different regional endpoints as specified by Riot's API documentation.

## Rate Limits

The client includes a 10-second timeout but does not implement rate limiting. Be mindful of Riot's rate limits when making requests in production applications.