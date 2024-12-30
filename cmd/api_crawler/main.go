package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lol-champ-recommender/internal/api"
	"lol-champ-recommender/internal/crawler"
	"lol-champ-recommender/internal/database"
)

func main() {
	ctx := context.Background()

	// Initialize database
	dbConn, err := database.Initialize(ctx)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer dbConn.Close(ctx)

	// Initialize API client
	apiKey := os.Getenv("RIOT_API_KEY")
	region := "europe"

	client, err := api.NewRiotClient(apiKey, region, ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Riot API client: %v", err)
	}

	crawler := crawler.Crawler{
		Queries: dbConn.Queries,
		Client:  client,
		Ctx:     ctx,
	}

	// Create a context that we can cancel
	runCtx, cancel := context.WithCancel(crawler.Ctx)

	// Set up channel to handle shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Run the crawler in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- crawler.RunCrawler(runCtx)
		close(errChan)
	}()

	// Wait for shutdown signal or error
	select {
	case <-shutdown:
		fmt.Println("Shutdown signal received, stopping crawler...")
		cancel()
	case err := <-errChan:
		fmt.Fprintf(os.Stderr, "Crawler stopped due to error: %v\n", err)
	}

	// Wait for the crawler to finish (with a timeout)
	select {
	case <-errChan:
		fmt.Println("Crawler stopped successfully")
	case <-time.After(10 * time.Second):
		fmt.Println("Crawler did not stop in time, forcing exit")
	}
}
