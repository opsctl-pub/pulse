package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoadFromEnvSuccess(t *testing.T) {
	t.Setenv("PULSE_API_URL", "http://127.0.0.1:8173/")
	t.Setenv("PULSE_TOKEN", "token-value")
	t.Setenv("PULSE_INTERVAL_SECONDS", "90")
	t.Setenv("PULSE_ENABLE_DOCKER", "true")

	cfg, err := LoadFromEnv("1.2.3")
	if err != nil {
		t.Fatalf("LoadFromEnv returned error: %v", err)
	}

	if cfg.APIURL != "http://127.0.0.1:8173" {
		t.Fatalf("unexpected APIURL: %s", cfg.APIURL)
	}
	if cfg.Token != "token-value" {
		t.Fatalf("unexpected token")
	}
	if cfg.IntervalSeconds != 90 {
		t.Fatalf("unexpected interval: %d", cfg.IntervalSeconds)
	}
	if !cfg.EnableDocker {
		t.Fatalf("expected EnableDocker to be true")
	}
	if cfg.PulseVersion != "1.2.3" {
		t.Fatalf("unexpected version: %s", cfg.PulseVersion)
	}
}

func TestLoadFromEnvDefaultsDockerAndInterval(t *testing.T) {
	t.Setenv("PULSE_API_URL", "http://127.0.0.1:8173")
	t.Setenv("PULSE_TOKEN", "token-value")
	os.Unsetenv("PULSE_INTERVAL_SECONDS")
	os.Unsetenv("PULSE_ENABLE_DOCKER")

	cfg, err := LoadFromEnv("dev")
	if err != nil {
		t.Fatalf("LoadFromEnv returned error: %v", err)
	}

	if cfg.IntervalSeconds != defaultIntervalSeconds {
		t.Fatalf("expected default interval %d, got %d", defaultIntervalSeconds, cfg.IntervalSeconds)
	}
	if cfg.EnableDocker {
		t.Fatalf("expected EnableDocker to be false by default")
	}
}

func TestLoadFromEnvRejectsInvalidValues(t *testing.T) {
	t.Setenv("PULSE_API_URL", "http://127.0.0.1:8173")
	t.Setenv("PULSE_TOKEN", "token-value")
	t.Setenv("PULSE_INTERVAL_SECONDS", "0")

	if _, err := LoadFromEnv("dev"); err == nil {
		t.Fatalf("expected interval validation error")
	}

	t.Setenv("PULSE_INTERVAL_SECONDS", "60")
	t.Setenv("PULSE_ENABLE_DOCKER", "not-bool")
	if _, err := LoadFromEnv("dev"); err == nil {
		t.Fatalf("expected docker boolean validation error")
	}
}

func TestLoadFromEnvAcceptsHTTPSURL(t *testing.T) {
	t.Setenv("PULSE_API_URL", "https://opsctl.example.com/")
	t.Setenv("PULSE_TOKEN", "token-value")

	cfg, err := LoadFromEnv("dev")
	if err != nil {
		t.Fatalf("LoadFromEnv returned error: %v", err)
	}

	if cfg.APIURL != "https://opsctl.example.com" {
		t.Fatalf("unexpected APIURL: %s", cfg.APIURL)
	}
}

func TestLoadFromEnvAcceptsLocalHTTPURLs(t *testing.T) {
	for _, rawURL := range []string{
		"http://127.0.0.1:8173/",
		"http://localhost:8173/",
	} {
		t.Run(rawURL, func(t *testing.T) {
			t.Setenv("PULSE_API_URL", rawURL)
			t.Setenv("PULSE_TOKEN", "token-value")

			cfg, err := LoadFromEnv("dev")
			if err != nil {
				t.Fatalf("LoadFromEnv returned error: %v", err)
			}
			if strings.HasSuffix(cfg.APIURL, "/") {
				t.Fatalf("expected normalized URL without trailing slash, got %s", cfg.APIURL)
			}
		})
	}
}

func TestLoadFromEnvRejectsInsecureRemoteHTTPURL(t *testing.T) {
	t.Setenv("PULSE_API_URL", "http://opsctl.example.com:8173")
	t.Setenv("PULSE_TOKEN", "token-value")

	_, err := LoadFromEnv("dev")
	if err == nil {
		t.Fatalf("expected insecure URL validation error")
	}
	if !strings.Contains(err.Error(), "must use https") {
		t.Fatalf("unexpected error: %v", err)
	}
}
