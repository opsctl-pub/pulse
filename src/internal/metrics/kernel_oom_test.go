package metrics

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestPairedKernelOOM(t *testing.T) {
	clock := time.Date(2026, time.November, 21, 11, 19, 11, 788221000, time.UTC)
	id := strings.Repeat("ab", 32)
	summary := kernelJournalRecord{
		Message: "oom-kill:constraint=CONSTRAINT_MEMCG,nodemask=(null),oom_memcg=/docker/" + id +
			",task_memcg=/system.slice/docker-" + id + ".scope,task=worker,pid=35,uid=0",
		Transport: "kernel", BootID: strings.Repeat("a", 32),
		JournalClock: strconv.FormatInt(clock.Add(-time.Millisecond).UnixMicro(), 10), JournalMonotonic: "1000",
	}
	kill := summary
	kill.Message = "Memory cgroup out of memory: Killed process 35 (worker) total-vm:400kB, anon-rss:300kB"
	kill.JournalClock, kill.JournalMonotonic = strconv.FormatInt(clock.UnixMicro(), 10), "2000"
	paired := func(records ...kernelJournalRecord) []byte {
		t.Helper()
		var lines []string
		for _, record := range records {
			line, err := json.Marshal(record)
			if err != nil {
				t.Fatal(err)
			}
			lines = append(lines, string(line))
		}
		return []byte(strings.Join(lines, "\n"))
	}
	now := clock.Add(time.Second)
	fact := pairedKernelOOM(paired(kill, summary), now)
	if fact == nil || fact.ContainerID != id || !fact.OccurredAt.Equal(clock) {
		t.Fatalf("pair did not preserve the exact victim and UTC kill clock: %+v", fact)
	}
	if pairedKernelOOM(paired(summary), now) != nil || pairedKernelOOM(paired(kill, summary, summary), now) != nil {
		t.Fatal("incomplete or ambiguous victim evidence produced a fact")
	}
}
