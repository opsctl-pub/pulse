package payload

type Snapshot struct {
	OS                string
	Kernel            string
	CPUUser           float64
	CPUSystem         float64
	CPUIdle           float64
	CPUIOWait         float64
	CPUSteal          float64
	CPUCount          int
	Load1             float64
	Load5             float64
	Load15            float64
	MemoryTotal       uint64
	MemoryUsed        uint64
	MemoryAvailable   uint64
	MemoryPercent     float64
	SwapTotal         uint64
	SwapUsed          uint64
	DiskTotal         uint64
	DiskUsed          uint64
	DiskPercent       float64
	DiskInodesPercent float64
	NetworkBytesSent  uint64
	NetworkBytesRecv  uint64
	UptimeSeconds     uint64
	ContainersRunning *int
	ContainersTotal   *int
}

type Payload struct {
	PulseVersion         string  `json:"pulse_version"`
	PulseIntervalSeconds int     `json:"pulse_interval_seconds"`
	OS                   string  `json:"os,omitempty"`
	Kernel               string  `json:"kernel,omitempty"`
	CPUUser              float64 `json:"cpu_user"`
	CPUSystem            float64 `json:"cpu_system"`
	CPUIdle              float64 `json:"cpu_idle"`
	CPUIOWait            float64 `json:"cpu_iowait"`
	CPUSteal             float64 `json:"cpu_steal"`
	CPUCount             int     `json:"cpu_count"`
	Load1                float64 `json:"load_1"`
	Load5                float64 `json:"load_5"`
	Load15               float64 `json:"load_15"`
	MemoryTotal          uint64  `json:"memory_total_bytes"`
	MemoryUsed           uint64  `json:"memory_used_bytes"`
	MemoryAvailable      uint64  `json:"memory_available_bytes"`
	MemoryPercent        float64 `json:"memory_percent"`
	SwapTotal            uint64  `json:"swap_total_bytes"`
	SwapUsed             uint64  `json:"swap_used_bytes"`
	DiskTotal            uint64  `json:"disk_total_bytes"`
	DiskUsed             uint64  `json:"disk_used_bytes"`
	DiskPercent          float64 `json:"disk_percent"`
	DiskInodesPercent    float64 `json:"disk_inodes_percent"`
	BytesSentTotal       uint64  `json:"bytes_sent_total"`
	BytesRecvTotal       uint64  `json:"bytes_recv_total"`
	UptimeSeconds        uint64  `json:"uptime_seconds"`
	ContainersRunning    *int    `json:"containers_running,omitempty"`
	ContainersTotal      *int    `json:"containers_total,omitempty"`
}

func Build(version string, intervalSeconds int, snapshot Snapshot) Payload {
	return Payload{
		PulseVersion:         version,
		PulseIntervalSeconds: intervalSeconds,
		OS:                   snapshot.OS,
		Kernel:               snapshot.Kernel,
		CPUUser:              snapshot.CPUUser,
		CPUSystem:            snapshot.CPUSystem,
		CPUIdle:              snapshot.CPUIdle,
		CPUIOWait:            snapshot.CPUIOWait,
		CPUSteal:             snapshot.CPUSteal,
		CPUCount:             snapshot.CPUCount,
		Load1:                snapshot.Load1,
		Load5:                snapshot.Load5,
		Load15:               snapshot.Load15,
		MemoryTotal:          snapshot.MemoryTotal,
		MemoryUsed:           snapshot.MemoryUsed,
		MemoryAvailable:      snapshot.MemoryAvailable,
		MemoryPercent:        snapshot.MemoryPercent,
		SwapTotal:            snapshot.SwapTotal,
		SwapUsed:             snapshot.SwapUsed,
		DiskTotal:            snapshot.DiskTotal,
		DiskUsed:             snapshot.DiskUsed,
		DiskPercent:          snapshot.DiskPercent,
		DiskInodesPercent:    snapshot.DiskInodesPercent,
		BytesSentTotal:       snapshot.NetworkBytesSent,
		BytesRecvTotal:       snapshot.NetworkBytesRecv,
		UptimeSeconds:        snapshot.UptimeSeconds,
		ContainersRunning:    snapshot.ContainersRunning,
		ContainersTotal:      snapshot.ContainersTotal,
	}
}
