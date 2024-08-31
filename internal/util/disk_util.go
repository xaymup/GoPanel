package util

import (
	"encoding/json"
	"syscall"
	"net/http"
)

type DiskUsage struct {
	Total     uint64  `json:"total"`
	Free      uint64  `json:"free"`
	Used      uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

func getDiskUsage(path string) (DiskUsage, error) {
	var stat syscall.Statfs_t

	err := syscall.Statfs(path, &stat)
	if err != nil {
		return DiskUsage{}, err
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	used := total - free

	usage := DiskUsage{
		Total: total,
		Free:  free,
		Used:  used,
		UsedPercent: (float64(used) / float64(total)) * 100,
	}

	return usage, nil
}

func DiskUsageHandler(w http.ResponseWriter, r *http.Request) {
	usage, err := getDiskUsage("/")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}
