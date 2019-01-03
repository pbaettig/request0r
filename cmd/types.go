package main

/*
import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pbaettig/randurl"
)

type WorkerConfig struct {
	inChannel         chan WorkItem
	OutChannel        chan Result
	StatsChannel      chan WorkerStats
	RequestsPerSecond int
	//NumRequests       int
	waitGroup *sync.WaitGroup
}

func NewWorkerConfig(reqps int) *WorkerConfig {
	return &WorkerConfig{
		inChannel:         nil,
		OutChannel:        nil,
		StatsChannel:      nil,
		RequestsPerSecond: reqps,
		waitGroup:         new(sync.WaitGroup),
	}
}

type Test struct {
	ID           string
	Specs        []randurl.URLSpec
	NumRequests  int
	Concurrency  int
	WorkerConfig *WorkerConfig
	running      bool
	//waitgroup    *sync.WaitGroup
}

func (t *Test) Start(out chan Result) {
	t.WorkerConfig.inChannel = make(chan WorkItem, len(t.Specs)*t.NumRequests)
	t.WorkerConfig.OutChannel = make(chan Result, len(t.Specs)*t.NumRequests)
	t.WorkerConfig.StatsChannel = make(chan WorkerStats, t.Concurrency)

	// Start Workers
	for i := 0; i < t.Concurrency; i++ {
		go t.WorkerConfig.Start(fmt.Sprintf("%s-%d", t.ID, i))
	}
	t.running = true

	// Populate in channel
	for _, spec := range t.Specs {
		for i := 0; i < t.NumRequests; i++ {
			t.WorkerConfig.inChannel <- WorkItem{ID: t.ID, URL: spec.String()}
		}
	}
	close(t.WorkerConfig.inChannel)

	go func() {
		t.WorkerConfig.waitGroup.Wait()
		fmt.Printf("%s finished.\n", t.ID)
		t.running = false
		close(t.WorkerConfig.OutChannel)

		//time.Sleep(100 * time.Microsecond) // -> Race condition when closing stats channel
		close(t.WorkerConfig.StatsChannel)
	}()
}

func (t *Test) Wait() {
	t.WorkerConfig.waitGroup.Wait()
}

func (t *Test) IsRunning() bool {
	return t.running
}

func (c WorkerConfig) Start(id string) {
	c.waitGroup.Add(1)
	workerStart := time.Now()
	processed := 0
	defer func() {
		t := time.Now().Sub(workerStart)
		c.StatsChannel <- WorkerStats{
			ID:                id,
			RequestsPerSecond: float64(processed) / t.Seconds(),
			RequestsProcessed: processed,
			Runtime:           t,
		}
		c.waitGroup.Done()
	}()
	for wi := range c.inChannel {
		var result Result
		result.ID = wi.ID
		result.URL = wi.URL

		requestStart := time.Now()
		resp, err := http.Get(wi.URL)
		if err != nil {
			result.Error = err
		} else {
			defer resp.Body.Close()
			result.StatusCode = resp.StatusCode
			result.Header = resp.Header
			result.ContentLength = resp.ContentLength
		}
		result.RequestDuration = time.Now().Sub(requestStart)

		c.OutChannel <- result
		processed++

		trps := time.Duration(c.RequestsPerSecond)
		if result.RequestDuration < (time.Second / trps) {
			td := (time.Second / trps) - result.RequestDuration
			time.Sleep(td)
		}

	}
}

type WorkerStats struct {
	ID                string
	RequestsProcessed int
	RequestsPerSecond float64
	Runtime           time.Duration
}

/* func worker(id string, wg *sync.WaitGroup, in <-chan WorkItem, out chan<- Result) {
	workerStart := time.Now()
	processed := 0
	fmt.Printf("Worker %s started.\n", id)
	defer func() {
		t := time.Now().Sub(workerStart)
		fmt.Printf("Worker %s finished. Processed %d items in %s (%f requests/second).\n",
			id, processed, t, float64(processed)/t.Seconds())
		wg.Done()
	}()
	for wi := range in {
		var result Result
		result.ID = wi.ID
		result.URL = wi.URL

		requestStart := time.Now()
		resp, err := http.Get(wi.URL)
		if err != nil {
			result.Error = err
		} else {
			defer resp.Body.Close()
			result.StatusCode = resp.StatusCode
			result.Header = resp.Header
			result.ContentLength = resp.ContentLength
		}
		result.RequestDuration = time.Now().Sub(requestStart)

		out <- result
		processed++

		trps := time.Duration(20)
		if result.RequestDuration < (time.Second / trps) {
			td := (time.Second / trps) - result.RequestDuration
			time.Sleep(td)
		}
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
*/
