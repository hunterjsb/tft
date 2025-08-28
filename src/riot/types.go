package riot

// API Base URLs
const (
	RIOT_AMERICAS_URL = "https://americas.api.riotgames.com"
	RIOT_NA1_URL      = "https://na1.api.riotgames.com"
)

// Account and Summoner Types

type Account struct {
	PUUID    string `json:"puuid"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
}

type Summoner struct {
	ID            string `json:"id"`
	AccountID     string `json:"accountId"`
	PUUID         string `json:"puuid"`
	Name          string `json:"name"`
	ProfileIconID int    `json:"profileIconId"`
	RevisionDate  int64  `json:"revisionDate"`
	SummonerLevel int    `json:"summonerLevel"`
}

// TFT Match API Types

type MatchDto struct {
	Metadata MetadataDto `json:"metadata"`
	Info     InfoDto     `json:"info"`
}

type MetadataDto struct {
	DataVersion  string   `json:"data_version"`
	MatchID      string   `json:"match_id"`
	Participants []string `json:"participants"`
}

type InfoDto struct {
	EndOfGameResult   string           `json:"endOfGameResult"`
	GameCreation      int64            `json:"gameCreation"`
	GameID            int64            `json:"gameId"`
	GameDatetime      int64            `json:"game_datetime"`
	GameLength        float64          `json:"game_length"`
	GameVersion       string           `json:"game_version"`
	GameVariation     string           `json:"game_variation"` // Deprecated
	MapID             int              `json:"mapId"`
	Participants      []ParticipantDto `json:"participants"`
	QueueID           int              `json:"queue_id"`
	QueueIDDeprecated int              `json:"queueId"` // Deprecated
	TftGameType       string           `json:"tft_game_type"`
	TftSetCoreName    string           `json:"tft_set_core_name"`
	TftSetNumber      int              `json:"tft_set_number"`
}

type ParticipantDto struct {
	Companion            CompanionDto `json:"companion"`
	GoldLeft             int          `json:"gold_left"`
	LastRound            int          `json:"last_round"`
	Level                int          `json:"level"`
	Placement            int          `json:"placement"`
	PlayersEliminated    int          `json:"players_eliminated"`
	PUUID                string       `json:"puuid"`
	RiotIDGameName       string       `json:"riotIdGameName"`
	RiotIDTagline        string       `json:"riotIdTagline"`
	TimeEliminated       float64      `json:"time_eliminated"`
	TotalDamageToPlayers int          `json:"total_damage_to_players"`
	Traits               []TraitDto   `json:"traits"`
	Units                []UnitDto    `json:"units"`
	Win                  bool         `json:"win"`
}

type CompanionDto struct {
	ContentID string `json:"content_ID"`
	ItemID    int    `json:"item_ID"`
	SkinID    int    `json:"skin_ID"`
	Species   string `json:"species"`
}

type TraitDto struct {
	Name        string `json:"name"`
	NumUnits    int    `json:"num_units"`
	Style       int    `json:"style"`
	TierCurrent int    `json:"tier_current"`
	TierTotal   int    `json:"tier_total"`
}

type UnitDto struct {
	Items       []int    `json:"items"`
	CharacterID string   `json:"character_id"`
	ItemNames   []string `json:"itemNames"`
	Chosen      string   `json:"chosen"`
	Name        string   `json:"name"`
	Rarity      int      `json:"rarity"`
	Tier        int      `json:"tier"`
}
