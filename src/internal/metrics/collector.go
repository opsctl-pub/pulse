package metrics

import (
	"context"
	"fmt"
	"time"

	gopsutilcpu "github.com/shirou/gopsutil/v4/cpu"
	gopsutildisk "github.com/shirou/gopsutil/v4/disk"
	gopsutilhost "github.com/shirou/gopsutil/v4/host"
	gopsutilload "github.com/shirou/gopsutil/v4/load"
	gopsutilmem "github.com/shirou/gopsutil/v4/mem"
	gopsutilnet "github.com/shirou/gopsutil/v4/net"

	"github.com/opsctl-pub/pulse/internal/payload"
)

const cpuSampleInterval = 250 * time.Millisecond

type Collector interface {
	Collect(ctx context.Context) (payload.Snapshot, error)
}

type GopsutilCollector struct{}

func NewGopsutilCollector() *GopsutilCollector {
	return &GopsutilCollector{}
}

func (c *GopsutilCollector) Collect(ctx context.Context) (payload.Snapshot, error) {
	hostInfo, err := gopsutilhost.InfoWithContext(ctx)
	if err != nil {
		return payload.Snapshot{}, fmt.Errorf("host info: %w", err)
	}

	cpuPercentages, err := sampleCPUPercentages(
		ctx,
		readOverallCPUTimes,
		sleepWithContext,
		cpuSampleInterval,
	)
	if err != nil {
		return payload.Snapshot{}, fmt.Errorf("cpu percentages: %w", err)
	}

	cpuCount, err := gopsutilcpu.CountsWithContext(ctx, true)
	if err != nil {
		return payload.Snapshot{}, fmt.Errorf("cpu count: %w", err)
	}

	loadAverage, err := gopsutilload.AvgWithContext(ctx)
	if err != nil {
		return payload.Snapshot{}, fmt.Errorf("load average: %w", err)
	}

	virtualMemory, err := gopsutilmem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return payload.Snapshot{}, fmt.Errorf("virtual memory: %w", err)
	}

	swapMemory, err := gopsutilmem.SwapMemoryWithContext(ctx)
	if err != nil {
		return payload.Snapshot{}, fmt.Errorf("swap memory: %w", err)
	}

	diskUsage, err := gopsutildisk.UsageWithContext(ctx, "/")
	if err != nil {
		return payload.Snapshot{}, fmt.Errorf("disk usage: %w", err)
	}

	networkStats, err := gopsutilnet.IOCountersWithContext(ctx, false)
	if err != nil {
		return payload.Snapshot{}, fmt.Errorf("network io: %w", err)
	}

	var bytesSent uint64
	var bytesRecv uint64
	for _, stat := range networkStats {
		bytesSent += stat.BytesSent
		bytesRecv += stat.BytesRecv
	}

	return payload.Snapshot{
		OS:                hostInfo.OS,
		Kernel:            hostInfo.KernelVersion,
		CPUUser:           cpuPercentages.User,
		CPUSystem:         cpuPercentages.System,
		CPUIdle:           cpuPercentages.Idle,
		CPUIOWait:         cpuPercentages.IOWait,
		CPUSteal:          cpuPercentages.Steal,
		CPUCount:          cpuCount,
		Load1:             loadAverage.Load1,
		Load5:             loadAverage.Load5,
		Load15:            loadAverage.Load15,
		MemoryTotal:       virtualMemory.Total,
		MemoryUsed:        virtualMemory.Used,
		MemoryAvailable:   virtualMemory.Available,
		MemoryPercent:     virtualMemory.UsedPercent,
		SwapTotal:         swapMemory.Total,
		SwapUsed:          swapMemory.Used,
		DiskTotal:         diskUsage.Total,
		DiskUsed:          diskUsage.Used,
		DiskPercent:       diskUsage.UsedPercent,
		DiskInodesPercent: diskUsage.InodesUsedPercent,
		NetworkBytesSent:  bytesSent,
		NetworkBytesRecv:  bytesRecv,
		UptimeSeconds:     hostInfo.Uptime,
	}, nil
}

type cpuBreakdown struct {
	User   float64
	System float64
	Idle   float64
	IOWait float64
	Steal  float64
}

type cpuTimesReader func(context.Context) (gopsutilcpu.TimesStat, error)
type sleeper func(context.Context, time.Duration) error

func readOverallCPUTimes(ctx context.Context) (gopsutilcpu.TimesStat, error) {
	cpuTimes, err := gopsutilcpu.TimesWithContext(ctx, false)
	if err != nil {
		return gopsutilcpu.TimesStat{}, err
	}
	if len(cpuTimes) == 0 {
		return gopsutilcpu.TimesStat{}, fmt.Errorf("no cpu samples")
	}

	return cpuTimes[0], nil
}

func sleepWithContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func sampleCPUPercentages(
	ctx context.Context,
	readTimes cpuTimesReader,
	sleep sleeper,
	interval time.Duration,
) (cpuBreakdown, error) {
	start, err := readTimes(ctx)
	if err != nil {
		return cpuBreakdown{}, err
	}

	if err := sleep(ctx, interval); err != nil {
		return cpuBreakdown{}, err
	}

	end, err := readTimes(ctx)
	if err != nil {
		return cpuBreakdown{}, err
	}

	totalDelta := deltaValue(end.Total(), start.Total())
	if totalDelta <= 0 {
		return cpuBreakdown{}, nil
	}

	return cpuBreakdown{
		User:   percentageForDelta(end.User, start.User, totalDelta),
		System: percentageForDelta(end.System, start.System, totalDelta),
		Idle:   percentageForDelta(end.Idle, start.Idle, totalDelta),
		IOWait: percentageForDelta(end.Iowait, start.Iowait, totalDelta),
		Steal:  percentageForDelta(end.Steal, start.Steal, totalDelta),
	}, nil
}

func percentageForDelta(current float64, previous float64, totalDelta float64) float64 {
	if totalDelta <= 0 {
		return 0
	}

	return deltaValue(current, previous) / totalDelta * 100
}

func deltaValue(current float64, previous float64) float64 {
	delta := current - previous
	if delta < 0 {
		return 0
	}

	return delta
}
