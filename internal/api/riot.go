package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

const (
	baseURL = "https://%s.api.riotgames.com"
)

type RiotClient struct {
	apiKey  string
	region  string
	client  *http.Client
	limiter *rate.Limiter
	ctx     context.Context
}

func NewRiotClient(apiKey, region string, ctx context.Context) (*RiotClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	return &RiotClient{
		apiKey: apiKey,
		region: region,
		ctx:    ctx,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
		limiter: rate.NewLimiter(rate.Every(2*time.Minute/100), 1),
	}, nil
}

func (c *RiotClient) request(url string) ([]byte, error) {
	if err := c.limiter.Wait(c.ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

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
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read error response body: %w", err)
		}

		return nil, fmt.Errorf("API request failed with status code: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return body, nil
}

func (c *RiotClient) GetRecentMatches(puuid string, count int) ([]byte, error) {
	url := fmt.Sprintf("%s/lol/match/v5/matches/by-puuid/%s/ids?count=%d",
		fmt.Sprintf(baseURL, c.region), puuid, count)

	body, err := c.request(url)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	return body, nil
}

func (c *RiotClient) GetMatchDetails(matchID string) ([]byte, error) {
	url := fmt.Sprintf("%s/lol/match/v5/matches/%s",
		fmt.Sprintf(baseURL, c.region), matchID)

	body, err := c.request(url)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	return body, nil
}
