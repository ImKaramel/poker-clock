package clock

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Clock interface {
	UpdateStats(ctx context.Context, gameID string, players int, chips int) error
}

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

type UpdateStatsRequest struct {
	PlayersCount int `json:"players_count"`
	TotalChips   int `json:"total_chips"`
}

func (c *Client) UpdateStats(ctx context.Context, tournamentID string, req UpdateStatsRequest) error {
	url := fmt.Sprintf("%s/internal/tournaments/%s/stats", c.baseURL, tournamentID)

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("clock backend returned %d", resp.StatusCode)
	}

	return nil
}

//type HTTPClock struct {
//	client *clock.Client
//}

//func NewHTTPClock(c *clock.Client) *HTTPClock {
//	return &HTTPClock{client: c}
//}

//func (h *HTTPClock) UpdateStats(ctx context.Context, gameID string, players int, chips int) error {
//	return h.client.UpdateStats(ctx, gameID, clock.UpdateStatsRequest{
//		PlayersCount: players,
//		TotalChips:   chips,
//	})
//}
