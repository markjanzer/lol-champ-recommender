package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		return nil, handleAPIErrors(resp)
	}

	var matchIDs []string
	if err := json.NewDecoder(resp.Body).Decode(&matchIDs); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return matchIDs, nil
}

func (c *RiotClient) GetMatchDetails(matchID string) ([]byte, error) {
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
		return nil, handleAPIErrors(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return body, nil
}

func handleAPIErrors(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read error response body: %w", err)
	}

	return fmt.Errorf("API request failed with status code: %d, body: %s", resp.StatusCode, string(body))
}
