package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"lol-champ-recommender/db"
	"lol-champ-recommender/internal/champions"
	"lol-champ-recommender/internal/database"
	"os"
)

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
	riotIds, err := queries.AllChampionRiotIds(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all champion ids: %w", err)
	}

	championStats := make(ChampionDataMap)
	for _, id := range riotIds {
		championStats[id] = ChampionData{
			Matchups:  make(map[int32]WinStats),
			Synergies: make(map[int32]WinStats),
		}
		for _, id2 := range riotIds {
			championStats[id].Matchups[id2] = WinStats{0, 0}
			championStats[id].Synergies[id2] = WinStats{0, 0}
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
	if _, exists := championStats[championID]; !exists {
		fmt.Print("Might need to run create_champions first")
		panic(fmt.Sprintf("champion %d not found in championStats", championID))
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

func championStatsToJSON(championStats ChampionDataMap) ([]byte, error) {
	// Create a map to hold the JSON-friendly structure
	jsonMap := make(map[string]interface{})

	for champID, champData := range championStats {
		champKey := fmt.Sprintf("%d", champID)
		champJSON := map[string]interface{}{
			"winrate":   champData.Winrate,
			"matchups":  champData.Matchups,
			"synergies": champData.Synergies,
		}
		jsonMap[champKey] = champJSON
	}

	// Marshal the map to JSON
	jsonData, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, fmt.Errorf("error marshalling champion stats to JSON: %w", err)
	}

	return jsonData, nil
}

func main() {
	ctx := context.Background()

	dbConn, err := database.Initialize(ctx)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer dbConn.Close(ctx)

	err = champions.UpsertChampions(ctx, dbConn.Queries)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error upserting champions: %v\n", err)
		os.Exit(1)
	}

	championStats, err := initChampionStats(ctx, dbConn.Queries)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing champion stats: %v\n", err)
		os.Exit(1)
	}

	// Limit the amount of matches we process to separate training and test matches
	var percentile int32 = 70
	lastMatchID, err := dbConn.Queries.GetMatchAtPercentileID(ctx, percentile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting match at percentile %d: %v\n", percentile, err)
		os.Exit(1)
	}

	match_ids, err := dbConn.Queries.MatchIDsUpToID(ctx, lastMatchID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting all match ids: %v\n", err)
		os.Exit(1)
	}

	for _, id := range match_ids {
		match, err := dbConn.Queries.GetMatch(ctx, id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting match with id %d: %v\n", id, err)
			os.Exit(1)
		}

		err = addMatchToChampionStats(championStats, match)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error adding match to champion stats: %v\n", err)
			os.Exit(1)
		}
	}

	json, err := championStatsToJSON(championStats)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting champion stats to JSON: %v\n", err)
		os.Exit(1)
	}

	err = dbConn.Queries.CreateChampionStats(ctx, db.CreateChampionStatsParams{
		Data:        json,
		LastMatchID: lastMatchID,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating champion stats: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Created champion stats from", len(match_ids), "matches")
}
