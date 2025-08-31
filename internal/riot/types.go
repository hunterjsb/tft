package riot

// API Base URLs
const (
	RIOT_AMERICAS_URL = "https://americas.api.riotgames.com"
	RIOT_ASIA_URL     = "https://asia.api.riotgames.com"
	RIOT_EUROPE_URL   = "https://europe.api.riotgames.com"
	RIOT_SEA_URL      = "https://sea.api.riotgames.com"
)

// Regional Platform URLs (for spectator API)
const (
	RIOT_NA1_URL  = "https://na1.api.riotgames.com"
	RIOT_EUW1_URL = "https://euw1.api.riotgames.com"
	RIOT_EUNE_URL = "https://eun1.api.riotgames.com"
	RIOT_KR_URL   = "https://kr.api.riotgames.com"
	RIOT_JP1_URL  = "https://jp1.api.riotgames.com"
	RIOT_BR1_URL  = "https://br1.api.riotgames.com"
	RIOT_LAN_URL  = "https://la1.api.riotgames.com"
	RIOT_LAS_URL  = "https://la2.api.riotgames.com"
	RIOT_OC1_URL  = "https://oc1.api.riotgames.com"
	RIOT_TR1_URL  = "https://tr1.api.riotgames.com"
	RIOT_RU_URL   = "https://ru.api.riotgames.com"
	RIOT_PH2_URL  = "https://ph2.api.riotgames.com"
	RIOT_SG2_URL  = "https://sg2.api.riotgames.com"
	RIOT_TH2_URL  = "https://th2.api.riotgames.com"
	RIOT_TW2_URL  = "https://tw2.api.riotgames.com"
	RIOT_VN2_URL  = "https://vn2.api.riotgames.com"
)

// RegionMapping maps region codes to their platform URLs (for spectator API)
var RegionMapping = map[string]string{
	"NA1":  RIOT_NA1_URL,
	"EUW1": RIOT_EUW1_URL,
	"EUNE": RIOT_EUNE_URL,
	"EUN1": RIOT_EUNE_URL,
	"KR":   RIOT_KR_URL,
	"JP1":  RIOT_JP1_URL,
	"BR1":  RIOT_BR1_URL,
	"LAN":  RIOT_LAN_URL,
	"LAS":  RIOT_LAS_URL,
	"OC1":  RIOT_OC1_URL,
	"TR1":  RIOT_TR1_URL,
	"RU":   RIOT_RU_URL,
	"PH2":  RIOT_PH2_URL,
	"SG2":  RIOT_SG2_URL,
	"TH2":  RIOT_TH2_URL,
	"TW2":  RIOT_TW2_URL,
	"VN2":  RIOT_VN2_URL,
}

// RegionalRouting maps region codes to their regional routing URLs (for match history API)
var RegionalRouting = map[string]string{
	// AMERICAS regions
	"NA1": RIOT_AMERICAS_URL,
	"BR1": RIOT_AMERICAS_URL,
	"LAN": RIOT_AMERICAS_URL,
	"LAS": RIOT_AMERICAS_URL,

	// ASIA regions
	"KR":  RIOT_ASIA_URL,
	"JP1": RIOT_ASIA_URL,
	"PH2": RIOT_ASIA_URL,
	"SG2": RIOT_ASIA_URL,
	"TH2": RIOT_ASIA_URL,
	"TW2": RIOT_ASIA_URL,
	"VN2": RIOT_ASIA_URL,

	// EUROPE regions
	"EUW1": RIOT_EUROPE_URL,
	"EUNE": RIOT_EUROPE_URL,
	"EUN1": RIOT_EUROPE_URL,
	"TR1":  RIOT_EUROPE_URL,
	"RU":   RIOT_EUROPE_URL,

	// OCE - also uses AMERICAS routing
	"OC1": RIOT_AMERICAS_URL,
}

// GetRegionalURL returns the platform URL for a given region (for spectator API)
func GetRegionalURL(region string) string {
	if url, ok := RegionMapping[region]; ok {
		return url
	}
	return RIOT_NA1_URL // default fallback
}

// GetRegionalRoutingURL returns the regional routing URL for a given region (for match history API)
func GetRegionalRoutingURL(region string) string {
	if url, ok := RegionalRouting[region]; ok {
		return url
	}
	return RIOT_AMERICAS_URL // default fallback to AMERICAS
}

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

// Spectator TFT v5 Types

type CurrentGameInfo struct {
	GameID            int64                    `json:"gameId"`
	GameType          string                   `json:"gameType"`
	GameStartTime     int64                    `json:"gameStartTime"`
	MapID             int64                    `json:"mapId"`
	GameLength        int64                    `json:"gameLength"`
	PlatformID        string                   `json:"platformId"`
	GameMode          string                   `json:"gameMode"`
	BannedChampions   []BannedChampion         `json:"bannedChampions"`
	GameQueueConfigID int64                    `json:"gameQueueConfigId"`
	Observers         Observer                 `json:"observers"`
	Participants      []CurrentGameParticipant `json:"participants"`
}

type BannedChampion struct {
	PickTurn   int   `json:"pickTurn"`
	ChampionID int64 `json:"championId"`
	TeamID     int64 `json:"teamId"`
}

type Observer struct {
	EncryptionKey string `json:"encryptionKey"`
}

type CurrentGameParticipant struct {
	ChampionID             int64                     `json:"championId"`
	Perks                  Perks                     `json:"perks"`
	ProfileIconID          int64                     `json:"profileIconId"`
	TeamID                 int64                     `json:"teamId"`
	PUUID                  string                    `json:"puuid"`
	Spell1ID               int64                     `json:"spell1Id"`
	Spell2ID               int64                     `json:"spell2Id"`
	GameCustomizationItems []GameCustomizationObject `json:"gameCustomizationObjects"`
}

type Perks struct {
	PerkIds      []int64 `json:"perkIds"`
	PerkStyle    int64   `json:"perkStyle"`
	PerkSubStyle int64   `json:"perkSubStyle"`
}

type GameCustomizationObject struct {
	Category string `json:"category"`
	Content  string `json:"content"`
}
