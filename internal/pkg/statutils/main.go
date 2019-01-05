package statutils

import (
	"math"
	"time"

	"github.com/pbaettig/request0r/internal/app"
)

func CollectStats(sc chan app.WorkerStats) []app.WorkerStats {
	var stats []app.WorkerStats
	for s := range sc {
		stats = append(stats, s)
	}

	return stats
}

func MinRuntime(ss []app.WorkerStats) (idx int, min time.Duration) {
	min = time.Duration(math.MaxInt64)
	idx = 0
	for i, s := range ss {
		d := s.Runtime
		if d < min {
			min = d
			idx = i
		}
	}
	return
}

func MaxRuntime(ss []app.WorkerStats) (idx int, max time.Duration) {
	max = time.Duration(0)
	idx = 0
	for i, s := range ss {
		d := s.Runtime
		if d > max {
			max = d
			idx = i
		}
	}
	return
}

func SumRequestsPerSecond(ss []app.WorkerStats) float64 {
	sum := 0.0
	for _, s := range ss {
		sum += s.RequestsPerSecond
	}
	return sum
}
