package main

import (
	"context"
	"log"
	"lol-champ-recommender/internal/champions"
	"lol-champ-recommender/internal/database"
)

func main() {
	ctx := context.Background()

	dbConn, err := database.Initialize(ctx)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer dbConn.Close(ctx)

	err = champions.UpsertChampions(ctx, dbConn.Queries)
	if err != nil {
		log.Fatalf("Error updating champions: %v", err)
	}
}
