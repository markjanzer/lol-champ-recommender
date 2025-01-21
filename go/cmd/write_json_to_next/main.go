package main

import (
	"context"
	"encoding/json"
	"log"
	"lol-champ-recommender/db"
	"lol-champ-recommender/internal/database"
	"lol-champ-recommender/internal/recommender"
	"os"
)

type ChampionJSON struct {
	Name  string `json:"name"`
	ApiID int    `json:"api_id"`
}

func toChampionJSON(c db.AllChampionsRow) ChampionJSON {
	return ChampionJSON{
		Name:  c.Name,
		ApiID: int(c.ApiID),
	}
}

func main() {
	ctx := context.Background()

	dbConn, err := database.Initialize(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close(ctx)

	champions, err := dbConn.Queries.AllChampions(ctx)
	if err != nil {
		log.Fatal(err)
	}

	championJSONList := make([]ChampionJSON, len(champions))
	for i, champion := range champions {
		championJSONList[i] = toChampionJSON(champion)
	}

	jsonData, err := json.Marshal(championJSONList)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("../next/src/data/champions.json", jsonData, 0644)
	if err != nil {
		log.Fatal(err)
	}

	championStats, err := dbConn.Queries.GetLastChampionStats(ctx)
	if err != nil {
		log.Fatal(err)
	}

	championStatsData, err := recommender.UnmarshalChampionStats(championStats.Data)
	if err != nil {
		log.Fatal(err)
	}

	jsonData, err = json.Marshal(championStatsData)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("../next/src/data/champion_stats.json", jsonData, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
