package main

import (
	"context"
	"fmt"
	"log"
	"lol-champ-recommender/internal/database"
	"lol-champ-recommender/internal/recommender"
	"os"
	"sort"
)

func RecommendChampions(championStats recommender.ChampionDataMap, champSelect recommender.ChampSelect) ([]recommender.ChampionPerformance, error) {
	allChampIDs := allChampionIDs(championStats)
	unavailableChampIDs := unavailableChampionIDs(champSelect)

	var results []recommender.ChampionPerformance

	for _, champID := range allChampIDs {
		if contains(unavailableChampIDs, champID) {
			continue
		}

		performance, err := championPerformance(champID, championStats, champSelect)
		if err != nil {
			return nil, fmt.Errorf("error getting performance for champion %d: %w", champID, err)
		}

		results = append(results, performance)
	}

	sortResults(results)

	return results, nil
}

func allChampionIDs(championStats recommender.ChampionDataMap) []int32 {
	ids := make([]int32, 0, len(championStats))
	for k := range championStats {
		ids = append(ids, k)
	}
	return ids
}

func unavailableChampionIDs(champSelect recommender.ChampSelect) []int32 {
	result := append([]int32{}, champSelect.Allies...)
	result = append(result, champSelect.Enemies...)
	result = append(result, champSelect.Bans...)
	return result
}

func championPerformance(champID int32, championStats recommender.ChampionDataMap, champSelect recommender.ChampSelect) (recommender.ChampionPerformance, error) {
	performance := recommender.ChampionPerformance{
		ChampionID: champID,
	}

	synergies, err := championInteractions(champID, championStats[champID].Synergies, champSelect.Allies)
	if err != nil {
		return recommender.ChampionPerformance{}, err
	}
	performance.Synergies = synergies

	matchups, err := championInteractions(champID, championStats[champID].Matchups, champSelect.Enemies)
	if err != nil {
		return recommender.ChampionPerformance{}, err
	}
	performance.Matchups = matchups

	performance.WinProbability = calculateWinProbability(synergies, matchups)

	return performance, nil
}

func championInteractions(champID int32, stats map[int32]recommender.WinStats, championIDs []int32) ([]recommender.ChampionInteraction, error) {
	var interactions []recommender.ChampionInteraction

	for _, targetID := range championIDs {
		stat, ok := stats[targetID]
		if !ok {
			return nil, fmt.Errorf("stats not found for champion %d and target %d", champID, targetID)
		}

		interactions = append(interactions, createInteraction(targetID, stat))
	}

	return interactions, nil
}

func createInteraction(championID int32, stats recommender.WinStats) recommender.ChampionInteraction {
	var winProbability float64
	if stats.Games == 0 {
		winProbability = 0.50
	} else {
		// Smoothing winrate
		winProbability = float64(stats.Wins+5) / float64(stats.Games+10)
	}

	return recommender.ChampionInteraction{
		ChampionID:     championID,
		WinProbability: winProbability,
		Wins:           stats.Wins,
		Games:          stats.Games,
	}
}

func calculateWinProbability(synergies, matchups []recommender.ChampionInteraction) float64 {
	interactions := append(synergies, matchups...)
	if len(interactions) == 0 {
		return 0.50
	}

	total := 0.0
	for _, interaction := range interactions {
		total += interaction.WinProbability
	}

	return total / float64(len(interactions))
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

	recordWithStats, err := db.Queries.LastChampionStats(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting last champion stats: %v\n", err)
		os.Exit(1)
	}

	championStats, err := recommender.UnmarshalChampionStats(recordWithStats.Data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshalling champion stats: %v\n", err)
		os.Exit(1)
	}

	// Banned Brand
	// Allies: Caitlyn, Morgana
	// Enemies: Ashe, Lulu
	// champSelect := recommender.ChampSelect{
	// 	Bans:    []int32{63},
	// 	Allies:  []int32{51, 25},
	// 	Enemies: []int32{22, 117},
	// }

	// Allies: Galio, Neeko
	champSelect := recommender.ChampSelect{
		Bans:    []int32{},
		Allies:  []int32{3, 67},
		Enemies: []int32{},
	}

	r, err := RecommendChampions(championStats, champSelect)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error recommending champions: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(r)

	err = recommender.FormatAnswer(ctx, db.Queries, champSelect, r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting answer: %v\n", err)
		os.Exit(1)
	}
}
