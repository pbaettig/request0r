package main

import (
	"io/ioutil"

	"github.com/pbaettig/randurl"
	yaml "gopkg.in/yaml.v2"
)

// Stub for parsing root of yaml document
type doc struct {
	Tests []testYaml `yaml:"tests"`
}

// Stub for parsing Test objects
type testYaml struct {
	ID                      string        `yaml:"id"`
	Specs                   []urlSpecYaml `yaml:"urlSpecs"`
	NumRequests             int           `yaml:"numRequests"`
	TargetRequestsPerSecond int           `yaml:"targetRequestsPerSecond"`
	Concurrency             int           `yaml:"concurrency"`
}

// Stub for parsing randurl.URLSpec
type urlSpecYaml struct {
	Scheme     string                        `yaml:"scheme"`
	Host       string                        `yaml:"host"`
	Components []map[interface{}]interface{} `yaml:"uriComponents"`
}

func loadTestsFromFile(path string) ([]*Test, error) {
	// This slice is filled with the final Test objects
	var loadedTests []*Test

	// data, err := ioutil.ReadFile("components.yaml")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return loadedTests, err
	}

	// Load list of testYaml into m
	m := new(doc)
	err = yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		return loadedTests, err
	}

	// Go thrhough all testYaml structs
	for _, mt := range m.Tests {
		// Prepare the proper Test object, whill be added to
		// loadedtests once all of its members have been populated
		lt := Test{
			ID:                      mt.ID,
			Concurrency:             mt.Concurrency,
			NumRequests:             mt.NumRequests,
			TargetRequestsPerSecond: mt.TargetRequestsPerSecond,
		}
		// Go through all urlSpecYaml structs in mt
		for _, urlSpec := range mt.Specs {
			// Prepare the Proper randurl.URLSpec object, whill will be
			// stored in lt.Specs
			spec := randurl.URLSpec{
				Scheme: urlSpec.Scheme,
				Host:   urlSpec.Host,
			}

			// Go through all uriComponents in urlSpecYaml
			// They are untyped, the logic below determines
			// the appropriate type by looking at the "type" field
			// and constructs the correct object
			for _, c := range urlSpec.Components {
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
			lt.Specs = append(lt.Specs, spec)
		}
		loadedTests = append(loadedTests, &lt)
	}
	return loadedTests, nil
}
