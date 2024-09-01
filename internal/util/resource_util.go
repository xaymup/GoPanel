package util

import (
    "encoding/json"
    "net/http"
    "github.com/shirou/gopsutil/cpu"
    "github.com/shirou/gopsutil/mem"
)

// Define a struct to hold the CPU and memory usage data
type SystemStats struct {
    CPUUsage    float64 `json:"cpu_usage"`
    MemoryTotal uint64  `json:"memory_total"`
    MemoryUsed  uint64  `json:"memory_used"`
    MemoryFree  uint64  `json:"memory_free"`
}

// Handler function to get system stats
func ResourceUtilHandler(w http.ResponseWriter, r *http.Request) {
    // Get CPU usage
    cpuPercent, err := cpu.Percent(0, false)
    if err != nil {
        http.Error(w, "Error getting CPU usage", http.StatusInternalServerError)
        return
    }

    // Get memory usage
    mem, err := mem.VirtualMemory()
    if err != nil {
        http.Error(w, "Error getting memory usage", http.StatusInternalServerError)
        return
    }

    // Create a SystemStats instance
    stats := SystemStats{
        CPUUsage:    cpuPercent[0], // cpu.Percent returns a slice
        MemoryTotal: mem.Total,
        MemoryUsed:  mem.Used,
        MemoryFree:  mem.Free,
    }

    // Set response header and encode stats as JSON
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(stats); err != nil {
        http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
        return
    }
}
