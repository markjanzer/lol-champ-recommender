package recommender

import (
	"encoding/json"
	"strconv"
)

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

// UnmarshalChampionStats converts JSON data to ChampionDataMap
func UnmarshalChampionStats(data []byte) (ChampionDataMap, error) {
	// Temporary map to unmarshal JSON into
	var tempMap map[string]ChampionData

	err := json.Unmarshal(data, &tempMap)
	if err != nil {
		return nil, err
	}

	// Create the final ChampionDataMap
	result := make(ChampionDataMap)

	for key, value := range tempMap {
		// Convert string key to int32
		champID, err := strconv.ParseInt(key, 10, 32)
		if err != nil {
			return nil, err
		}

		// Copy the ChampionData
		champData := ChampionData{
			Winrate:   value.Winrate,
			Matchups:  make(map[int32]WinStats),
			Synergies: make(map[int32]WinStats),
		}

		// Convert matchups and synergies keys to int32
		for k, v := range value.Matchups {
			champData.Matchups[int32(k)] = v
		}
		for k, v := range value.Synergies {
			champData.Synergies[int32(k)] = v
		}

		result[int32(champID)] = champData
	}

	return result, nil
}
