package main

import (
	"fmt"
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
	waitgroup   *sync.WaitGroup
}

func (t *Test) Start(out chan Result) {
	in := make(chan WorkItem, len(t.Specs)*t.NumRequests)
	t.waitgroup = new(sync.WaitGroup)

	// Start Workers
	for i := 0; i < t.Concurrency; i++ {
		go worker(fmt.Sprintf("%s-%d", t.ID, i), t.waitgroup, in, out)
		t.waitgroup.Add(1)
	}
	t.running = true

	// Populate in channel
	for _, spec := range t.Specs {
		for i := 0; i < t.NumRequests; i++ {
			in <- WorkItem{ID: t.ID, URL: spec.String(), Delay: t.Delay}
		}
	}
	close(in)

	go func() {
		t.waitgroup.Wait()
		fmt.Printf("%s finished.\n", t.ID)
		t.running = false
		close(out)
	}()
}

func (t *Test) Wait() {
	t.waitgroup.Wait()
}

func (t *Test) IsRunning() bool {
	return t.running
}

func worker(id string, wg *sync.WaitGroup, in <-chan WorkItem, out chan<- Result) {
	processed := 0
	fmt.Printf("Worker %s started.\n", id)
	defer func() {
		fmt.Printf("Worker %s finished. Processed %d items.\n", id, processed)
		wg.Done()
	}()
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
		processed++

		if wi.Delay > 0 {
			time.Sleep(wi.Delay)
		}

	}
}

type WorkItem struct {
	ID    string
	URL   string
	Delay time.Duration
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
