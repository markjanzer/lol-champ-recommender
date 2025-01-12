package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"lol-champ-recommender/internal/database"
	"lol-champ-recommender/internal/recommender"
	"os"
	"sort"
	"strconv"
)

// RecommendChampions is the main function for recommending champions
func RecommendChampions(championStats recommender.ChampionDataMap, champSelect recommender.ChampSelect) ([]recommender.ChampionPerformance, error) {
	allChampIds := []int32{}
	for k := range championStats {
		allChampIds = append(allChampIds, k)
	}

	// Get all champions that are an ally, enemy, or banned
	unavailableChampIDs := append(champSelect.Allies, champSelect.Enemies...)
	unavailableChampIDs = append(unavailableChampIDs, champSelect.Bans...)

	var results []recommender.ChampionPerformance

	for _, champID := range allChampIds {
		if contains(unavailableChampIDs, champID) {
			continue
		}

		championPerformance := recommender.ChampionPerformance{
			ChampionID: champID,
		}

		for _, allyID := range champSelect.Allies {
			synergy, ok := championStats[champID].Synergies[allyID]
			if !ok {
				return nil, fmt.Errorf("synergy not found for champion %d and ally %d", champID, allyID)
			}

			var winProbability float64
			if synergy.Games == 0 {
				winProbability = 0.50
			} else {
				// Smoothing winrate
				winProbability = float64(synergy.Wins+5) / float64(synergy.Games+10)
			}

			championPerformance.Synergies = append(championPerformance.Synergies, recommender.ChampionInteraction{
				ChampionID:     allyID,
				WinProbability: winProbability,
				Wins:           synergy.Wins,
				Games:          synergy.Games,
			})
		}

		for _, enemyID := range champSelect.Enemies {
			matchup, ok := championStats[champID].Matchups[enemyID]
			if !ok {
				return nil, fmt.Errorf("matchup not found for champion %d and enemy %d", champID, enemyID)
			}
			var winProbability float64
			if matchup.Games == 0 {
				winProbability = 0.50
			} else {
				// Smoothing winrate
				winProbability = float64(matchup.Wins+5) / float64(matchup.Games+10)
			}
			championPerformance.Matchups = append(championPerformance.Matchups, recommender.ChampionInteraction{
				ChampionID:     enemyID,
				WinProbability: winProbability,
				Wins:           matchup.Wins,
				Games:          matchup.Games,
			})
		}

		winProbability := 0.0
		dataPoints := len(championPerformance.Synergies) + len(championPerformance.Matchups)
		if dataPoints > 0 {
			for _, synergy := range championPerformance.Synergies {
				winProbability += synergy.WinProbability
			}

			for _, matchup := range championPerformance.Matchups {
				winProbability += matchup.WinProbability
			}

			winProbability /= float64(dataPoints)
		} else {
			winProbability = 0.50
		}

		championPerformance.WinProbability = winProbability
		results = append(results, championPerformance)
	}

	sortResults(results)

	return results, nil
}

// UnmarshalChampionStats converts JSON data to ChampionDataMap
func unmarshalChampionStats(data []byte) (recommender.ChampionDataMap, error) {
	// Temporary map to unmarshal JSON into
	var tempMap map[string]recommender.ChampionData

	err := json.Unmarshal(data, &tempMap)
	if err != nil {
		return nil, err
	}

	// Create the final ChampionDataMap
	result := make(recommender.ChampionDataMap)

	for key, value := range tempMap {
		// Convert string key to int32
		champID, err := strconv.ParseInt(key, 10, 32)
		if err != nil {
			return nil, err
		}

		// Copy the ChampionData
		champData := recommender.ChampionData{
			Winrate:   value.Winrate,
			Matchups:  make(map[int32]recommender.WinStats),
			Synergies: make(map[int32]recommender.WinStats),
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

// Utils
func contains(arr []int32, val int32) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

func sortResults(results []recommender.ChampionPerformance) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].WinProbability > results[j].WinProbability
	})
}

func main() {
	ctx := context.Background()

	db, err := database.Initialize(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close(ctx)

	recordWithStats, err := db.Queries.GetLastChampionStats(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting last champion stats: %v\n", err)
		os.Exit(1)
	}

	championStats, err := unmarshalChampionStats(recordWithStats.Data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshalling champion stats: %v\n", err)
		os.Exit(1)
	}

	// Banned Brand
	// Allies: Caitlyn, Morgana
	// Enemies: Ashe, Lulu
	champSelect := recommender.ChampSelect{
		Bans:    []int32{63},
		Allies:  []int32{51, 25},
		Enemies: []int32{22, 117},
	}

	// Allies: Galio, Neeko
	// champSelect := ChampSelect{
	// 	Bans:    []int32{},
	// 	Allies:  []int32{3, 518},
	// 	Enemies: []int32{},
	// }

	r, err := RecommendChampions(championStats, champSelect)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error recommending champions: %v\n", err)
		os.Exit(1)
	}

	err = recommender.FormatAnswer(ctx, db.Queries, champSelect, r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting answer: %v\n", err)
		os.Exit(1)
	}
}
