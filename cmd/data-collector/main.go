package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
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
	fmt.Println("Creating match", matchID)
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

	return nil
}

func (c *Crawler) PrintAllMatches() error {
	matches, err := c.queries.AllMatches(c.ctx)
	if err != nil {
		return fmt.Errorf("error getting matches: %v", err)
	}

	for _, match := range matches {
		fmt.Printf("Match ID: %s\n", match.MatchID)
		fmt.Printf("Game Start: %s\n", match.GameStart.Time)
		fmt.Printf("Game Version: %s\n", match.GameVersion)
		fmt.Printf("Winning Team: %s\n", match.WinningTeam)
		fmt.Printf("Blue Team: %d, %d, %d, %d, %d\n", match.Blue1ChampionID, match.Blue2ChampionID, match.Blue3ChampionID, match.Blue4ChampionID, match.Blue5ChampionID)
		fmt.Printf("Red Team: %d, %d, %d, %d, %d\n", match.Red1ChampionID, match.Red2ChampionID, match.Red3ChampionID, match.Red4ChampionID, match.Red5ChampionID)
		fmt.Println()
	}

	return nil
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
	defer conn.Close(context.Background())

	err = initDatabase(ctx, conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}

	queries := db.New(conn)

	// Initialize API client
	apiKey := os.Getenv("RIOT_API_KEY")
	region := "americas"

	client, err := api.NewRiotClient(apiKey, region)
	if err != nil {
		log.Fatalf("Failed to initialize Riot API client: %v", err)
	}
	fmt.Println(client)

	crawler := Crawler{
		queries: queries,
		client:  client,
		ctx:     ctx,
	}

	// Get recent matches
	puuid := "b_b4LgRodsouwsgcYp-DhD5Fd0eY2VPd6A8zi1VSsFlnwitTSyWOzModIzDeFSt7_VgUEd4Pt7I0FA"

	matchIDs, err := crawler.GetRecentMatches(puuid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting recent matches: %v\n", err)
	}

	fmt.Println("Recent matches:", matchIDs)

	err = crawler.CreateMatch("NA1_5115775401")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating match: %v\n", err)
	}

	err = crawler.PrintAllMatches()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error printing all matches: %v\n", err)
	}
}
