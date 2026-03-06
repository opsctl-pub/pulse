package payload

import (
	"encoding/json"
	"testing"
)

func TestBuildIncludesExpectedFields(t *testing.T) {
	running := 3
	total := 5
	result := Build("1.0.0", 60, Snapshot{
		OS:                "linux",
		Kernel:            "6.8.0",
		CPUUser:           1.5,
		CPUSystem:         2.5,
		CPUIdle:           96.0,
		CPUIOWait:         0.0,
		CPUSteal:          0.0,
		CPUCount:          4,
		Load1:             0.1,
		Load5:             0.2,
		Load15:            0.3,
		MemoryTotal:       1024,
		MemoryUsed:        512,
		MemoryAvailable:   256,
		MemoryPercent:     50.0,
		SwapTotal:         128,
		SwapUsed:          32,
		DiskTotal:         2048,
		DiskUsed:          1024,
		DiskPercent:       50.0,
		DiskInodesPercent: 25.0,
		NetworkBytesSent:  111,
		NetworkBytesRecv:  222,
		UptimeSeconds:     333,
		ContainersRunning: &running,
		ContainersTotal:   &total,
	})

	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(encoded, &body); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	for _, key := range []string{
		"pulse_version",
		"pulse_interval_seconds",
		"memory_total_bytes",
		"memory_used_bytes",
		"memory_available_bytes",
		"swap_total_bytes",
		"swap_used_bytes",
		"disk_total_bytes",
		"disk_used_bytes",
		"bytes_sent_total",
		"bytes_recv_total",
		"uptime_seconds",
		"containers_running",
		"containers_total",
	} {
		if _, ok := body[key]; !ok {
			t.Fatalf("expected payload to contain key %q", key)
		}
	}

	for _, key := range []string{
		"memory_total",
		"memory_used",
		"memory_available",
		"swap_total",
		"swap_used",
		"disk_total",
		"disk_used",
	} {
		if _, ok := body[key]; ok {
			t.Fatalf("expected payload to omit legacy key %q", key)
		}
	}
}

func TestBuildOmitsContainerFieldsWhenUnavailable(t *testing.T) {
	result := Build("1.0.0", 60, Snapshot{})

	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(encoded, &body); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if _, ok := body["containers_running"]; ok {
		t.Fatalf("expected containers_running to be omitted")
	}
	if _, ok := body["containers_total"]; ok {
		t.Fatalf("expected containers_total to be omitted")
	}
}
