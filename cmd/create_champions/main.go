package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"lol-champ-recommender/db"
	"net/http"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

const championDataURL = "https://ddragon.leagueoflegends.com/cdn/14.20.1/data/en_US/champion.json"

type ChampionData struct {
	Data map[string]Champion `json:"data"`
}

type Champion struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

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

	// Fetch the data
	resp, err := http.Get(championDataURL)
	if err != nil {
		fmt.Println("Error fetching data:", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	// Parse the JSON
	var championData ChampionData
	err = json.Unmarshal(body, &championData)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	// Process and print the champion data
	for _, champion := range championData.Data {
		apiID, err := strconv.Atoi(champion.Key)
		if err != nil {
			fmt.Println("Error converting API ID:", err)
			return
		}

		err = queries.CreateChampion(ctx, db.CreateChampionParams{
			Name:  champion.Name,
			ApiID: int32(apiID),
		})
		if err != nil {
			fmt.Println("Error creating champion:", err)
			return
		}
	}
}
