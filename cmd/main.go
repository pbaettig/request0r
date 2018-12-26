package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/pbaettig/randurl"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	numRequests       = 10
	concurrencyFactor = 5
)

func calculateDurationPercentiles(d []time.Duration) map[float64]time.Duration {
	// convert duration to int to facilitate sorting
	ints := make([]int, len(d))
	for i, v := range d {
		ints[i] = int(v)
	}
	sort.Ints(ints)

	percentiles := [...]float64{0.01, 0.05, 0.10, 0.20, 0.30, 0.40, 0.50, 0.60, 0.70, 0.80, 0.90, 0.99}

	// prepare return value
	ret := make(map[float64]time.Duration)

	for _, p := range percentiles {
		length := float64(len(d))
		index := length * p
		ret[p] = time.Duration(ints[int(index)])

	}

	return ret
}

func requestWorker(wg *sync.WaitGroup, in <-chan WorkItem, out chan<- Result) {
	defer wg.Done()
	for wi := range in {
		var result Result
		result.ID = wi.ID
		result.URL = wi.URL

		start := time.Now()
		resp, err := http.Get(wi.URL)
		if err != nil {
			result.Error = err
		} else {
			defer resp.Body.Close()
			result.StatusCode = resp.StatusCode
			result.Header = resp.Header
			result.ContentLength = resp.ContentLength
		}
		result.RequestDuration = time.Now().Sub(start)

		out <- result

	}
}

func main() {
	urlSpecs := map[string]randurl.URLSpec{
		"status_test": randurl.URLSpec{
			Scheme: "http",
			Host:   "httpbin.org",
			Components: []randurl.PathComponent{
				randurl.StringComponent("status"),
				randurl.IntegerComponent{Min: 100, Max: 511},
			},
		},
		"delay_test": randurl.URLSpec{
			Scheme: "http",
			Host:   "httpbin.org",
			Components: []randurl.PathComponent{
				randurl.StringComponent("delay"),
				randurl.IntegerComponent{Min: 1, Max: 2},
			},
		},
	}

	wg := new(sync.WaitGroup)
	in := make(chan WorkItem, len(urlSpecs)*numRequests)
	out := make(chan Result, concurrencyFactor)

	for i := 0; i < concurrencyFactor; i++ {
		wg.Add(1)
		go requestWorker(wg, in, out)
	}

	for id, spec := range urlSpecs {
		for i := 0; i < numRequests; i++ {
			in <- WorkItem{id, spec.String()}
		}
	}
	close(in)

	// Collect results
	results := make(map[string][]Result)
	for i := 0; i < numRequests*len(urlSpecs); i++ {
		r := <-out

		results[r.ID] = append(results[r.ID], r)
		fmt.Printf("%s: %s [%d] [%d ms]\n", r.ID, r.URL, r.StatusCode, r.RequestDuration/time.Millisecond)
	}
	wg.Wait()
	close(out)

	// calculate request time percentiles
	for id, rs := range results {
		ds := make([]time.Duration, len(rs))
		for i, v := range rs {
			ds[i] = v.RequestDuration
		}
		fmt.Println(id)
		fmt.Println(calculateDurationPercentiles(ds))

	}

}
