package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	baseURL = "https://%s.api.riotgames.com"
)

type RiotClient struct {
	apiKey     string
	Region     string
	client     *http.Client
	limiter    *rate.Limiter
	ctx        context.Context
	mu         sync.Mutex
	retryAfter time.Time
}

func NewRiotClient(apiKey, region string, ctx context.Context) (*RiotClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	return &RiotClient{
		apiKey: apiKey,
		Region: region,
		ctx:    ctx,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
		limiter: rate.NewLimiter(rate.Every(2*time.Minute/100), 1),
	}, nil
}

func (c *RiotClient) request(url string) ([]byte, error) {
	c.mu.Lock()
	if time.Now().Before(c.retryAfter) {
		sleepDur := time.Until(c.retryAfter)
		c.mu.Unlock()
		time.Sleep(sleepDur)
	} else {
		c.mu.Unlock()
	}

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode == 429 {
		c.mu.Lock()
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				c.retryAfter = time.Now().Add(time.Duration(seconds) * time.Second)
			}
		} else {
			c.retryAfter = time.Now().Add(10 * time.Second)
		}
		c.mu.Unlock()
		return nil, fmt.Errorf("rate limited, retry after: %s", c.retryAfter)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (c *RiotClient) RecentMatches(puuid string, count int) ([]byte, error) {
	match_type := "ranked"
	url := fmt.Sprintf("%s/lol/match/v5/matches/by-puuid/%s/ids?count=%d&type=%s",
		fmt.Sprintf(baseURL, c.Region), puuid, count, match_type)

	body, err := c.request(url)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	return body, nil
}

func (c *RiotClient) MatchDetails(matchID string) ([]byte, error) {
	url := fmt.Sprintf("%s/lol/match/v5/matches/%s",
		fmt.Sprintf(baseURL, c.Region), matchID)

	body, err := c.request(url)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	return body, nil
}
