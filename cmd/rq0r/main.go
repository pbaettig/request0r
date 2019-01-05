package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/pbaettig/request0r/internal/pkg/statutils"

	"github.com/pbaettig/request0r/internal/app"
	"github.com/pbaettig/request0r/internal/pkg/resultutils"
	log "github.com/sirupsen/logrus"
)

var (
	testsFilename string
	debug         bool
)

func init() {
	rand.Seed(time.Now().UnixNano())

	flag.StringVar(&testsFilename, "tests", "", "Path to file containing the test definitions")
	flag.BoolVar(&debug, "debug", false, "Enable verbose debug logging")
}

func usage() {
	fmt.Println(`rq0r executes HTTP tests described in a YAML file. It's purpose is to provide an
easy way to generate load on a target service and provide some statistics about the responses
it received.`)
	fmt.Println()

	fmt.Println("PARAMETERS:")
	flag.PrintDefaults()
}

func main() {
	flag.Parse()

	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if testsFilename == "" {
		usage()
		os.Exit(1)
	}

	tests, err := app.LoadTestsFromFile(testsFilename)
	if err != nil {
		log.Fatalf("Unable to load tests from file: %s", err)
	}

	if len(tests) == 0 {
		log.Fatalln("No tests defined.")
	}

	testWait := new(sync.WaitGroup)
	testResults := make(map[string][]app.WorkerResult)
	testStats := make(map[string][]app.WorkerStats)

	testStart := time.Now()
	for _, test := range tests {
		test.Start()
		testWait.Add(1)

		log.WithFields(log.Fields{
			"test": test.ID,
		}).Info("Started")

		go func(t *app.Test, wg *sync.WaitGroup) {
			t.Wait()
			log.WithFields(log.Fields{
				"test": t.ID,
			}).Info("Finished")
			wg.Done()

			// Collect worker stats
			for s := range t.Stats {
				testStats[t.ID] = append(testStats[t.ID], s)
			}

			i := 0
			log.WithFields(log.Fields{
				"test": t.ID,
			}).Debugf("Reading results from %p", t.Out)

			// Collect test results
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
	testsDuration := time.Now().Sub(testStart)
	log.Infof("Ran %d Tests in %s", len(tests), testsDuration)

	// sleep some more to ensure any remaining log output is not
	// mixed in with the results below
	time.Sleep(200 * time.Millisecond)
	fmt.Printf("\n-----------------------\n\n")
	for _, test := range tests {
		fmt.Printf("# Results for test \"%s\"\n", test.ID)
		fmt.Println("## Worker Stats")
		fmt.Printf("Worker Runtime\n")
		for _, s := range testStats[test.ID] {
			fmt.Printf("%s:\t%.1f requests/second\t%d requests processed\t(in %s)\n", s.ID, s.RequestsPerSecond, s.RequestsProcessed, s.Runtime)
		}
		trps := statutils.SumRequestsPerSecond(testStats[test.ID])
		fmt.Printf("\t\t%.1f total\t\t%d total\n", trps, test.NumRequests)
		fmt.Println()
		fmt.Println("## Respone duration Percentiles")
		pd := resultutils.GetDurationPercentiles(testResults[test.ID])
		fmt.Printf("%d%%\t%s\n", 99, pd[0.99])
		fmt.Printf("%d%%\t%s\n", 95, pd[0.95])
		fmt.Printf("%d%%\t%s\n", 90, pd[0.9])
		fmt.Printf("%d%%\t%s\n", 50, pd[0.5])
		fmt.Printf("%d%%\t%s\n", 10, pd[0.1])
		fmt.Printf("%d%%\t%s\n", 1, pd[0.01])
		fmt.Println()
		fmt.Println("## Errors")
		errors := resultutils.GetErrors(testResults[test.ID])
		errorPercent := float64(len(errors)) * 100.0 / float64(len(testResults[test.ID]))
		if errorPercent > 0 {
			fmt.Printf("%.1f%% (%d/%d) of requests failed.\n", errorPercent, len(errors), len(testResults[test.ID]))

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
		for s, c := range resultutils.CountResponseStatusCodes(testResults[test.ID]) {
			p := float64(c) / float64(len(testResults[test.ID]))
			fmt.Printf("HTTP%d\t%.1f%%\t(%d)\n", s, p*100, c)
			sum += c
		}
		fmt.Printf("\t\t(%d total)", sum)
		fmt.Println()
		fmt.Println()
	}

}
