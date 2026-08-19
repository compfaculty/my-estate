package httpx

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	HTTP       *http.Client
	UserAgent  string
	MaxRetries int
}

func New() *Client {
	return &Client{
		HTTP: &http.Client{
			Timeout: 30 * time.Second,
		},
		UserAgent:  "EstateTerminal/0.2 (research; local lab)",
		MaxRetries: 3,
	}
}

func (c *Client) Get(ctx context.Context, url string) ([]byte, int, error) {
	var lastErr error
	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, 0, ctx.Err()
			case <-time.After(time.Duration(attempt) * 400 * time.Millisecond):
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, 0, err
		}
		req.Header.Set("User-Agent", c.UserAgent)
		req.Header.Set("Accept", "application/json, text/html;q=0.9, */*;q=0.8")
		req.Header.Set("Accept-Language", "uk-UA,uk;q=0.9,en;q=0.8")
		resp, err := c.HTTP.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("http %d", resp.StatusCode)
			continue
		}
		return body, resp.StatusCode, nil
	}
	return nil, 0, lastErr
}
