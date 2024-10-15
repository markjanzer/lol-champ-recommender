package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"lol-champ-recommender/db"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

// Copied from api_crawler/main.go
func initDatabase(ctx context.Context, db *pgx.Conn) error {
	// Read the schema file
	schemaSQL, err := os.ReadFile("db/schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Execute the schema SQL
	_, err = db.Exec(ctx, string(schemaSQL))
	if err != nil {
		return fmt.Errorf("failed to execute schema SQL: %w", err)
	}

	fmt.Println("Database schema created successfully")
	return nil
}

// Eventually want to add specific data to this
type RecommendedChamp struct {
	ChampionID     int32
	WinProbability float64
}

func contains(arr []int32, val int32) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

func sortResults(results []RecommendedChamp) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].WinProbability > results[j].WinProbability
	})
}

func RecommendChampions(ctx context.Context, queries *db.Queries, championStats ChampionDataMap, champSelect ChampSelect) ([]RecommendedChamp, error) {
	allChampIds, err := queries.AllChampionIds(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting all champion IDs: %v", err)
	}

	// Get all champions that are not an ally or an enemy
	unavailableChampIDs := append(champSelect.Allies, champSelect.Enemies...)
	unavailableChampIDs = append(unavailableChampIDs, champSelect.Bans...)

	var results []RecommendedChamp

	for _, champID := range allChampIds {
		if contains(unavailableChampIDs, champID) {
			continue
		}

		var synergies []float64
		var matchups []float64

		for _, allyID := range champSelect.Allies {
			synergy, ok := championStats[champID].Synergies[allyID]
			if !ok {
				return nil, fmt.Errorf("synergy not found for champion %d and ally %d", champID, allyID)
			}
			if synergy.Games == 0 {
				continue
			}
			synergies = append(synergies, float64(synergy.Wins)/float64(synergy.Games))
		}

		for _, enemyID := range champSelect.Enemies {
			matchup, ok := championStats[champID].Matchups[enemyID]
			if !ok {
				return nil, fmt.Errorf("matchup not found for champion %d and enemy %d", champID, enemyID)
			}
			if matchup.Games == 0 {
				continue
			}
			matchups = append(matchups, float64(matchup.Wins)/float64(matchup.Games))
		}

		winProbability := 0.0
		dataPoints := len(synergies) + len(matchups)
		if dataPoints > 0 {
			for _, synergy := range synergies {
				winProbability += synergy
			}

			for _, matchup := range matchups {
				winProbability += matchup
			}

			winProbability /= float64(dataPoints)
		} else {
			winProbability = 0.50
		}

		results = append(results, RecommendedChamp{
			ChampionID:     champID,
			WinProbability: winProbability,
		})
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

func formatAnswer(ctx context.Context, queries *db.Queries, champSelect ChampSelect, results []RecommendedChamp) error {
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
		probabilityAsPercentage := result.WinProbability * 100
		fmt.Printf("%s: %.2f%%\n", IDToName(champsToIDs, result.ChampionID), probabilityAsPercentage)
	}

	return nil
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
	// Copied from api_crawler/main.go
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()
	connString := os.Getenv("DATABASE_URL")

	// Connect to database
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	err = initDatabase(ctx, conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}

	queries := db.New(conn)

	// New code from here on
	recordWithStats, err := queries.GetLastChampionStats(ctx)
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

	r, err := RecommendChampions(ctx, queries, championStats, champSelect)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error recommending champions: %v\n", err)
		os.Exit(1)
	}

	err = formatAnswer(ctx, queries, champSelect, r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting answer: %v\n", err)
		os.Exit(1)
	}
}
