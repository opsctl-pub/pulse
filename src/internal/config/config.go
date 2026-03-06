package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	defaultIntervalSeconds = 60
)

type Config struct {
	APIURL          string
	Token           string
	IntervalSeconds int
	EnableDocker    bool
	PulseVersion    string
}

func LoadFromEnv(pulseVersion string) (Config, error) {
	apiURL := strings.TrimSpace(os.Getenv("PULSE_API_URL"))
	if apiURL == "" {
		return Config{}, fmt.Errorf("PULSE_API_URL is required")
	}
	normalizedAPIURL, err := normalizeAPIURL(apiURL)
	if err != nil {
		return Config{}, err
	}

	token := strings.TrimSpace(os.Getenv("PULSE_TOKEN"))
	if token == "" {
		return Config{}, fmt.Errorf("PULSE_TOKEN is required")
	}

	intervalSeconds := defaultIntervalSeconds
	if rawInterval := strings.TrimSpace(os.Getenv("PULSE_INTERVAL_SECONDS")); rawInterval != "" {
		parsedInterval, err := strconv.Atoi(rawInterval)
		if err != nil {
			return Config{}, fmt.Errorf("PULSE_INTERVAL_SECONDS must be an integer")
		}
		if parsedInterval <= 0 {
			return Config{}, fmt.Errorf("PULSE_INTERVAL_SECONDS must be greater than zero")
		}
		intervalSeconds = parsedInterval
	}

	enableDocker := false
	if rawEnableDocker := strings.TrimSpace(os.Getenv("PULSE_ENABLE_DOCKER")); rawEnableDocker != "" {
		parsedEnableDocker, err := strconv.ParseBool(rawEnableDocker)
		if err != nil {
			return Config{}, fmt.Errorf("PULSE_ENABLE_DOCKER must be a boolean")
		}
		enableDocker = parsedEnableDocker
	}

	return Config{
		APIURL:          normalizedAPIURL,
		Token:           token,
		IntervalSeconds: intervalSeconds,
		EnableDocker:    enableDocker,
		PulseVersion:    strings.TrimSpace(pulseVersion),
	}, nil
}

func normalizeAPIURL(rawValue string) (string, error) {
	parsedURL, err := url.Parse(rawValue)
	if err != nil {
		return "", fmt.Errorf("PULSE_API_URL is invalid")
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return "", fmt.Errorf("PULSE_API_URL must include scheme and host")
	}

	scheme := strings.ToLower(parsedURL.Scheme)
	host := strings.ToLower(parsedURL.Hostname())

	if scheme == "https" {
		return strings.TrimRight(rawValue, "/"), nil
	}

	if scheme == "http" && (host == "127.0.0.1" || host == "localhost") {
		return strings.TrimRight(rawValue, "/"), nil
	}

	return "", fmt.Errorf("PULSE_API_URL must use https unless it targets localhost or 127.0.0.1")
}
