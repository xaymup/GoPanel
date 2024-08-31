package util

import (
    "github.com/shirou/gopsutil/load"
    "runtime"
    "encoding/json"
    "net/http"
)

type LoadAvg struct {
    Cores  int     `json:"cores"`
    Load1  float64 `json:"load1"`
    Load5  float64 `json:"load5"`
    Load15 float64 `json:"load15"`
}

func LoadHandler(w http.ResponseWriter, r *http.Request) {
    avg, err := load.Avg()
    if err != nil {
        http.Error(w, "Could not retrieve load average", http.StatusInternalServerError)
        return
    }

	numCPU := runtime.NumCPU()

    loadAvg := LoadAvg{
		Cores:  numCPU,
        Load1:  avg.Load1,
        Load5:  avg.Load5,
        Load15: avg.Load15,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(loadAvg)

}
