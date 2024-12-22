package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"lol-champ-recommender/db"
	"lol-champ-recommender/internal/database"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Eventually want to add specific data to this
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

func contains(arr []int32, val int32) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

func sortResults(results []ChampionPerformance) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].WinProbability > results[j].WinProbability
	})
}

func RecommendChampions(ctx context.Context, queries *db.Queries, championStats ChampionDataMap, champSelect ChampSelect) ([]ChampionPerformance, error) {
	allChampIds, err := queries.AllChampionIds(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting all champion IDs: %v", err)
	}

	// Get all champions that are not an ally or an enemy
	unavailableChampIDs := append(champSelect.Allies, champSelect.Enemies...)
	unavailableChampIDs = append(unavailableChampIDs, champSelect.Bans...)

	var results []ChampionPerformance

	for _, champID := range allChampIds {
		if contains(unavailableChampIDs, champID) {
			continue
		}

		championPerformance := ChampionPerformance{
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
				winProbability = float64(synergy.Wins) / float64(synergy.Games)
			}

			championPerformance.Synergies = append(championPerformance.Synergies, ChampionInteraction{
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
				winProbability = float64(matchup.Wins) / float64(matchup.Games)
			}
			championPerformance.Matchups = append(championPerformance.Matchups, ChampionInteraction{
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

// Taken from match_data/main.go
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

// UnmarshalChampionStats converts JSON data to ChampionDataMap
func unmarshalChampionStats(data []byte) (ChampionDataMap, error) {
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

func mapChampionsToIds(ctx context.Context, queries *db.Queries) (map[string]int32, error) {
	champions, err := queries.AllChampions(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting all champions: %v", err)
	}

	result := make(map[string]int32)
	for _, champ := range champions {
		result[champ.Name] = champ.ID
	}

	return result, nil
}

type ChampSelect struct {
	Bans    []int32
	Allies  []int32
	Enemies []int32
}

func formatAnswer(ctx context.Context, queries *db.Queries, champSelect ChampSelect, results []ChampionPerformance) error {
	champsToIDs, err := mapChampionsToIds(ctx, queries)
	if err != nil {
		return fmt.Errorf("error mapping champions to IDs: %v", err)
	}

	bannedChamps := []string{}
	for _, ban := range champSelect.Bans {
		bannedChamps = append(bannedChamps, IDToName(champsToIDs, ban))
	}
	bannedChampsString := strings.Join(bannedChamps, ", ")
	fmt.Println("Bans:", bannedChampsString)

	allyChamps := []string{}
	for _, ally := range champSelect.Allies {
		allyChamps = append(allyChamps, IDToName(champsToIDs, ally))
	}
	allyChampsString := strings.Join(allyChamps, ", ")
	fmt.Println("Allies:", allyChampsString)

	enemyChamps := []string{}
	for _, enemy := range champSelect.Enemies {
		enemyChamps = append(enemyChamps, IDToName(champsToIDs, enemy))
	}
	enemyChampsString := strings.Join(enemyChamps, ", ")
	fmt.Println("Enemies:", enemyChampsString)

	fmt.Println("Recommended:")
	for _, result := range results {
		printChampionPerformance(champsToIDs, result)
	}

	return nil
}

func printChampionPerformance(champsToIDs map[string]int32, champion ChampionPerformance) {
	championName := IDToName(champsToIDs, champion.ChampionID)
	winPercentage := probabilityAsPercentage(champion.WinProbability)
	matchupsString := printChampionInteractions(champsToIDs, champion.Matchups)
	synergiesString := printChampionInteractions(champsToIDs, champion.Synergies)

	fmt.Printf("%s: %s — MATCHUPS [ %s ] — SYNERGIES [ %s ]\n",
		championName,
		winPercentage,
		matchupsString,
		synergiesString)
}

func probabilityAsPercentage(probability float64) string {
	return fmt.Sprintf("%.2f%%", probability*100)
}

func printChampionInteractions(champsToIDs map[string]int32, interactions []ChampionInteraction) string {
	var result []string
	for _, interaction := range interactions {
		championName := IDToName(champsToIDs, interaction.ChampionID)
		result = append(result, fmt.Sprintf("%s: %d/%d", championName, interaction.Wins, interaction.Games))
	}
	return strings.Join(result, ", ")
}

func IDToName(champions map[string]int32, id int32) string {
	for name, champID := range champions {
		if champID == id {
			return name
		}
	}
	return "Unknown"
}

func main() {
	ctx := context.Background()

	// Initialize database
	db, err := database.Initialize(ctx)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer db.Close(ctx)

	// New code from here on
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

	champSelect := ChampSelect{
		Bans:    []int32{1, 2, 3},
		Allies:  []int32{4, 5, 6},
		Enemies: []int32{7, 8, 9},
	}

	r, err := RecommendChampions(ctx, db.Queries, championStats, champSelect)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error recommending champions: %v\n", err)
		os.Exit(1)
	}

	err = formatAnswer(ctx, db.Queries, champSelect, r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting answer: %v\n", err)
		os.Exit(1)
	}
}
