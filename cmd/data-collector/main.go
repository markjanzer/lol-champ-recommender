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

func saveMatch(conn *pgx.Conn, match *Match) error {
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

	queries := db.New(conn)
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

	// Initialize API client
	apiKey := "RGAPI-7cb21c9d-ad57-41fa-8bec-df91ce7a59c2"
	region := "americas"

	client, err := api.NewRiotClient(apiKey, region)
	if err != nil {
		log.Fatalf("Failed to initialize Riot API client: %v", err)
	}
	fmt.Println(client)

	// Get match details
	matchData, err := client.GetMatchDetails("NA1_5129114460")
	if err != nil {
		log.Fatalf("Failed to get match details: %v", err)
	}

	var match Match
	if err := json.Unmarshal(matchData, &match); err != nil {
		log.Fatalf("Failed to unmarshal match data: %v", err)
	}

	err = saveMatch(conn, &match)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save match: %v", err)
	}

	// Get all matches
	queries := db.New(conn)
	matches, err := queries.AllMatches(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting matches: %v\n", err)
		os.Exit(1)
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
}
