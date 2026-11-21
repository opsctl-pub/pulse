package metrics

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/opsctl-pub/pulse/internal/payload"
)

const (
	kernelOOMWindow = 10 * time.Minute
	journalRecords  = 32
	journalBytes    = 16 * 1024
)

var (
	victimSummary = regexp.MustCompile(`^oom-kill:constraint=[A-Z_]+,.*,(?:global_oom|oom_memcg=[^,]+),task_memcg=([^,]+),task=.*?,pid=([1-9][0-9]*),uid=[0-9]+$`)
	killedVictim  = regexp.MustCompile(`^(?:Memory cgroup out of memory:|Out of memory:) Killed process ([1-9][0-9]*) \(.*\) total-vm:[0-9]+kB,`)
	dockerCgroup  = regexp.MustCompile(`^/(?:[^/]+/)*(?:docker/([0-9a-f]{64})|docker-([0-9a-f]{64})\.scope)$`)
	journalBootID = regexp.MustCompile(`^[0-9a-f]{32}$`)
)

type kernelJournalRecord struct {
	Message          string `json:"MESSAGE"`
	Transport        string `json:"_TRANSPORT"`
	BootID           string `json:"_BOOT_ID"`
	JournalClock     string `json:"__REALTIME_TIMESTAMP"`
	JournalMonotonic string `json:"__MONOTONIC_TIMESTAMP"`
	SourceClock      string `json:"_SOURCE_REALTIME_TIMESTAMP"`
	SourceMonotonic  string `json:"_SOURCE_MONOTONIC_TIMESTAMP"`
}

type kernelVictim struct {
	boot, pid, container       string
	monotonic, sourceMonotonic int64
	occurredAt                 time.Time
}

func collectKernelOOM(ctx context.Context) *payload.KernelOOM {
	if runtime.GOOS != "linux" {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "journalctl", "--dmesg", "--boot=0", "--since=10 minutes ago",
		"--lines=32", "--case-sensitive=yes", "--no-pager", "--quiet", "--output=json",
		"--output-fields=MESSAGE,_TRANSPORT,_BOOT_ID,_SOURCE_REALTIME_TIMESTAMP,_SOURCE_MONOTONIC_TIMESTAMP",
		"--grep=^(oom-kill:|Memory cgroup out of memory: Killed process |Out of memory: Killed process )")
	stdout, err := cmd.StdoutPipe()
	if err != nil || cmd.Start() != nil {
		return nil
	}
	raw, err := io.ReadAll(io.LimitReader(stdout, journalBytes+1))
	if err != nil || len(raw) > journalBytes {
		cancel()
	}
	if waitErr := cmd.Wait(); waitErr != nil || ctx.Err() != nil || err != nil || len(raw) > journalBytes {
		return nil
	}
	return pairedKernelOOM(raw, time.Now())
}

func pairedKernelOOM(raw []byte, now time.Time) *payload.KernelOOM {
	lines := bytes.Split(bytes.TrimSpace(raw), []byte("\n"))
	if len(lines) > journalRecords {
		return nil
	}
	var summaries, kills []kernelVictim
	for _, line := range lines {
		var record kernelJournalRecord
		if json.Unmarshal(line, &record) != nil {
			return nil
		}
		if record.Transport != "kernel" || !journalBootID.MatchString(record.BootID) {
			continue
		}
		mono, err := strconv.ParseInt(record.JournalMonotonic, 10, 64)
		if err != nil || mono <= 0 {
			continue
		}
		sourceMono, _ := strconv.ParseInt(record.SourceMonotonic, 10, 64)
		clock := record.SourceClock
		if clock == "" {
			clock = record.JournalClock
		}
		unixMicros, err := strconv.ParseInt(clock, 10, 64)
		if err != nil || unixMicros <= 0 {
			continue
		}
		occurredAt := time.UnixMicro(unixMicros).UTC()
		if occurredAt.Before(now.Add(-kernelOOMWindow)) || occurredAt.After(now) {
			continue
		}
		victim := kernelVictim{boot: record.BootID, monotonic: mono, sourceMonotonic: sourceMono, occurredAt: occurredAt}
		if fields := victimSummary.FindStringSubmatch(record.Message); fields != nil {
			victim.pid = fields[2]
			if strings.Count(record.Message, ",task_memcg=") == 1 {
				if ids := dockerCgroup.FindStringSubmatch(fields[1]); ids != nil {
					victim.container = ids[1] + ids[2]
				}
			}
			summaries = append(summaries, victim)
		} else if fields := killedVictim.FindStringSubmatch(record.Message); fields != nil {
			victim.pid = fields[1]
			kills = append(kills, victim)
		}
	}

	var latest *payload.KernelOOM
	var latestMonotonic int64
	for _, kill := range kills {
		var summary kernelVictim
		matches := 0
		for _, candidate := range summaries {
			if sameKernelVictim(candidate, kill) {
				summary = candidate
				matches++
			}
		}
		if matches != 1 || summary.container == "" {
			continue
		}
		matches = 0
		for _, candidate := range kills {
			if sameKernelVictim(summary, candidate) {
				matches++
			}
		}
		if matches == 1 && kill.monotonic > latestMonotonic {
			latestMonotonic = kill.monotonic
			latest = &payload.KernelOOM{ContainerID: summary.container, OccurredAt: kill.occurredAt}
		}
	}
	return latest
}

func sameKernelVictim(summary, kill kernelVictim) bool {
	if summary.boot != kill.boot || summary.pid != kill.pid {
		return false
	}
	start, end := summary.monotonic, kill.monotonic
	if summary.sourceMonotonic > 0 && kill.sourceMonotonic > 0 {
		start, end = summary.sourceMonotonic, kill.sourceMonotonic
	}
	return end > start && end-start <= time.Second.Microseconds()
}
