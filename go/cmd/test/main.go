package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"lol-champ-recommender/db"
	"lol-champ-recommender/internal/api"
	"os"

	"github.com/joho/godotenv"
)

type MatchPuuids struct {
	Info struct {
		Participants []struct {
			Puuid string `json:"puuid"`
		}
	}
}

const Region = "americas"
const Server = "NA1"

func getPUUIDs(queries *db.Queries, client *api.RiotClient, ctx context.Context) (map[string]bool, error) {
	last_matches_ids, err := queries.LastMatchesFromServer(ctx, Server)
	if err != nil {
		return nil, fmt.Errorf("error getting last matches: %v", err)
	}

	allPUUIDs := map[string]bool{}

	for _, match_id := range last_matches_ids {
		matchData, err := client.MatchDetails(match_id)
		if err != nil {
			return nil, fmt.Errorf("error getting match details: %v", err)
		}

		var matchPuuids MatchPuuids
		if err := json.Unmarshal(matchData, &matchPuuids); err != nil {
			return nil, fmt.Errorf("error unmarshalling match data: %w", err)
		}

		for _, participant := range matchPuuids.Info.Participants {
			allPUUIDs[participant.Puuid] = true
		}
	}

	fmt.Println(len(allPUUIDs))

	return allPUUIDs, nil
}

const SpicasPUUID = "Uv1YAju21gW6XdmPi4X4Kcn7efgLNcwZmy8-3Uf7Ubt4zIPPHr8Kp7JX4cUqce_lPoAc0JOVK2mKIg"

type Match struct {
	Metadata struct {
		MatchID string `json:"matchId"`
	} `json:"metadata"`
	Info struct {
		EndOfGameResult    string `json:"endOfGameResult"`
		GameStartTimestamp int64  `json:"gameStartTimestamp"`
		GameVersion        string `json:"gameVersion"`
		QueueID            int    `json:"queueId"`
		PlatformID         string `json:"platformId"`
		Participants       []struct {
			ChampionName string `json:"championName"`
			ChampionID   int    `json:"championId"`
			TeamID       int    `json:"teamId"`
		} `json:"participants"`
		Teams []struct {
			TeamID int  `json:"teamId"`
			Win    bool `json:"win"`
		} `json:"teams"`
	} `json:"info"`
}

func main() {
	ctx := context.Background()

	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize database
	// dbPool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	// if err != nil {
	// 	log.Fatalf("Error creating connection pool: %v", err)
	// }
	// defer dbPool.Close()

	// queries := db.New(dbPool)

	// Initialize API client
	apiKey := os.Getenv("RIOT_API_KEY")

	client, err := api.NewRiotClient(apiKey, Region, ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize Riot API client for %s: %v", Region, err))
	}

	matchData, err := client.MatchDetails("NA1_5216346874")
	if err != nil {
		panic(fmt.Sprintf("error getting match details: %v", err))
	}

	var match Match
	if err := json.Unmarshal(matchData, &match); err != nil {
		fmt.Println("error unmarshalling match data: %w", err)
	}
	fmt.Println(match)

	// allPUUIDs, err := getPUUIDs(queries, client, ctx)
	// if err != nil {
	// 	panic(fmt.Sprintf("error getting PUUIDs: %v", err))
	// }

	// for puuid := range allPUUIDs {
	// 	hasBeenSearched, err := queries.PlayerHasBeenSearched(ctx, puuid)
	// 	if err != nil {
	// 		panic(fmt.Sprintf("error checking if player has been searched: %v", err))
	// 	}

	// 	lastSearched, err := queries.LastSearched(ctx, puuid)
	// 	if err != nil {
	// 		panic(fmt.Sprintf("error getting last searched: %v", err))
	// 	}

	// 	fmt.Println(puuid, hasBeenSearched, lastSearched)
	// }
}
