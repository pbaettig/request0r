package main

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/pbaettig/randurl"
)

type Test struct {
	ID                      string
	Specs                   []randurl.URLSpec
	NumRequests             int
	TargetRequestsPerSecond int
	Concurrency             int
	Out                     chan WorkerResult
	Stats                   chan WorkerStats
	in                      chan string

	running   bool
	waitGroup *sync.WaitGroup
}

func (t *Test) Start() {
	t.in = make(chan string, len(t.Specs)*t.NumRequests)
	t.Out = make(chan WorkerResult, len(t.Specs)*t.NumRequests)
	t.Stats = make(chan WorkerStats, t.Concurrency)
	t.waitGroup = new(sync.WaitGroup)
	wp := WorkerParams{
		In:         t.in,
		Out:        t.Out,
		Stats:      t.Stats,
		parentTest: t,
	}
	// Start Workers
	for i := 0; i < t.Concurrency; i++ {
		go wp.RunWorker(fmt.Sprintf("%s-%d", t.ID, i))
		// 60% of the time the Workers don't actually start without sleeping for a inbetween iterations...
		time.Sleep(1 * time.Microsecond)
	}
	t.running = true

	// Populate in channel
	for _, spec := range t.Specs {
		for i := 0; i < t.NumRequests; i++ {
			t.in <- spec.String()
		}
	}
	close(t.in)

	// Cleanup after all Workers finish
	go func() {
		t.waitGroup.Wait()
		t.running = false
		close(t.Out)

		//time.Sleep(100 * time.Microsecond) // -> Race condition when closing stats channel
		close(t.Stats)
	}()
}
func (t *Test) Wait() {
	t.waitGroup.Wait()
}

func (t *Test) IsRunning() bool {
	return t.running
}

type WorkerResult struct {
	URL             string
	RequestDuration time.Duration
	StatusCode      int
	ContentLength   int64
	Header          http.Header
	Error           *url.Error
}
type WorkerStats struct {
	ID                string
	RequestsProcessed int
	RequestsPerSecond float64
	Runtime           time.Duration
}
type WorkerParams struct {
	In    chan string
	Out   chan WorkerResult
	Stats chan WorkerStats
	//WaitGroup         *sync.WaitGroup
	//RequestsPerSecond int
	parentTest *Test
}

func (wp *WorkerParams) RunWorker(id string) {
	wp.parentTest.waitGroup.Add(1)
	workerStart := time.Now()
	processed := 0
	defer func() {
		t := time.Now().Sub(workerStart)

		wp.Stats <- WorkerStats{
			ID:                id,
			RequestsPerSecond: float64(processed) / t.Seconds(),
			RequestsProcessed: processed,
			Runtime:           t,
		}
		wp.parentTest.waitGroup.Done()
	}()
	for _url := range wp.In {
		var result WorkerResult

		result.URL = _url

		requestStart := time.Now()
		resp, err := http.Get(_url)
		if err != nil {
			result.Error = err.(*url.Error)
		} else {
			// Don't defer close, we want to get rid of it immediately
			resp.Body.Close()

			result.StatusCode = resp.StatusCode
			result.Header = resp.Header
			result.ContentLength = resp.ContentLength
		}
		result.RequestDuration = time.Now().Sub(requestStart)

		wp.Out <- result
		processed++
		if wp.parentTest.TargetRequestsPerSecond > 0 {
			// Calculate target rps for worker
			wrps := wp.parentTest.TargetRequestsPerSecond / wp.parentTest.Concurrency
			trps := time.Duration(wrps)

			// Sleep if the request was faster than the target rps allows
			if result.RequestDuration < (time.Second / trps) {
				td := (time.Second / trps) - result.RequestDuration
				time.Sleep(td)
			}
		}

	}
}
