package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	baseURL = "https://%s.api.riotgames.com"
)

type RiotClient struct {
	apiKey string
	region string
	client *http.Client
}

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

func NewRiotClient(apiKey, region string) (*RiotClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	return &RiotClient{
		apiKey: apiKey,
		region: region,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}, nil
}

func (c *RiotClient) GetRecentMatches(puuid string, count int) ([]string, error) {
	url := fmt.Sprintf("%s/lol/match/v5/matches/by-puuid/%s/ids?count=%d",
		fmt.Sprintf(baseURL, c.region), puuid, count)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-Riot-Token", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var matchIDs []string
	if err := json.NewDecoder(resp.Body).Decode(&matchIDs); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return matchIDs, nil
}

func (c *RiotClient) GetMatchDetails(matchID string) (*Match, error) {
	url := fmt.Sprintf("%s/lol/match/v5/matches/%s",
		fmt.Sprintf(baseURL, c.region), matchID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-Riot-Token", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleAPIError(resp)
	}

	var match Match
	if err := json.NewDecoder(resp.Body).Decode(&match); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	extractMatchInfo(&match)

	return &match, nil
}

// Helper function to get champion information
func getChampionInfo(match *Match, teamID int, position int) (string, int) {
	count := 0
	for _, participant := range match.Info.Participants {
		if participant.TeamID == teamID {
			count++
			if count == position {
				return participant.ChampionName, participant.ChampionID
			}
		}
	}
	return "", 0
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

// Function to extract relevant information
func extractMatchInfo(match *Match) map[string]string {
	info := make(map[string]string)

	info["match_id"] = match.Metadata.MatchID
	info["game_start_timestamp"] = strconv.FormatInt(match.Info.GameStartTimestamp, 10)
	info["game_version"] = match.Info.GameVersion
	info["winning_team"] = getWinningTeam(match)

	for i := 1; i <= 5; i++ {
		redChampion, _ := getChampionInfo(match, 200, i)
		blueChampion, _ := getChampionInfo(match, 100, i)
		info[fmt.Sprintf("red_%d_champion", i)] = redChampion
		info[fmt.Sprintf("blue_%d_champion", i)] = blueChampion
	}

	fmt.Println("match_id:", info["match_id"])
	fmt.Println("game_start_timestamp:", info["game_start_timestamp"])
	fmt.Println("game_version:", info["game_version"])
	fmt.Println("winning_team:", info["winning_team"])
	fmt.Println("red_1_champion:", info["red_1_champion"])
	fmt.Println("red_2_champion:", info["red_2_champion"])
	fmt.Println("red_3_champion:", info["red_3_champion"])
	fmt.Println("red_4_champion:", info["red_4_champion"])
	fmt.Println("red_5_champion:", info["red_5_champion"])
	fmt.Println("blue_1_champion:", info["blue_1_champion"])
	fmt.Println("blue_2_champion:", info["blue_2_champion"])
	fmt.Println("blue_3_champion:", info["blue_3_champion"])
	fmt.Println("blue_4_champion:", info["blue_4_champion"])
	fmt.Println("blue_5_champion:", info["blue_5_champion"])

	return info
}

func handleAPIError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read error response body: %w", err)
	}

	return fmt.Errorf("API request failed with status code: %d, body: %s", resp.StatusCode, string(body))
}

// Add more methods as needed, e.g., GetSummonerByName, GetChampionMasteries, etc.
