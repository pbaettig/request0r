package main

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

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
	log.WithFields(log.Fields{
		"test": t.ID,
	}).Debugf("Created  in channel %p", t.in)

	t.Out = make(chan WorkerResult, len(t.Specs)*t.NumRequests)
	log.WithFields(log.Fields{
		"test": t.ID,
	}).Debugf("Created  out channel %p", t.Out)

	t.Stats = make(chan WorkerStats, t.Concurrency)
	log.WithFields(log.Fields{
		"test": t.ID,
	}).Debugf("Created  stats channel %p", t.Stats)

	t.waitGroup = new(sync.WaitGroup)
	log.WithFields(log.Fields{
		"test": t.ID,
	}).Debugf("Created  WaitGroup %p", t.waitGroup)

	// Start Workers
	for i := 0; i < t.Concurrency; i++ {
		//t.waitGroup.Add(1)
		wid := fmt.Sprintf("%s-%d", t.ID, i)
		log.WithFields(log.Fields{
			"test": t.ID,
		}).Debugf("Starting worker %s", wid)
		go t.runWorker(wid)
		// 60% of the time the Workers don't actually start without sleeping inbetween iterations...
		time.Sleep(1 * time.Microsecond)
	}
	t.running = true

	// Populate in channel
	log.WithFields(log.Fields{
		"test": t.ID,
	}).Debugf("Filling in channel %p...", t.in)
	for _, spec := range t.Specs {
		for i := 0; i < t.NumRequests; i++ {
			t.in <- spec.String()
		}
	}
	close(t.in)

	// Cleanup after all Workers finish
	go func() {
		// Allow a bit of time for workers to register in waitGroup
		// before we start waiting
		time.Sleep(200 * time.Millisecond)
		t.waitGroup.Wait()
		t.running = false
		close(t.Out)
		close(t.Stats)
	}()
}
func (t *Test) Wait() {
	log.WithFields(log.Fields{
		"test": t.ID,
	}).Debugf("Wait() on WaitGroup %p", t.waitGroup)
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

// type WorkerParams struct {
// 	In    chan string
// 	Out   chan WorkerResult
// 	Stats chan WorkerStats
// 	//WaitGroup         *sync.WaitGroup
// 	//RequestsPerSecond int
// 	parentTest *Test
// }

func (t *Test) runWorker(id string) {
	t.waitGroup.Add(1)
	log.WithFields(log.Fields{
		"test":   t.ID,
		"worker": id,
	}).Debugf("Worker starting")
	workerStart := time.Now()
	processed := 0
	defer func(t Test) {
		rt := time.Now().Sub(workerStart)

		t.Stats <- WorkerStats{
			ID:                id,
			RequestsPerSecond: float64(processed) / rt.Seconds(),
			RequestsProcessed: processed,
			Runtime:           rt,
		}
		log.WithFields(log.Fields{
			"test":   t.ID,
			"worker": id,
		}).Debugf("Finished. Calling Done on waitGroup %p", t.waitGroup)
		t.waitGroup.Done()
	}(*t)

	log.WithFields(log.Fields{
		"test":   t.ID,
		"worker": id,
	}).Debugf("Reading URLs to process from %p", t.in)
	for _url := range t.in {
		log.WithFields(log.Fields{
			"test":   t.ID,
			"worker": id,
		}).Debugf("Processing %s", _url)
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
		log.WithFields(log.Fields{
			"test":   t.ID,
			"worker": id,
		}).Debugf("Putting result to out channel %p", t.Out)
		t.Out <- result
		processed++
		if t.TargetRequestsPerSecond > 0 {
			// Calculate target rps for worker
			wrps := t.TargetRequestsPerSecond / t.Concurrency
			trps := time.Duration(wrps)
			if trps > 0 {
				// Sleep if the request was faster than the target rps allows
				if result.RequestDuration < (time.Second / trps) {
					td := (time.Second / trps) - result.RequestDuration
					time.Sleep(td)
				}
			}

		}

	}
}
