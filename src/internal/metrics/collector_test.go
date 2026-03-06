package metrics

import (
	"context"
	"testing"
	"time"

	gopsutilcpu "github.com/shirou/gopsutil/v4/cpu"
)

func TestSampleCPUPercentagesNormalizesDeltas(t *testing.T) {
	samples := []gopsutilcpu.TimesStat{
		{
			User:   10,
			System: 20,
			Idle:   70,
			Iowait: 0,
			Steal:  0,
		},
		{
			User:   20,
			System: 40,
			Idle:   130,
			Iowait: 5,
			Steal:  5,
		},
	}

	readCalls := 0
	readTimes := func(_ context.Context) (gopsutilcpu.TimesStat, error) {
		sample := samples[readCalls]
		readCalls++
		return sample, nil
	}

	sleepCalls := 0
	sleep := func(_ context.Context, duration time.Duration) error {
		sleepCalls++
		if duration != cpuSampleInterval {
			t.Fatalf("expected sample interval %v, got %v", cpuSampleInterval, duration)
		}
		return nil
	}

	result, err := sampleCPUPercentages(context.Background(), readTimes, sleep, cpuSampleInterval)
	if err != nil {
		t.Fatalf("sampleCPUPercentages failed: %v", err)
	}

	if readCalls != 2 {
		t.Fatalf("expected 2 CPU samples, got %d", readCalls)
	}
	if sleepCalls != 1 {
		t.Fatalf("expected 1 sample sleep, got %d", sleepCalls)
	}

	assertClose(t, result.User, 10)
	assertClose(t, result.System, 20)
	assertClose(t, result.Idle, 60)
	assertClose(t, result.IOWait, 5)
	assertClose(t, result.Steal, 5)
}

func assertClose(t *testing.T, actual float64, expected float64) {
	t.Helper()

	diff := actual - expected
	if diff < 0 {
		diff = -diff
	}
	if diff > 0.0001 {
		t.Fatalf("expected %.4f, got %.4f", expected, actual)
	}
}
