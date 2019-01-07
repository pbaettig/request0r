package randurl

import (
	"math/rand"
	"regexp"
	"testing"
	"time"
)

type randomStringComponentTestPair struct {
}

func TestRandomStringComponent_String(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	testTable := []struct {
		in     RandomStringComponent
		verify *regexp.Regexp
	}{
		{RandomStringComponent{Format: "%1,8s", Chars: []rune("ABCDEFGHIJKLMNOPQRSTUVWYZ")}, regexp.MustCompile(`^[ABCDEFGHIJKLMNOPQRSTUVWYZ]{1,8}$`)},
		{RandomStringComponent{Format: "user-%7,7s", Chars: []rune("abcdef0123456789")}, regexp.MustCompile(`^user-[abcdef0123456789]{7}$`)},
		{RandomStringComponent{Format: "num-%1,32s", Chars: []rune("0123456789")}, regexp.MustCompile(`^num-\d{1,32}$`)},
	}

	for _, test := range testTable {
		// run the test a bunch of times to test
		// a bunch of generated strings.
		for i := 0; i < 50; i++ {
			actual := test.in.String()

			if m := test.verify.MatchString(actual); !m {
				t.Errorf("Run %d: Regex verification failed, generated string \"%s\" does not match regex \"%s\"", i, actual, test.verify)
			}
		}

	}
}
