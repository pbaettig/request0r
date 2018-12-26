package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/pbaettig/randurl"
)

type Test struct {
	ID          string
	Specs       []randurl.URLSpec
	NumRequests int
	Concurrency int
	Delay       time.Duration
	running     bool
}

func (t *Test) Start(out chan Result) {
	in := make(chan string, len(t.Specs)*t.NumRequests)
	wg := new(sync.WaitGroup)

	// Start Workers
	for i := 0; i < len(t.Specs); i++ {
		go worker(wg, in, out)
	}
	t.running = true

	// Populate in channel
	for _, spec := range t.Specs {
		for i := 0; i < numRequests; i++ {
			in <- spec.String()
		}
	}
	close(in)

	go func() {
		wg.Wait()
		t.running = false
	}()
}

func (t *Test) IsRunning() bool {
	return t.running
}

func worker(wg *sync.WaitGroup, in <-chan string, out chan<- Result) {
	defer wg.Done()
	for url := range in {
		var result Result
		result.URL = url

		start := time.Now()
		resp, err := http.Get(url)
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

type WorkItem struct {
	ID  string
	URL string
}

type Result struct {
	ID              string
	URL             string
	RequestDuration time.Duration
	StatusCode      int
	ContentLength   int64
	Header          http.Header
	Error           error
}
