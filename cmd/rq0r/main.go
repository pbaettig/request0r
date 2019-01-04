package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

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

	flag.StringVar(&testsFilename, "filename", "", "Path to file containing the test definitions")
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
	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if testsFilename == "" {
		usage()
		os.Exit(1)
	}

	tests, err := app.LoadTestsFromFile("/home/bpc/go/src/github.com/pbaettig/request0r/tests.yaml")
	if err != nil {
		log.Fatalln("Unable to load tests from file.")
	}

	if len(tests) == 0 {
		log.Fatalln("No tests defined.")
	}

	testWait := new(sync.WaitGroup)
	testResults := make(map[string][]app.WorkerResult)

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

	// sleep some more to ensure any remaining log output is not
	// mixed in with the results below
	time.Sleep(200 * time.Millisecond)
	fmt.Printf("\n-----------------------\n\n")
	for id, results := range testResults {
		fmt.Printf("# Results for %s\n", id)
		fmt.Println("## Respone duration Percentiles")
		pd := resultutils.GetDurationPercentiles(results)
		fmt.Printf("%d%%\t%s\n", 99, pd[0.99])
		fmt.Printf("%d%%\t%s\n", 95, pd[0.95])
		fmt.Printf("%d%%\t%s\n", 90, pd[0.9])
		fmt.Printf("%d%%\t%s\n", 50, pd[0.5])
		fmt.Printf("%d%%\t%s\n", 10, pd[0.1])
		fmt.Printf("%d%%\t%s\n", 1, pd[0.01])
		fmt.Println()
		fmt.Println("## Errors")
		errors := resultutils.GetErrors(results)
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
		for s, c := range resultutils.CountResponseStatusCodes(results) {
			p := float64(c) / float64(len(results))
			fmt.Printf("HTTP%d\t%.1f%%\t(%d)\n", s, p*100, c)
			sum += c
		}
		fmt.Printf("\t\t(%d total)", sum)
		fmt.Println()
		fmt.Println()
	}

}
