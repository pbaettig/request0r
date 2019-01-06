package app

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/pbaettig/request0r/pkg/randurl"
)

func TestLoadTestsFromFile(t *testing.T) {
	tmpFile1, err := ioutil.TempFile(os.TempDir(), "prefix-")
	if err != nil {
		t.Errorf("Cannot create temporary file: %s", err)
	}
	defer os.Remove(tmpFile1.Name())
	text := []byte(`
tests:
- id: unit-test
  numRequests: 10
  concurrency: 10
  targetRequestsPerSecond: 20
  urlSpecs:
  - scheme: https
    host: test-host.tester.local
    uriComponents:
    - type: string
      value: user
    - type: randomString
      minLength: 7
      maxLength: 7
      chars: 1234abc
      format: "user-%s"  
    - type: integer
      min: 10
      max: 20
    - type: httpStatus
      ranges:
      - 500
`)
	if _, err = tmpFile1.Write(text); err != nil {
		t.Errorf("Failed to write to temporary file: %s", err)
	}

	correctTest := Test{
		ID:                      "unit-test",
		NumRequests:             10,
		Concurrency:             10,
		TargetRequestsPerSecond: 20,
		Specs: []randurl.URLSpec{
			randurl.URLSpec{
				Scheme: "https",
				Host:   "test-host.tester.local",
				Components: []randurl.PathComponent{
					randurl.StringComponent("user"),
					randurl.RandomStringComponent{
						MinLength: 7,
						MaxLength: 7,
						Chars:     []rune("1234abc"),
						Format:    "user-%s",
					},
					randurl.IntegerComponent{
						Min: 10,
						Max: 20,
					},
					randurl.HTTPStatusComponent{
						Ranges: []int{500},
					},
				},
			},
		},
	}

	loadedTests, err := LoadTestsFromFile(tmpFile1.Name())
	if err != nil {
		t.Error(err)
	}
	loadedTest := loadedTests[0]
	if loadedTest.ID != correctTest.ID {
		t.Errorf("Loaded test is incorrect, wanted test ID \"%s\", got \"%s\"", correctTest.ID, loadedTest.ID)
	}
	if loadedTest.NumRequests != correctTest.NumRequests {
		t.Errorf("Loaded test is incorrect, wanted test NumRequests \"%d\", got \"%d\"", correctTest.NumRequests, loadedTest.NumRequests)
	}
	if loadedTest.Concurrency != correctTest.Concurrency {
		t.Errorf("Loaded test is incorrect, wanted test Concurrency \"%d\", got \"%d\"", correctTest.Concurrency, loadedTest.Concurrency)
	}
	if loadedTest.TargetRequestsPerSecond != correctTest.TargetRequestsPerSecond {
		t.Errorf("Loaded test is incorrect, wanted test TargetRequestsPerSecond \"%d\", got \"%d\"", correctTest.TargetRequestsPerSecond, loadedTest.TargetRequestsPerSecond)
	}

	for i := range loadedTest.Specs[0].Components {
		switch i {
		case 0:
			loaded := loadedTest.Specs[0].Components[i].(randurl.StringComponent)
			correct := correctTest.Specs[0].Components[i].(randurl.StringComponent)
			if loaded != correct {
				t.Errorf("Loaded test uriComponent[%d] is incorrect, wanted %s, got %s", i, correct, loaded)

			}
		case 1:
			loaded := loadedTest.Specs[0].Components[i].(randurl.RandomStringComponent)
			correct := correctTest.Specs[0].Components[i].(randurl.RandomStringComponent)
			if !reflect.DeepEqual(loaded, correct) {
			}
		case 2:
			loaded := loadedTest.Specs[0].Components[i].(randurl.IntegerComponent)
			correct := correctTest.Specs[0].Components[i].(randurl.IntegerComponent)
			if !reflect.DeepEqual(loaded, correct) {
				t.Errorf("Loaded test uriComponent[%d] is incorrect, wanted %+v, got %+v", i, correct, loaded)
			}
		case 3:
			loaded := loadedTest.Specs[0].Components[i].(randurl.HTTPStatusComponent)
			correct := correctTest.Specs[0].Components[i].(randurl.HTTPStatusComponent)
			if !reflect.DeepEqual(loaded, correct) {
				t.Errorf("Loaded test uriComponent[%d] is incorrect, wanted %+v, got %+v", i, correct, loaded)
			}
		}
	}

}
