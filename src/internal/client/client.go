package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/opsctl-pub/pulse/internal/payload"
)

type Sender interface {
	Send(ctx context.Context, body payload.Payload) error
}

type HTTPClient struct {
	endpoint string
	token    string
	client   *http.Client
}

func New(apiURL string, token string) *HTTPClient {
	return &HTTPClient{
		endpoint: strings.TrimRight(apiURL, "/") + "/api/v1/pulse/ingest",
		token:    token,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *HTTPClient) Send(ctx context.Context, body payload.Payload) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(encoded))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("send pulse: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("pulse ingest returned status %d", resp.StatusCode)
	}

	return nil
}
