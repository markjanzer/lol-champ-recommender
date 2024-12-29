package main

import (
	"context"
	"log"
	"lol-champ-recommender/internal/database"
)

func main() {
	ctx := context.Background()

	dbConn, err := database.Initialize(ctx)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer dbConn.Close(ctx)

	_, err = dbConn.Conn.Exec(ctx, `
        DROP TABLE IF EXISTS matches CASCADE;
        DROP TABLE IF EXISTS champion_stats CASCADE;
				DROP TABLE IF EXISTS player_search_log CASCADE;
    `)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Tables dropped successfully")
}
