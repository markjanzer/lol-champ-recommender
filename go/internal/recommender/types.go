package recommender

type ChampionPerformance struct {
	ChampionID     int32
	WinProbability float64
	Synergies      []ChampionInteraction
	Matchups       []ChampionInteraction
}

type ChampionInteraction struct {
	ChampionID     int32
	WinProbability float64
	Wins           int
	Games          int
}

// Taken from create_champion_stats/main.go
type WinStats struct {
	Wins  int `json:"wins"`
	Games int `json:"games"`
}

type ChampionData struct {
	Winrate   WinStats           `json:"winrate"`
	Matchups  map[int32]WinStats `json:"matchups"`
	Synergies map[int32]WinStats `json:"synergies"`
}

type ChampionDataMap map[int32]ChampionData

// Ints are Riot IDs
type ChampSelect struct {
	Bans    []int32
	Allies  []int32
	Enemies []int32
}
