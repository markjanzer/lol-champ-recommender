package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"lol-champ-recommender/db"
	"lol-champ-recommender/internal/api"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Crawler struct {
	Queries *db.Queries
	Client  *api.RiotClient
	Ctx     context.Context
}

type Match struct {
	Metadata struct {
		MatchID string `json:"matchId"`
	} `json:"metadata"`
	Info struct {
		EndOfGameResult    string `json:"endOfGameResult"`
		GameStartTimestamp int64  `json:"gameStartTimestamp"`
		GameVersion        string `json:"gameVersion"`
		QueueID            int    `json:"queueId"`
		PlatformID         string `json:"platformId"`
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

type SeedAccount struct {
	PUUID  string `json:"puuid"`
	Server string `json:"server"`
}

func (c *Crawler) RunCrawler(runCtx context.Context) error {
	for {
		select {
		case <-runCtx.Done():
			return runCtx.Err()
		default:
			puuid, err := c.findNextPlayer()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error finding next player: %v\n", err)
			}
			fmt.Printf("Crawling player: %s\n", puuid)

			err = c.crawlPlayer(runCtx, puuid)
			if err != nil {
				if err == runCtx.Err() {
					return err
				}
				fmt.Fprintf(os.Stderr, "Error during crawl: %v\n", err)
			}
		}
	}
}

func (c *Crawler) crawlPlayer(ctx context.Context, puuid string) error {
	matchIDs, err := c.getRecentMatches(puuid)
	if err != nil {
		return err
	}

	for _, matchID := range matchIDs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = c.createMatch(matchID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating match: %v\n", err)
			}
		}
	}

	// Log the search
	err = c.Queries.LogPlayerSearch(c.Ctx, puuid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error logging player search: %v\n", err)
	}

	return nil
}

func saveMatch(queries *db.Queries, match *Match) error {
	gameStart := pgtype.Timestamp{}
	err := gameStart.Scan(time.Unix(match.Info.GameStartTimestamp/1000, 0)) // Note: Divided by 1000 to convert milliseconds to seconds
	if err != nil {
		return fmt.Errorf("error scanning game start time: %w", err)
	}
	winningTeam, err := getWinningTeam(match)
	if err != nil {
		return err
	}
	createMatchParams := db.CreateMatchParams{
		MatchID:         match.Metadata.MatchID,
		GameStart:       gameStart,
		GameVersion:     match.Info.GameVersion,
		QueueID:         int32(match.Info.QueueID),
		ServerID:        match.Info.PlatformID,
		WinningTeam:     winningTeam,
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

func getWinningTeam(match *Match) (string, error) {
	for _, team := range match.Info.Teams {
		if team.Win {
			if team.TeamID == 100 {
				return "blue", nil
			} else if team.TeamID == 200 {
				return "red", nil
			}
		}
	}
	return "", fmt.Errorf("no winning team found for match: %s, end of game result: %s", match.Metadata.MatchID, match.Info.EndOfGameResult)
}

func (c *Crawler) getRecentMatches(puuid string) ([]string, error) {
	body, err := c.Client.GetRecentMatches(puuid, 20)
	if err != nil {
		return nil, err
	}

	var matchIDs []string
	if err := json.Unmarshal(body, &matchIDs); err != nil {
		return nil, fmt.Errorf("error unmarshalling match IDs: %w", err)
	}

	return matchIDs, nil
}

func (c *Crawler) createMatch(matchID string) error {
	// Check if match already exists
	matchExists, err := c.Queries.MatchExists(c.Ctx, matchID)
	if err != nil {
		return fmt.Errorf("error checking if match exists: %w", err)
	}
	if matchExists {
		fmt.Println("Match already exists", matchID)
		return nil
	}

	matchData, err := c.Client.GetMatchDetails(matchID)
	if err != nil {
		return err
	}

	var match Match
	if err := json.Unmarshal(matchData, &match); err != nil {
		return fmt.Errorf("error unmarshalling match data: %w", err)
	}

	if match.Info.QueueID == 1700 {
		fmt.Println("Skipping ranked arena match", matchID)
		return nil
	}

	err = saveMatch(c.Queries, &match)
	if err != nil {
		return fmt.Errorf("error saving match: %w", err)
	}
	fmt.Println("Created match", matchID)

	return nil
}

func (c *Crawler) seedAccount() (SeedAccount, error) {
	data, err := os.ReadFile("config/seed_accounts.json")
	if err != nil {
		return SeedAccount{}, fmt.Errorf("error reading seed accounts: %v", err)
	}

	var seedAccounts map[string]SeedAccount
	if err := json.Unmarshal(data, &seedAccounts); err != nil {
		return SeedAccount{}, fmt.Errorf("error unmarshalling seed accounts: %v", err)
	}

	if _, ok := seedAccounts[c.Client.Region]; !ok {
		return SeedAccount{}, fmt.Errorf("no seed account found for region: %v", c.Client.Region)
	}

	return seedAccounts[c.Client.Region], nil
}

// This is a little confusing, because we are passing regions, but currently each region has one server
// and the server is what is stored in the matches table. So for a given region we will find the relevant
// server via the seed accounts, and then see if there are any matches from that server.
func (c *Crawler) findNextPlayer() (string, error) {
	seedAccount, err := c.seedAccount()
	if err != nil {
		return "", fmt.Errorf("error seeding account: %v", err)
	}
	server := seedAccount.Server

	any_matches, err := c.Queries.AnyMatchesFromServer(c.Ctx, server)
	if err != nil {
		return "", fmt.Errorf("find next player: %w", err)
	}
	if !any_matches {
		fmt.Println("No matches found for server", server)
		return seedAccount.PUUID, nil
	}

	last_matches_ids, err := c.Queries.LastMatchesFromServer(c.Ctx, server)
	if err != nil {
		return "", fmt.Errorf("error getting last matches: %v", err)
	}

	for _, match_id := range last_matches_ids {
		matchData, err := c.Client.GetMatchDetails(match_id)
		if err != nil {
			return "", err
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
			has_been_searched, err := c.Queries.PlayerHasBeenSearched(c.Ctx, puuid)
			if err != nil {
				return "", fmt.Errorf("error checking if player has been searched: %v", err)
			}
			if !has_been_searched {
				return puuid, nil
			}

			last_searched, err := c.Queries.LastSearched(c.Ctx, puuid)
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
