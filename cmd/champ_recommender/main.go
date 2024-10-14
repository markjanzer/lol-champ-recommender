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

func RecommendChampions(ctx context.Context, queries *db.Queries, championStats ChampionDataMap, allies []int32, enemies []int32, bans []int32) ([]RecommendedChamp, error) {
	allChampIds, err := queries.AllChampionIds(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting all champion IDs: %v", err)
	}

	// Get all champions that are not an ally or an enemy
	unavailableChampIDs := append(allies, enemies...)
	unavailableChampIDs = append(unavailableChampIDs, bans...)

	var results []RecommendedChamp

	for _, champID := range allChampIds {
		if contains(unavailableChampIDs, champID) {
			continue
		}

		var stats []float64

		for _, allyID := range allies {
			synergy, ok := championStats[champID].Synergies[allyID]
			if ok {
				stats = append(stats, float64(synergy.Wins)/float64(synergy.Games))
			}
		}

		for _, enemyID := range enemies {
			matchup, ok := championStats[champID].Matchups[enemyID]
			if ok {
				stats = append(stats, float64(matchup.Wins)/float64(matchup.Games))
			}
		}

		winProbability := 0.0
		if len(stats) > 0 {
			for _, stat := range stats {
				winProbability += stat
			}
			winProbability /= float64(len(stats))
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

	fmt.Println(mapChampionsToIds(ctx, queries))

	r, err := RecommendChampions(ctx, queries, championStats, []int32{1, 2, 3}, []int32{4, 5, 6}, []int32{7, 8, 9})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error recommending champions: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(r)
}
