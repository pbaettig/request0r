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

func calculateDurationPercentiles(rs []Result) map[float64]time.Duration {
	// convert duration to int to facilitate sorting
	ints := make([]int, len(rs))
	for i, r := range rs {
		ints[i] = int(r.RequestDuration)
	}
	sort.Ints(ints)

	percentiles := [...]float64{0.01, 0.05, 0.10, 0.20, 0.30, 0.40, 0.50, 0.60, 0.70, 0.80, 0.90, 0.99}

	// prepare return value
	ret := make(map[float64]time.Duration)

	for _, p := range percentiles {
		length := float64(len(rs))
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

func collectResults(out <-chan Result) map[string][]Result {
	ret := make(map[string][]Result)
	for r := range out {
		ret[r.ID] = append(ret[r.ID], r)
		//fmt.Printf("%s: %s [%d] [%d ms]\n", r.ID, r.URL, r.StatusCode, r.RequestDuration/time.Millisecond)
	}
	return ret
}

func main() {
	tests := []Test{
		Test{
			ID:          "status-test",
			NumRequests: 100,
			Concurrency: 4,
			Delay:       0,
			Specs: []randurl.URLSpec{
				randurl.URLSpec{
					Scheme: "http",
					Host:   "httpbin.org",
					Components: []randurl.PathComponent{
						randurl.StringComponent("status"),
						randurl.IntegerComponent{Min: 100, Max: 511},
					},
				},
			},
		},
		Test{
			ID:          "delay-test",
			NumRequests: 50,
			Concurrency: 50,
			Delay:       0, //100 * time.Millisecond,
			Specs: []randurl.URLSpec{
				randurl.URLSpec{
					Scheme: "http",
					Host:   "httpbin.org",
					Components: []randurl.PathComponent{
						randurl.StringComponent("delay"),
						randurl.IntegerComponent{Min: 1, Max: 2},
					},
				},
			},
		},
	}
	test := tests[1]

	out := make(chan Result, 100)
	go test.Start(out)
	fmt.Println("***************** Test started")
	// Race Condition
	time.Sleep(1 * time.Millisecond)
	fmt.Println(test.IsRunning())
	test.Wait()
	fmt.Println("***************** Test finished")
	for id, r := range collectResults(out) {
		fmt.Printf("%s: Request duration percentiles\n", id)
		for p, d := range calculateDurationPercentiles(r) {
			fmt.Printf("%d%%: %v\n", int(p*100), d)
		}
	}
	fmt.Println(test.IsRunning())

}
