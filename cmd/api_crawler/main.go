package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lol-champ-recommender/db"
	"lol-champ-recommender/internal/api"
	"lol-champ-recommender/internal/crawler"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var regions = []string{"americas", "asia", "europe", "sea"}

func runRegionCrawler(ctx context.Context, region string, queries *db.Queries, apiKey string) error {
	client, err := api.NewRiotClient(apiKey, region, ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize Riot API client for %s: %v", region, err)
	}

	crawler := crawler.Crawler{
		Queries: queries,
		Client:  client,
		Ctx:     ctx,
	}

	return crawler.RunCrawler(ctx)
}

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize database
	dbPool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error creating connection pool: %v", err)
	}
	defer dbPool.Close()

	queries := db.New(dbPool)

	// Initialize API client
	apiKey := os.Getenv("RIOT_API_KEY")

	// Create a cancellable context for all crawlers
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set up channel to handle shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Create error channel for all crawlers
	errChan := make(chan error, len(regions))

	// Start a crawler for each region
	for _, region := range regions {
		go func(r string) {
			fmt.Println(apiKey)
			errChan <- runRegionCrawler(runCtx, r, queries, apiKey)
		}(region)
	}

	// Wait for shutdown signal or first error
	select {
	case <-shutdown:
		fmt.Println("Shutdown signal received, stopping all crawlers...")
		cancel()
	case err := <-errChan:
		fmt.Fprintf(os.Stderr, "Crawler stopped due to error: %v\n", err)
		cancel()
	}

	// Wait for all crawlers to finish (with a timeout)
	finished := 0
	timeout := time.After(10 * time.Second)
	for finished < len(regions) {
		select {
		case err := <-errChan:
			if err != nil {
				fmt.Fprintf(os.Stderr, "Crawler error during shutdown: %v\n", err)
			}
			finished++
		case <-timeout:
			fmt.Println("Not all crawlers stopped in time, forcing exit")
			return
		}
	}
	fmt.Println("All crawlers stopped successfully")
}
