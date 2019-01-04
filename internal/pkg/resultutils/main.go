package resultutils

import (
	"net/url"
	"sort"
	"time"

	"github.com/pbaettig/request0r/internal/app"
)

func CollectResults(c chan app.WorkerResult) []app.WorkerResult {
	var rs []app.WorkerResult
	for r := range c {
		rs = append(rs, r)
	}
	return rs
}

func CountResponseStatusCodes(rs []app.WorkerResult) map[int]int {
	sc := make(map[int]int)
	for _, r := range rs {
		if r.Error == nil {
			sc[r.StatusCode]++
		}

	}
	return sc
}

func GetErrors(rs []app.WorkerResult) []*url.Error {
	var errors []*url.Error
	for _, r := range rs {
		if r.Error == nil {
			continue
		}
		errors = append(errors, r.Error)

	}
	return errors
}

func GetDurationPercentiles(rs []app.WorkerResult) map[float64]time.Duration {
	var durations []int
	for _, r := range rs {
		durations = append(durations, int(r.RequestDuration))
	}

	sort.Ints(durations)

	percentiles := [...]float64{0.01, 0.05, 0.10, 0.20, 0.30, 0.40, 0.50, 0.60, 0.70, 0.80, 0.90, 0.95, 0.99}

	// prepare return value
	ret := make(map[float64]time.Duration)

	for _, p := range percentiles {
		length := float64(len(durations))
		index := length * p
		ret[p] = time.Duration(durations[int(index)])

	}

	return ret
}
