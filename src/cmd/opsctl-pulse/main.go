package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/opsctl-pub/pulse/internal/app"
	"github.com/opsctl-pub/pulse/internal/client"
	"github.com/opsctl-pub/pulse/internal/config"
	"github.com/opsctl-pub/pulse/internal/docker"
	"github.com/opsctl-pub/pulse/internal/metrics"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-version", "version":
			fmt.Println(version)
			return
		}
	}

	logger := log.New(os.Stdout, "opsctl-pulse: ", log.LstdFlags|log.LUTC)

	cfg, err := config.LoadFromEnv(version)
	if err != nil {
		logger.Printf("configuration error: %v", err)
		os.Exit(1)
	}

	httpClient := client.New(cfg.APIURL, cfg.Token)
	systemCollector := metrics.NewGopsutilCollector()
	var dockerCollector app.DockerCollector
	if cfg.EnableDocker {
		dockerCollector = docker.NewCollector()
	}

	service := app.NewService(
		cfg,
		httpClient,
		systemCollector,
		dockerCollector,
		logger,
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := service.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Printf("service stopped with error: %v", err)
		os.Exit(1)
	}
}
