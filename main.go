package main

import (
	"fmt"

	"github.com/hunterjsb/tft/src/dotenv"
	"github.com/hunterjsb/tft/src/riot"
)

func main() {
	dotenv.LoadDefault()

	summoner, account, err := riot.GetSummonerByRiotId("mubs", "NA1")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Found %s#%s (Level %d)\n", account.GameName, account.TagLine, summoner.SummonerLevel)
}
