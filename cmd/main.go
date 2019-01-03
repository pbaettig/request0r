package main

import (
	"fmt"
	"math/rand"
	"net/url"
	"sort"
	"time"

	"github.com/pbaettig/randurl"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func collectResults(c chan WorkerResult) []WorkerResult {
	var rs []WorkerResult
	for r := range c {
		rs = append(rs, r)
	}
	return rs
}

func countResponseStatusCodes(rs []WorkerResult) map[int]int {
	sc := make(map[int]int)
	for _, r := range rs {
		if _, ok := sc[r.StatusCode]; !ok {
			sc[r.StatusCode] = 1
		} else {
			sc[r.StatusCode]++
		}

	}
	return sc
}

func getErrors(rs []WorkerResult) []*url.Error {
	var errors []*url.Error
	for _, r := range rs {
		if r.Error == nil {
			continue
		}
		errors = append(errors, r.Error)

	}
	return errors
}

func getDurationPercentiles(rs []WorkerResult) map[float64]time.Duration {
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

/*
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
*/

// func collectResults(out <-chan WorkerResult) map[string][]WorkerResult {
// 	ret := make(map[string][]WorkerResult)
// 	for r := range out {
// 		ret[r.ID] = append(ret[r.ID], r)
// 		//fmt.Printf("%s: %s [%d] [%d ms]\n", r.ID, r.URL, r.StatusCode, r.RequestDuration/time.Millisecond)
// 	}
// 	return ret
// }

func main() {

	test := Test{
		ID:          "status-test",
		NumRequests: 100000,
		Concurrency: 100,
		//TargetRequestsPerSecond: 5000,
		Specs: []randurl.URLSpec{
			randurl.URLSpec{
				Scheme: "http",
				Host:   "localhost:8080",
				Components: []randurl.PathComponent{
					randurl.StringComponent("data"),
					randurl.IntegerComponent{Min: 1024, Max: 1288},
					// randurl.RandomStringComponent{
					// 	Format:    "user-%s",
					// 	Chars:     []rune(randurl.DigitChars),
					// 	MinLength: 4,
					// 	MaxLength: 8,
					// },
					//randurl.IntegerComponent{Min: 200, Max: 511},
					//randurl.HTTPStatus{Ranges: []int{400, 200}},
				},
			},
		},
	}

	test.Start()
	fmt.Println("***************** Test started")

	// Display Status and collect results
	results := make([]WorkerResult, 0)
	go func() {
		i := 0
		for r := range test.Out {
			i++
			results = append(results, r)
			if i%(test.NumRequests/20) == 0 {
				fmt.Printf("\r%d%% complete", i*100/test.NumRequests)
			}
		}
		fmt.Println()
	}()

	test.Wait()
	fmt.Println("***************** Test finished")

	fmt.Println(countResponseStatusCodes(results))
	fmt.Println(getErrors(results))
	dpc := getDurationPercentiles(results)
	for _, p := range []float64{0.99, 0.95, 0.8, 0.5, 0.1, 0.01} {
		fmt.Printf("%d%%:\t%s\n", int(p*100), dpc[p])
	}

	fmt.Println(test.IsRunning())

}
