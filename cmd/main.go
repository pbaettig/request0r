package main

import (
	"fmt"
	"math/rand"
	"net/url"
	"sort"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
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
		if r.Error == nil {
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

func main() {
	log.SetLevel(log.DebugLevel)

	tests, err := loadTestsFromFile("../tests.yaml")
	if err != nil {
		log.Fatalln("Unable to load tests from file.")
	}

	if len(tests) == 0 {
		log.Fatalln("No tests defined.")
	}

	testWait := new(sync.WaitGroup)
	testResults := make(map[string][]WorkerResult)

	for _, test := range tests {
		test.Start()
		testWait.Add(1)

		log.WithFields(log.Fields{
			"test": test.ID,
		}).Info("Test started")

		go func(t *Test, wg *sync.WaitGroup) {
			t.Wait()
			log.WithFields(log.Fields{
				"test": t.ID,
			}).Debug("Finished")
			wg.Done()

			i := 0
			log.WithFields(log.Fields{
				"test": t.ID,
			}).Debugf("Reading results from %p", t.Out)
			for r := range t.Out {
				i++
				testResults[t.ID] = append(testResults[t.ID], r)
				log.WithFields(log.Fields{
					"test": t.ID,
				}).Debugf("Got result for %s (%d/%d)", r.URL, i, t.NumRequests)
			}
			log.WithFields(log.Fields{
				"test": t.ID,
			}).Debug("All results processed")
		}(test, testWait)
	}

	log.Info("Waiting for all tests to finish...")
	testWait.Wait()
	log.Info("All done.")
	// sleep some more to ensure any remaining log output is not
	// mixed in with the results below
	time.Sleep(200 * time.Millisecond)
	fmt.Printf("\n-----------------------\n\n")
	for id, results := range testResults {
		fmt.Printf("# Results for %s\n", id)
		fmt.Println("## Respone duration Percentiles")
		pd := getDurationPercentiles(results)
		fmt.Printf("%d%%\t%s\n", 99, pd[0.99])
		fmt.Printf("%d%%\t%s\n", 95, pd[0.95])
		fmt.Printf("%d%%\t%s\n", 90, pd[0.9])
		fmt.Printf("%d%%\t%s\n", 50, pd[0.5])
		fmt.Printf("%d%%\t%s\n", 10, pd[0.1])
		fmt.Printf("%d%%\t%s\n", 1, pd[0.01])
		fmt.Println()
		fmt.Println("## Errors")
		errors := getErrors(results)
		errorPercent := float64(len(errors)) * 100.0 / float64(len(results))
		if errorPercent > 0 {
			fmt.Printf("%.1f%% (%d/%d) of requests failed.\n", errorPercent, len(errors), len(results))

			if len(errors) <= 10 {
				fmt.Println("Error messages:")
				for _, e := range errors {
					fmt.Printf("- %s\n", e)
				}
			} else {
				fmt.Println("More than 10 errors occured. First 10 error messages:")
				for _, e := range errors[:10] {
					fmt.Printf("- %s\n", e)
				}
			}

			fmt.Println()

		} else {
			fmt.Println("No errors occured.")
		}
		fmt.Println()
		fmt.Println("## Response Status codes")
		sum := 0
		for s, c := range countResponseStatusCodes(results) {
			p := float64(c) / float64(len(results))
			fmt.Printf("HTTP%d\t%.1f%%\t(%d)\n", s, p*100, c)
			sum += c
		}
		fmt.Printf("\t\t(%d total)", sum)
		fmt.Println()
		fmt.Println()
	}

}
