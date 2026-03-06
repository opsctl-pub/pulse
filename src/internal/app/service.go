package app

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/opsctl-pub/pulse/internal/client"
	"github.com/opsctl-pub/pulse/internal/config"
	"github.com/opsctl-pub/pulse/internal/docker"
	"github.com/opsctl-pub/pulse/internal/metrics"
	"github.com/opsctl-pub/pulse/internal/payload"
)

type DockerCollector interface {
	Counts(ctx context.Context) (*docker.Counts, error)
}

type Service struct {
	config          config.Config
	sender          client.Sender
	systemCollector metrics.Collector
	dockerCollector DockerCollector
	logger          *log.Logger
}

func NewService(
	cfg config.Config,
	sender client.Sender,
	systemCollector metrics.Collector,
	dockerCollector DockerCollector,
	logger *log.Logger,
) *Service {
	return &Service{
		config:          cfg,
		sender:          sender,
		systemCollector: systemCollector,
		dockerCollector: dockerCollector,
		logger:          logger,
	}
}

func (s *Service) Run(ctx context.Context) error {
	if err := s.sendOnce(ctx); err != nil {
		s.logger.Printf("initial pulse failed: %v", err)
	}

	ticker := time.NewTicker(time.Duration(s.config.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := s.sendOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
				s.logger.Printf("pulse send failed: %v", err)
			}
		}
	}
}

func (s *Service) sendOnce(ctx context.Context) error {
	snapshot, err := s.systemCollector.Collect(ctx)
	if err != nil {
		return err
	}

	if s.config.EnableDocker && s.dockerCollector != nil {
		counts, countErr := s.dockerCollector.Counts(ctx)
		if countErr != nil {
			s.logger.Printf("docker metrics unavailable: %v", countErr)
		} else if counts != nil {
			snapshot.ContainersRunning = &counts.Running
			snapshot.ContainersTotal = &counts.Total
		}
	}

	body := payload.Build(s.config.PulseVersion, s.config.IntervalSeconds, snapshot)
	if err := s.sender.Send(ctx, body); err != nil {
		return err
	}

	s.logger.Printf("pulse sent successfully")
	return nil
}
