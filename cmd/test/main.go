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

	region := "sea"
	ctx := context.Background()

	client, err := api.NewRiotClient(apiKey, region, ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Riot API client: %v", err)
	}

	body, err := client.GetMatchDetails("VN2_696785697")
	if err != nil {
		log.Fatalf("Failed to get match details: %v", err)
	}

	fmt.Println(string(body))
}
