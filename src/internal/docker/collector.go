package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Counts struct {
	Running int
	Total   int
}

type Collector struct {
	client *http.Client
	host   string
}

type containerSummary struct {
	State string `json:"State"`
}

func NewCollector() *Collector {
	return &Collector{
		host: resolveDockerHost(),
	}
}

func (c *Collector) Counts(ctx context.Context) (*Counts, error) {
	if c.client == nil {
		client, err := newDockerHTTPClient(c.host)
		if err != nil {
			return nil, err
		}
		c.client = client
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://docker/containers/json?all=1", nil)
	if err != nil {
		return nil, fmt.Errorf("docker request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("docker list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("docker list returned status %d", resp.StatusCode)
	}

	var containers []containerSummary
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, fmt.Errorf("docker decode: %w", err)
	}

	running := 0
	for _, container := range containers {
		if container.State == "running" {
			running++
		}
	}

	return &Counts{
		Running: running,
		Total:   len(containers),
	}, nil
}

func resolveDockerHost() string {
	rawHost := strings.TrimSpace(os.Getenv("DOCKER_HOST"))
	if rawHost == "" {
		return "unix:///var/run/docker.sock"
	}

	return rawHost
}

func newDockerHTTPClient(rawHost string) (*http.Client, error) {
	parsedHost, err := url.Parse(rawHost)
	if err != nil {
		return nil, fmt.Errorf("docker host parse: %w", err)
	}

	if parsedHost.Scheme != "unix" {
		return nil, fmt.Errorf("docker host scheme %q is not supported", parsedHost.Scheme)
	}

	socketPath := parsedHost.Path
	if socketPath == "" {
		socketPath = parsedHost.Opaque
	}
	if socketPath == "" {
		return nil, fmt.Errorf("docker host socket path is required")
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, "unix", socketPath)
		},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}, nil
}
