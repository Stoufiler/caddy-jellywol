package dashboard

import (
	"os"
	"runtime"
	"time"
)

// SystemInfo holds system information
type SystemInfo struct {
	Hostname      string  `json:"hostname"`
	OS            string  `json:"os"`
	Arch          string  `json:"arch"`
	NumCPU        int     `json:"numCpu"`
	GoVersion     string  `json:"goVersion"`
	NumGoroutines int     `json:"numGoroutines"`
	MemAllocMB    float64 `json:"memAllocMB"`
	MemTotalMB    float64 `json:"memTotalMB"`
	MemSysMB      float64 `json:"memSysMB"`
	GCCount       uint32  `json:"gcCount"`
}

// GetSystemInfo returns current system information
func GetSystemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	hostname, _ := os.Hostname()

	return SystemInfo{
		Hostname:      hostname,
		OS:            runtime.GOOS,
		Arch:          runtime.GOARCH,
		NumCPU:        runtime.NumCPU(),
		GoVersion:     runtime.Version(),
		NumGoroutines: runtime.NumGoroutine(),
		MemAllocMB:    float64(m.Alloc) / 1024 / 1024,
		MemTotalMB:    float64(m.TotalAlloc) / 1024 / 1024,
		MemSysMB:      float64(m.Sys) / 1024 / 1024,
		GCCount:       m.NumGC,
	}
}

// ProcessInfo holds process information
type ProcessInfo struct {
	PID           int       `json:"pid"`
	StartTime     time.Time `json:"startTime"`
	UptimeSeconds int64     `json:"uptimeSeconds"`
}

// GetProcessInfo returns current process information
func GetProcessInfo(startTime time.Time) ProcessInfo {
	return ProcessInfo{
		PID:           os.Getpid(),
		StartTime:     startTime,
		UptimeSeconds: int64(time.Since(startTime).Seconds()),
	}
}
