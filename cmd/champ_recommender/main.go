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

func RecommendChampion(ctx context.Context, queries *db.Queries, allies []int32, enemies []int32, bans []int32) {
	// Get all champions that are not an ally or an enemy
	unavailableChampIDs := append(allies, enemies...)
	unavailableChampIDs = append(unavailableChampIDs, bans...)
	champions, err := queries.GetChampionsNotIn(ctx, unavailableChampIDs)
	if err != nil {
		log.Fatalf("Failed to get champions: %v", err)
	}

	fmt.Println("Champions not in allies, enemies, or bans:")
	for _, champion := range champions {
		fmt.Printf("%s\n", champion.Name)
	}
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
	RecommendChampion(ctx, queries, []int32{1, 2, 3}, []int32{4, 5, 6}, []int32{7, 8, 9})
}
