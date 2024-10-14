package main

import (
	"context"
	"fmt"
	"log"
	"lol-champ-recommender/db"
	"os"

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

func createMatchData(ctx context.Context, queries *db.Queries, match db.Match) error {
	blueWins := match.WinningTeam == "blue"
	blueChampions := []int32{match.Blue1ChampionID, match.Blue2ChampionID, match.Blue3ChampionID, match.Blue4ChampionID, match.Blue5ChampionID}
	redChampions := []int32{match.Red1ChampionID, match.Red2ChampionID, match.Red3ChampionID, match.Red4ChampionID, match.Red5ChampionID}

	// Process all champions
	for i, champion := range append(blueChampions, redChampions...) {
		isBlue := i < 5
		err := processChampion(ctx, queries, champion, blueChampions, redChampions, isBlue, blueWins)
		if err != nil {
			return fmt.Errorf("failed to process champion %d: %w", champion, err)
		}
	}

	return nil
}

func processChampion(ctx context.Context, queries *db.Queries, championID int32, blueChampions, redChampions []int32, isBlue, blueWins bool) error {
	wins := 0
	if (isBlue && blueWins) || (!isBlue && !blueWins) {
		wins = 1
	}

	// Process synergies
	teammates := blueChampions
	if !isBlue {
		teammates = redChampions
	}
	for _, teammate := range teammates {
		if teammate != championID {
			err := queries.CreateOrUpdateSynergy(ctx, db.CreateOrUpdateSynergyParams{
				Champion1ID: championID,
				Champion2ID: teammate,
				Wins:        int32(wins),
			})
			if err != nil {
				return fmt.Errorf("failed to create or update synergy: %w", err)
			}
		}
	}

	// Process matchups
	opponents := redChampions
	if !isBlue {
		opponents = blueChampions
	}
	for _, opponent := range opponents {
		err := queries.CreateOrUpdateMatchup(ctx, db.CreateOrUpdateMatchupParams{
			Champion1ID: championID,
			Champion2ID: opponent,
			Wins:        int32(wins),
		})
		if err != nil {
			return fmt.Errorf("failed to create or update matchup: %w", err)
		}
	}

	return nil
}

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

func initChampionStats(ctx context.Context, queries *db.Queries) (ChampionDataMap, error) {
	allChampionIds, err := queries.AllChampionIds(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all champion ids: %w", err)
	}

	championStats := make(ChampionDataMap)
	for _, id := range allChampionIds {
		championStats[id] = ChampionData{
			Matchups:  make(map[int32]WinStats),
			Synergies: make(map[int32]WinStats),
		}
		for _, id2 := range allChampionIds {
			championStats[id].Matchups[id2] = WinStats{}
			championStats[id].Synergies[id2] = WinStats{}
		}
	}

	return championStats, nil
}

func addMatchToChampionStats(championStats ChampionDataMap, match db.Match) error {
	blueWins := match.WinningTeam == "blue"
	blueChampions := []int32{match.Blue1ChampionID, match.Blue2ChampionID, match.Blue3ChampionID, match.Blue4ChampionID, match.Blue5ChampionID}
	redChampions := []int32{match.Red1ChampionID, match.Red2ChampionID, match.Red3ChampionID, match.Red4ChampionID, match.Red5ChampionID}

	// Process all champions
	for i, champion := range append(blueChampions, redChampions...) {
		isBlue := i < 5
		err := addChampionToStats(championStats, champion, blueChampions, redChampions, isBlue, blueWins)
		if err != nil {
			return fmt.Errorf("failed to add champion to stats: %w", err)
		}
	}

	return nil
}

func addChampionToStats(championStats ChampionDataMap, championID int32, blueChampions, redChampions []int32, isBlue, blueWins bool) error {
	// Ensure the champion exists in the map
	if _, exists := championStats[championID]; !exists {
		championStats[championID] = ChampionData{
			Winrate:   WinStats{},
			Matchups:  make(map[int32]WinStats),
			Synergies: make(map[int32]WinStats),
		}
	}

	// Process synergies
	teammates := blueChampions
	if !isBlue {
		teammates = redChampions
	}
	won := (isBlue && blueWins) || (!isBlue && !blueWins)
	for _, teammate := range teammates {
		if teammate == championID {
			continue
		}

		if _, exists := championStats[championID].Synergies[teammate]; !exists {
			championStats[championID].Synergies[teammate] = WinStats{}
		}

		synergyStats := championStats[championID].Synergies[teammate]
		synergyStats.Games++
		if won {
			synergyStats.Wins++
		}
		championStats[championID].Synergies[teammate] = synergyStats
	}

	// Process matchups
	opponents := redChampions
	if !isBlue {
		opponents = blueChampions
	}
	for _, opponent := range opponents {
		if _, exists := championStats[championID].Matchups[opponent]; !exists {
			championStats[championID].Matchups[opponent] = WinStats{}
		}

		matchupStats := championStats[championID].Matchups[opponent]
		matchupStats.Games++
		if won {
			matchupStats.Wins++
		}
		championStats[championID].Matchups[opponent] = matchupStats
	}

	// Update the overall winrate for the champion
	cs := championStats[championID]

	cs.Winrate.Games++
	if won {
		cs.Winrate.Wins++
	}
	championStats[championID] = cs

	return nil
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

	championStats, err := initChampionStats(ctx, queries)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing champion stats: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(championStats)

	all_match_ids, err := queries.AllMatchIds(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting all match ids: %v\n", err)
		os.Exit(1)
	}

	for index, id := range all_match_ids {
		fmt.Println("Processing match", index+1, "of", len(all_match_ids))
		match, err := queries.GetMatch(ctx, id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting match with id %d: %v\n", id, err)
			// Don't think I want to exit here
			os.Exit(1)
		}

		err = addMatchToChampionStats(championStats, match)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error adding match to champion stats: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println(championStats)
}
