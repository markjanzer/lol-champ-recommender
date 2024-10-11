package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"lol-champ-recommender/internal/api"

	"github.com/jackc/pgx/v5"
)

type Match struct {
	Metadata struct {
		MatchID string `json:"matchId"`
	} `json:"metadata"`
	Info struct {
		GameStartTimestamp int64  `json:"gameStartTimestamp"`
		GameVersion        string `json:"gameVersion"`
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
	fmt.Println("Starting data collection")

	apiKey := "RGAPI-7cb21c9d-ad57-41fa-8bec-df91ce7a59c2"
	region := "americas"

	// Initialize API client
	client, err := api.NewRiotClient(apiKey, region)
	if err != nil {
		log.Fatalf("Failed to initialize Riot API client: %v", err)
	}
	fmt.Println(client)

	matchData, err := client.GetMatchDetails("NA1_5129114460")
	if err != nil {
		log.Fatalf("Failed to get match details: %v", err)
	}

	var match Match
	if err := json.Unmarshal(matchData, &match); err != nil {
		log.Fatalf("Failed to unmarshal match data: %v", err)
	}

	matchInfo := extractMatchInfo(&match)
	// Here you would call a function to write matchInfo to the database
	fmt.Printf("Processed match: %s\n", matchInfo["match_id"])

	// Connect to database
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())
}

// Helper function to get champion information
func getChampionInfo(match *Match, teamID int, position int) (string, int) {
	count := 0
	for _, participant := range match.Info.Participants {
		if participant.TeamID == teamID {
			count++
			if count == position {
				return participant.ChampionName, participant.ChampionID
			}
		}
	}
	return "", 0
}

// Helper function to determine winning team
// Probably should have an error if neither works
func getWinningTeam(match *Match) string {
	for _, team := range match.Info.Teams {
		if team.Win {
			if team.TeamID == 100 {
				return "blue"
			} else {
				return "red"
			}
		}
	}
	return ""
}

// Function to extract relevant information
func extractMatchInfo(match *Match) map[string]string {
	info := make(map[string]string)

	info["match_id"] = match.Metadata.MatchID
	info["game_start_timestamp"] = strconv.FormatInt(match.Info.GameStartTimestamp, 10)
	info["game_version"] = match.Info.GameVersion
	info["winning_team"] = getWinningTeam(match)

	for i := 1; i <= 5; i++ {
		redChampion, _ := getChampionInfo(match, 200, i)
		blueChampion, _ := getChampionInfo(match, 100, i)
		info[fmt.Sprintf("red_%d_champion", i)] = redChampion
		info[fmt.Sprintf("blue_%d_champion", i)] = blueChampion
	}

	fmt.Println("match_id:", info["match_id"])
	fmt.Println("game_start_timestamp:", info["game_start_timestamp"])
	fmt.Println("game_version:", info["game_version"])
	fmt.Println("winning_team:", info["winning_team"])
	fmt.Println("red_1_champion:", info["red_1_champion"])
	fmt.Println("red_2_champion:", info["red_2_champion"])
	fmt.Println("red_3_champion:", info["red_3_champion"])
	fmt.Println("red_4_champion:", info["red_4_champion"])
	fmt.Println("red_5_champion:", info["red_5_champion"])
	fmt.Println("blue_1_champion:", info["blue_1_champion"])
	fmt.Println("blue_2_champion:", info["blue_2_champion"])
	fmt.Println("blue_3_champion:", info["blue_3_champion"])
	fmt.Println("blue_4_champion:", info["blue_4_champion"])
	fmt.Println("blue_5_champion:", info["blue_5_champion"])

	return info
}
