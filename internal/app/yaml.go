package app

import (
	"io/ioutil"
	"strconv"

	"github.com/pbaettig/request0r/pkg/randurl"
	yaml "gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
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

func castString(sourceValue interface{}) string {
	switch sourceValue.(type) {
	case string:
		return sourceValue.(string)
	case int:
		return strconv.Itoa(sourceValue.(int))
	default:
		log.Fatalln("Uh oh")
	}

	return ""
}

func castInt(sourceValue interface{}) int {
	switch sourceValue.(type) {
	case int:
		return sourceValue.(int)
	case string:
		i, err := strconv.Atoi(sourceValue.(string))
		if err != nil {
			log.Fatalf(err.Error())
		}
		return i
	default:
		log.Fatalln("Uh oh")
	}

	return 0
}

// LoadTestsFromFile parses the specified yaml file and return a slice of *Test
func LoadTestsFromFile(path string) ([]*Test, error) {
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
					spec.Components = append(spec.Components, randurl.StringComponent(castString(c["value"])))
				case "integer":
					spec.Components = append(spec.Components, randurl.IntegerComponent{
						Min: castInt(c["min"]),
						Max: castInt(c["max"]),
					})
				case "randomString":
					spec.Components = append(spec.Components, randurl.RandomStringComponent{
						MinLength: castInt(c["minLength"]),
						MaxLength: castInt(c["maxLength"]),
						Chars:     []rune(castString(c["chars"])),
						Format:    castString(c["format"]),
					})

				case "httpStatus":
					ns := make([]int, 0)
					for _, n := range c["ranges"].([]interface{}) {
						ns = append(ns, castInt(n))
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
