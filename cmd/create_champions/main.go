package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"lol-champ-recommender/db"
	"lol-champ-recommender/internal/database"
	"net/http"
	"strconv"
)

const championDataURL = "https://ddragon.leagueoflegends.com/cdn/14.20.1/data/en_US/champion.json"

type ChampionData struct {
	Data map[string]Champion `json:"data"`
}

type Champion struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

func main() {
	ctx := context.Background()

	// Initialize database
	dbConn, err := database.Initialize(ctx)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer dbConn.Close(ctx)

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

		err = dbConn.Queries.CreateChampion(ctx, db.CreateChampionParams{
			Name:  champion.Name,
			ApiID: int32(apiID),
		})
		if err != nil {
			fmt.Println("Error creating champion:", err)
			return
		}
	}
}
