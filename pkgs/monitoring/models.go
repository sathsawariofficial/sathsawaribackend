package monitoring

import "time"

type EndpointStats struct {
	Method          string
	Path            string
	TotalRequests   uint64
	SuccessCount    uint64
	FailureCount    uint64
	StatusCodes     map[int]uint64
	FailureMessages map[string]uint64
	TotalLatency    time.Duration
}

type contextKey string

type ResourceStats struct {
	CPUPercent float64

	MemPercent float32
	MemMB      float64

	FDPercent float64
	FDCount   int
}

type ResourceLogEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	CPUPercent float64   `json:"cpu_percent"`
	MemPercent float32   `json:"mem_percent"`
	MemMB      float64   `json:"mem_mb"`
	FDCount    int       `json:"fd_count"`
	FDPercent  float64   `json:"fd_percent"`
}
