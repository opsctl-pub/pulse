package app

import (
	"context"
	"io"
	"log"
	"testing"

	"github.com/opsctl-pub/pulse/internal/config"
	"github.com/opsctl-pub/pulse/internal/docker"
	"github.com/opsctl-pub/pulse/internal/payload"
)

type stubSender struct {
	calls int
	last  payload.Payload
}

func (s *stubSender) Send(_ context.Context, body payload.Payload) error {
	s.calls++
	s.last = body
	return nil
}

type stubMetricsCollector struct {
	snapshot payload.Snapshot
}

func (s *stubMetricsCollector) Collect(_ context.Context) (payload.Snapshot, error) {
	return s.snapshot, nil
}

type stubDockerCollector struct {
	calls  int
	counts *docker.Counts
	err    error
}

func (s *stubDockerCollector) Counts(_ context.Context) (*docker.Counts, error) {
	s.calls++
	return s.counts, s.err
}

func TestSendOnceOmitsDockerWhenDisabled(t *testing.T) {
	sender := &stubSender{}
	dockerCollector := &stubDockerCollector{
		counts: &docker.Counts{Running: 2, Total: 4},
	}

	service := NewService(
		config.Config{
			APIURL:          "http://127.0.0.1:8173",
			Token:           "token",
			IntervalSeconds: 60,
			EnableDocker:    false,
			PulseVersion:    "1.0.0",
		},
		sender,
		&stubMetricsCollector{},
		dockerCollector,
		log.New(io.Discard, "", 0),
	)

	if err := service.sendOnce(context.Background()); err != nil {
		t.Fatalf("sendOnce failed: %v", err)
	}

	if dockerCollector.calls != 0 {
		t.Fatalf("expected docker collector not to be called, got %d", dockerCollector.calls)
	}
	if sender.last.ContainersRunning != nil || sender.last.ContainersTotal != nil {
		t.Fatalf("expected container fields to be omitted")
	}
}

func TestSendOnceIncludesDockerCountsWhenEnabled(t *testing.T) {
	sender := &stubSender{}
	dockerCollector := &stubDockerCollector{
		counts: &docker.Counts{Running: 2, Total: 4},
	}

	service := NewService(
		config.Config{
			APIURL:          "http://127.0.0.1:8173",
			Token:           "token",
			IntervalSeconds: 60,
			EnableDocker:    true,
			PulseVersion:    "1.0.0",
		},
		sender,
		&stubMetricsCollector{},
		dockerCollector,
		log.New(io.Discard, "", 0),
	)

	if err := service.sendOnce(context.Background()); err != nil {
		t.Fatalf("sendOnce failed: %v", err)
	}

	if dockerCollector.calls != 1 {
		t.Fatalf("expected docker collector to be called once, got %d", dockerCollector.calls)
	}
	if sender.last.ContainersRunning == nil || *sender.last.ContainersRunning != 2 {
		t.Fatalf("unexpected running container count")
	}
	if sender.last.ContainersTotal == nil || *sender.last.ContainersTotal != 4 {
		t.Fatalf("unexpected total container count")
	}
}
