package main

import (
	"context"
	"encoding/json"
	"log"
	"lol-champ-recommender/internal/database"
	"lol-champ-recommender/internal/recommender"
	"os"
)

type Champion struct {
	Name  string `json:"name"`
	ApiID int    `json:"api_id"`
}

func writeJSONToNext[T any](data T, filename string) error {
	const basePath = "../next/src/data/"

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(basePath+filename, jsonData, 0644)
}

func writeChampionsToNext(ctx context.Context, dbConn *database.DB) {
	dbChampions, err := dbConn.Queries.AllChampions(ctx)
	if err != nil {
		log.Println(err)
	}

	championJSONList := make([]Champion, len(dbChampions))
	for i, champion := range dbChampions {
		championJSONList[i] = Champion{
			Name:  champion.Name,
			ApiID: int(champion.ApiID),
		}
	}

	err = writeJSONToNext(championJSONList, "champions.json")
	if err != nil {
		log.Println(err)
	}
}

func writeChampionStatsToNext(ctx context.Context, dbConn *database.DB) {
	championStats, err := dbConn.Queries.GetLastChampionStats(ctx)
	if err != nil {
		log.Println(err)
	}

	championStatsData, err := recommender.UnmarshalChampionStats(championStats.Data)
	if err != nil {
		log.Println(err)
	}

	err = writeJSONToNext(championStatsData, "champion_stats.json")
	if err != nil {
		log.Println(err)
	}
}

func main() {
	ctx := context.Background()

	dbConn, err := database.Initialize(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close(ctx)

	writeChampionsToNext(ctx, dbConn)
	writeChampionStatsToNext(ctx, dbConn)
}
