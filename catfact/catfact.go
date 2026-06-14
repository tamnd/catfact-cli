// Package catfact is the library behind the catfact command line:
// the HTTP client, request shaping, and the typed data models for
// the catfact.ninja random cat facts API.
package catfact

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Host is the site this client talks to.
const Host = "catfact.ninja"

// Config holds all tunable parameters for the Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://catfact.ninja",
		UserAgent: "Mozilla/5.0 (compatible; catfact-cli/dev; +https://github.com/tamnd/catfact-cli)",
		Rate:      200 * time.Millisecond,
		Timeout:   15 * time.Second,
		Retries:   3,
	}
}

// Client talks to catfact.ninja over HTTP.
type Client struct {
	cfg  Config
	http *http.Client
	mu   sync.Mutex
	last time.Time
}

// NewClient returns a Client configured with cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// Facts fetches up to limit cat facts from /facts at the given page.
// limit<=0 defaults to 10; page<=0 defaults to 1. The API supports up to 332 total facts.
func (c *Client) Facts(ctx context.Context, limit, page int) ([]Fact, error) {
	n := limit
	if n <= 0 {
		n = 10
	}
	p := page
	if p <= 0 {
		p = 1
	}
	u := fmt.Sprintf("%s/facts?limit=%d&page=%d", c.cfg.BaseURL, n, p)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var pg factsPage
	if err := json.Unmarshal(body, &pg); err != nil {
		return nil, fmt.Errorf("decode facts: %w", err)
	}
	// offset rank by page so rows 11-20 on page 2 show correct ranks
	offset := (p - 1) * n
	items := make([]Fact, 0, len(pg.Data))
	for i, f := range pg.Data {
		items = append(items, Fact{
			Rank:   offset + i + 1,
			Fact:   f.Fact,
			Length: f.Length,
		})
	}
	return items, nil
}

// Fact fetches one random cat fact from /fact.
// The returned Fact always has Rank=1.
func (c *Client) Fact(ctx context.Context) (Fact, error) {
	u := c.cfg.BaseURL + "/fact"
	body, err := c.get(ctx, u)
	if err != nil {
		return Fact{}, err
	}
	var fr factResponse
	if err := json.Unmarshal(body, &fr); err != nil {
		return Fact{}, fmt.Errorf("decode fact: %w", err)
	}
	return Fact{Rank: 1, Fact: fr.Fact, Length: fr.Length}, nil
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	return b, err != nil, err
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	return min(time.Duration(attempt)*500*time.Millisecond, 5*time.Second)
}
