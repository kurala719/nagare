package handler

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// SystemStatus represents the runtime status of the application
type SystemStatus struct {
	Uptime      int64   `json:"uptime"`
	Goroutines  int     `json:"goroutines"`
	MemoryAlloc uint64  `json:"memory_alloc"`
	MemoryTotal uint64  `json:"memory_total"`
	MemorySys   uint64  `json:"memory_sys"`
	NumGC       uint32  `json:"num_gc"`
	CpuUsage    float64 `json:"cpu_usage"`
	GoVersion   string  `json:"go_version"`
	NumCPU      int     `json:"num_cpu"`
}

var startTime = time.Now()

// GetSystemStatusCtrl returns the current system status
func GetSystemStatusCtrl(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	status := SystemStatus{
		Uptime:      int64(time.Since(startTime).Seconds()),
		Goroutines:  runtime.NumGoroutine(),
		MemoryAlloc: m.Alloc,
		MemoryTotal: m.TotalAlloc,
		MemorySys:   m.Sys,
		NumGC:       m.NumGC,
		GoVersion:   runtime.Version(),
		NumCPU:      runtime.NumCPU(),
	}

	respondSuccess(c, http.StatusOK, status)
}
