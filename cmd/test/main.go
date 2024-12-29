package main

import (
	"context"
	"fmt"
	"log"
	"lol-champ-recommender/internal/api"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("RIOT_API_KEY")
	if apiKey == "" {
		log.Fatal("RIOT_API_KEY environment variable is not set")
	}

	region := "americas"
	ctx := context.Background()

	// Remove or comment out this debug line in production
	fmt.Printf("API Key length: %d\n", len(apiKey))

	client, err := api.NewRiotClient(apiKey, region, ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Riot API client: %v", err)
	}

	body, err := client.GetMatchDetails("NA1_5115775401")
	if err != nil {
		log.Fatalf("Failed to get match details: %v", err)
	}

	fmt.Println(string(body))
}
