package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lol-champ-recommender/db"
	"lol-champ-recommender/internal/api"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
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

type MatchPuuids struct {
	Info struct {
		Participants []struct {
			Puuid string `json:"puuid"`
		}
	}
}

func initDatabase(ctx context.Context, db *pgx.Conn) error {
	// Read the schema file
	schemaSQL, err := os.ReadFile("db/schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Execute the schema SQL
	_, err = db.Exec(ctx, string(schemaSQL))
	if err != nil {
		return fmt.Errorf("failed to execute schema SQL: %w", err)
	}

	fmt.Println("Database schema created successfully")
	return nil
}

func saveMatch(queries *db.Queries, match *Match) error {
	gameStart := pgtype.Timestamp{}
	err := gameStart.Scan(time.Unix(match.Info.GameStartTimestamp/1000, 0)) // Note: Divided by 1000 to convert milliseconds to seconds
	if err != nil {
		return fmt.Errorf("error scanning game start time: %w", err)
	}
	createMatchParams := db.CreateMatchParams{
		MatchID:         match.Metadata.MatchID,
		GameStart:       gameStart,
		GameVersion:     match.Info.GameVersion,
		WinningTeam:     getWinningTeam(match),
		Blue1ChampionID: getChampionId(match, 100, 1),
		Blue2ChampionID: getChampionId(match, 100, 2),
		Blue3ChampionID: getChampionId(match, 100, 3),
		Blue4ChampionID: getChampionId(match, 100, 4),
		Blue5ChampionID: getChampionId(match, 100, 5),
		Red1ChampionID:  getChampionId(match, 200, 1),
		Red2ChampionID:  getChampionId(match, 200, 2),
		Red3ChampionID:  getChampionId(match, 200, 3),
		Red4ChampionID:  getChampionId(match, 200, 4),
		Red5ChampionID:  getChampionId(match, 200, 5),
	}

	err = queries.CreateMatch(context.Background(), createMatchParams)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving match: %v\n", err)
	}

	return nil
}

// Helper function to get champion information
func getChampionId(match *Match, teamID int, position int) int32 {
	count := 0
	for _, participant := range match.Info.Participants {
		if participant.TeamID == teamID {
			count++
			if count == position {
				return int32(participant.ChampionID)
			}
		}
	}
	return 0
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

type Crawler struct {
	queries *db.Queries
	client  *api.RiotClient
	ctx     context.Context
}

// Should these be capitalized? I'm not sure if they'll be needed outside of this file
func (c *Crawler) GetRecentMatches(puuid string) ([]string, error) {
	body, err := c.client.GetRecentMatches(puuid, 20)
	if err != nil {
		return nil, fmt.Errorf("error getting recent matches: %w", err)
	}

	var matchIDs []string
	if err := json.Unmarshal(body, &matchIDs); err != nil {
		return nil, fmt.Errorf("error unmarshalling match IDs: %w", err)
	}

	return matchIDs, nil
}

func (c *Crawler) CreateMatch(matchID string) error {
	// Check if match already exists
	matchExists, err := c.queries.MatchExists(c.ctx, matchID)
	if err != nil {
		return fmt.Errorf("error checking if match exists: %w", err)
	}
	if matchExists {
		fmt.Println("Match already exists", matchID)
		return nil
	}

	matchData, err := c.client.GetMatchDetails(matchID)
	if err != nil {
		return fmt.Errorf("error getting match details: %w", err)
	}

	var match Match
	if err := json.Unmarshal(matchData, &match); err != nil {
		return fmt.Errorf("error unmarshalling match data: %w", err)
	}

	err = saveMatch(c.queries, &match)
	if err != nil {
		return fmt.Errorf("error saving match: %w", err)
	}
	fmt.Println("Created match", matchID)

	return nil
}

func (c *Crawler) FindNextPlayer() (string, error) {
	any_matches, err := c.queries.AnyMatches(c.ctx)
	if err != nil {
		return "", fmt.Errorf("error checking if there are any matches: %v", err)
	}
	if !any_matches {
		return "b_b4LgRodsouwsgcYp-DhD5Fd0eY2VPd6A8zi1VSsFlnwitTSyWOzModIzDeFSt7_VgUEd4Pt7I0FA", nil
	}

	last_matches_ids, err := c.queries.LastMatches(c.ctx)
	if err != nil {
		return "", fmt.Errorf("error getting last matches: %v", err)
	}

	for _, match_id := range last_matches_ids {
		matchData, err := c.client.GetMatchDetails(match_id)
		if err != nil {
			return "", fmt.Errorf("error getting match details: %v", err)
		}

		var matchPuuids MatchPuuids
		if err := json.Unmarshal(matchData, &matchPuuids); err != nil {
			return "", fmt.Errorf("error unmarshalling match data: %w", err)
		}

		var puuids []string
		for _, puuid := range matchPuuids.Info.Participants {
			puuids = append(puuids, puuid.Puuid)
		}

		for _, puuid := range puuids {
			has_been_searched, err := c.queries.PlayerHasBeenSearched(c.ctx, puuid)
			if err != nil {
				return "", fmt.Errorf("error checking if player has been searched: %v", err)
			}
			if !has_been_searched {
				return puuid, nil
			}

			last_searched, err := c.queries.LastSearched(c.ctx, puuid)
			if err != nil {
				return "", fmt.Errorf("error getting last searched: %v", err)
			}
			// Search puuid if it's been longer than 21 days since last search
			if time.Since(last_searched.Time) > 504*time.Hour {
				return puuid, nil
			}
		}
	}
	return "", fmt.Errorf("no new players found")
}

func (c *Crawler) crawlOnePlayer() error {
	puuid, err := c.FindNextPlayer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding next player: %v\n", err)
	}
	fmt.Println(puuid)

	matchIDs, err := c.GetRecentMatches(puuid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting recent matches: %v\n", err)
	}

	for _, matchID := range matchIDs {
		err = c.CreateMatch(matchID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating match: %v\n", err)
		}
	}

	// Log the search
	err = c.queries.LogPlayerSearch(c.ctx, db.LogPlayerSearchParams{
		PlayerID:   puuid,
		SearchTime: pgtype.Timestamp{Time: time.Now()},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error logging player search: %v\n", err)
	}

	return nil
}

func (c *Crawler) runCrawler(runCtx context.Context) error {
	for {
		select {
		case <-runCtx.Done():
			return runCtx.Err()
		default:
			if err := c.crawlOnePlayer(); err != nil {
				fmt.Fprintf(os.Stderr, "Error during crawl: %v\n", err)
			}
		}
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()
	connString := os.Getenv("DATABASE_URL")

	// Connect to database
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	err = initDatabase(ctx, conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}

	queries := db.New(conn)

	// Initialize API client
	apiKey := os.Getenv("RIOT_API_KEY")
	region := "americas"

	client, err := api.NewRiotClient(apiKey, region, ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Riot API client: %v", err)
	}

	crawler := Crawler{
		queries: queries,
		client:  client,
		ctx:     ctx,
	}

	// Create a context that we can cancel
	runCtx, cancel := context.WithCancel(crawler.ctx)

	// Set up channel to handle shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Run the crawler in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- crawler.runCrawler(runCtx)
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
	case <-time.After(30 * time.Second):
		fmt.Println("Crawler did not stop in time, forcing exit")
	}
}
