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
	// Matchups/Synergies for Blue1Champion
	var wins int
	if match.WinningTeam == "blue" {
		wins = 1
	} else {
		wins = 0
	}

	err := queries.CreateOrUpdateSynergy(ctx, db.CreateOrUpdateSynergyParams{
		Champion1ID: match.Blue1ChampionID,
		Champion2ID: match.Blue2ChampionID,
		Wins:        int32(wins),
	})
	if err != nil {
		return fmt.Errorf("failed to create or update synergy: %w", err)
	}

	err = queries.CreateOrUpdateSynergy(ctx, db.CreateOrUpdateSynergyParams{
		Champion1ID: match.Blue1ChampionID,
		Champion2ID: match.Blue3ChampionID,
		Wins:        int32(wins),
	})
	if err != nil {
		return fmt.Errorf("failed to create or update synergy: %w", err)
	}

	err = queries.CreateOrUpdateSynergy(ctx, db.CreateOrUpdateSynergyParams{
		Champion1ID: match.Blue1ChampionID,
		Champion2ID: match.Blue4ChampionID,
		Wins:        int32(wins),
	})
	if err != nil {
		return fmt.Errorf("failed to create or update synergy: %w", err)
	}

	err = queries.CreateOrUpdateSynergy(ctx, db.CreateOrUpdateSynergyParams{
		Champion1ID: match.Blue1ChampionID,
		Champion2ID: match.Blue5ChampionID,
		Wins:        int32(wins),
	})
	if err != nil {
		return fmt.Errorf("failed to create or update synergy: %w", err)
	}

	err = queries.CreateOrUpdateMatchup(ctx, db.CreateOrUpdateMatchupParams{
		Champion1ID: match.Blue1ChampionID,
		Champion2ID: match.Red1ChampionID,
		Wins:        int32(wins),
	})
	if err != nil {
		return fmt.Errorf("failed to create or update matchup: %w", err)
	}

	err = queries.CreateOrUpdateMatchup(ctx, db.CreateOrUpdateMatchupParams{
		Champion1ID: match.Blue1ChampionID,
		Champion2ID: match.Red2ChampionID,
		Wins:        int32(wins),
	})
	if err != nil {
		return fmt.Errorf("failed to create or update matchup: %w", err)
	}

	err = queries.CreateOrUpdateMatchup(ctx, db.CreateOrUpdateMatchupParams{
		Champion1ID: match.Blue1ChampionID,
		Champion2ID: match.Red3ChampionID,
		Wins:        int32(wins),
	})
	if err != nil {
		return fmt.Errorf("failed to create or update matchup: %w", err)
	}

	err = queries.CreateOrUpdateMatchup(ctx, db.CreateOrUpdateMatchupParams{
		Champion1ID: match.Blue1ChampionID,
		Champion2ID: match.Red4ChampionID,
		Wins:        int32(wins),
	})
	if err != nil {
		return fmt.Errorf("failed to create or update matchup: %w", err)
	}

	err = queries.CreateOrUpdateMatchup(ctx, db.CreateOrUpdateMatchupParams{
		Champion1ID: match.Blue1ChampionID,
		Champion2ID: match.Red5ChampionID,
		Wins:        int32(wins),
	})
	if err != nil {
		return fmt.Errorf("failed to create or update matchup: %w", err)
	}

	// Matchups/Synergies for Blue2Champion
	// Matchups/Synergies for Blue3Champion
	// Matchups/Synergies for Blue4Champion
	// Matchups/Synergies for Blue5Champion
	// Matchups/Synergies for Red1Champion
	// Matchups/Synergies for Red2Champion
	// Matchups/Synergies for Red3Champion
	// Matchups/Synergies for Red4Champion
	// Matchups/Synergies for Red5Champion
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

	all_match_ids, err := queries.AllMatchIds(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting all match ids: %v\n", err)
		os.Exit(1)
	}

	for _, id := range all_match_ids {
		match, err := queries.GetMatch(ctx, id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting match with id %d: %v\n", id, err)
			// Don't think I want to exit here
			os.Exit(1)
		}

		err = createMatchData(ctx, queries, match)
	}
}
