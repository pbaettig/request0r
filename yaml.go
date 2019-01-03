package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"sync"

	"github.com/pbaettig/randurl"

	yaml "gopkg.in/yaml.v2"
)

type TestYaml struct {
	ID                      string        `yaml:"id"`
	Specs                   []URLSpecYaml `yaml:"urlSpecs"`
	NumRequests             int           `yaml:"numRequests"`
	TargetRequestsPerSecond int           `yaml:"targetRequestsPerSecond"`
	Concurrency             int           `yaml:"concurrency"`
	// Out                     chan WorkerResult
	// Stats                   chan WorkerStats
	// in                      chan string

	// running   bool
	// waitGroup *sync.WaitGroup
}
type URLSpecYaml struct {
	Scheme     string                        `yaml:"scheme"`
	Host       string                        `yaml:"host"`
	Components []map[interface{}]interface{} `yaml:"uriComponents"`
}

type PathComponent struct {
	ComponentType string `yaml:"type"`
}

type doc struct {
	Tests []TestYaml
}

type Test struct {
	ID                      string
	Specs                   []randurl.URLSpec
	NumRequests             int
	TargetRequestsPerSecond int
	Concurrency             int
	Out                     chan string
	Stats                   chan string
	in                      chan string

	running   bool
	waitGroup *sync.WaitGroup
}

func main() {
	// data, err := ioutil.ReadFile("components.yaml")
	data, err := ioutil.ReadFile("tests.yaml")
	if err != nil {
		log.Fatalln("Unable to open file")
	}

	//m := make(map[interface{}]interface{})
	m := new(doc)
	// m := new(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("%#v\n", m.Tests[0].Specs[0].Components)
	for _, test := range m.Tests {
		pTest := Test{
			ID:                      test.ID,
			Concurrency:             test.Concurrency,
			NumRequests:             test.NumRequests,
			TargetRequestsPerSecond: test.TargetRequestsPerSecond,
		}

		for _, urlSpec := range test.Specs {
			spec := randurl.URLSpec{
				Scheme: urlSpec.Scheme,
				Host:   urlSpec.Host,
			}
			for i, c := range urlSpec.Components {
				fmt.Printf("======== %d =======\n", i)
				t, ok := c["type"]
				if !ok {
					continue
				}
				switch t {
				case "string":
					spec.Components = append(spec.Components, randurl.StringComponent(c["value"].(string)))
				case "integer":
					spec.Components = append(spec.Components, randurl.IntegerComponent{
						Min: c["min"].(int),
						Max: c["max"].(int),
					})
				case "randomString":
					spec.Components = append(spec.Components, randurl.RandomStringComponent{
						MinLength: c["minLength"].(int),
						MaxLength: c["maxLength"].(int),
						Chars:     []rune(c["chars"].(string)),
						Format:    c["format"].(string),
					})

				case "httpStatus":
					ns := make([]int, 0)
					for _, n := range c["ranges"].([]interface{}) {
						ns = append(ns, n.(int))
					}

					spec.Components = append(spec.Components, randurl.HTTPStatus{
						Ranges: ns,
					})

				}

			}
			pTest.Specs = append(pTest.Specs, spec)
		}
		fmt.Printf("%#v", pTest)
	}

}
