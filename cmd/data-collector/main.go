package main

import (
	"fmt"
	"log"

	"lol-champ-recommender/internal/api"
)

func main() {
	fmt.Println("Starting data collection")

	apiKey := "RGAPI-7cb21c9d-ad57-41fa-8bec-df91ce7a59c2"
	region := "americas"

	// Initialize API client
	client, err := api.NewRiotClient(apiKey, region)
	if err != nil {
		log.Fatalf("Failed to initialize Riot API client: %v", err)
	}
	fmt.Println(client)

	match, err := client.GetMatchDetails("NA1_5129114460")
	if err != nil {
		log.Fatalf("Failed to get match details: %v", err)
	}

	fmt.Printf("Match: %+v\n", match)
}
