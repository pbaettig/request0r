package randurl

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

type PathComponent interface {
	String() string
}

type URLSpec struct {
	Scheme, Host string
	Components   []PathComponent
}

func (u URLSpec) String() string {
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("%s://%s", u.Scheme, u.Host))
	for _, s := range u.Components {
		b.WriteString(fmt.Sprintf("/%s", s.String()))
	}
	return b.String()
}

type StringComponent string

func (s StringComponent) String() string {
	return string(s)
}

var validStatuses = []int{100, 101, 102, 103, 200, 201, 202, 203, 204, 205, 206,
	207, 208, 226, 300, 301, 303, 304, 305, 307, 308, 400, 401, 402, 403, 404,
	405, 406, 407, 408, 409, 410, 411, 412, 413, 414, 415, 416, 417, 421, 422,
	423, 424, 425, 426, 427, 429, 430, 451, 500, 501, 502, 503, 504, 505, 506,
	507, 508, 509, 510, 511,
}

type RandomHTTPStatusComponent struct {
	Ranges []int
}

func (hs RandomHTTPStatusComponent) String() string {
	var vc []int
	// Build a list of all valid codes that were requested in `Ranges`
	for _, r := range hs.Ranges {
		for _, v := range validStatuses {
			if r/100 == v/100 {
				vc = append(vc, v)
			}
		}

	}
	return strconv.Itoa(vc[rand.Intn(len(vc))])

}

type RandomIntegerComponent struct {
	Min int
	Max int
}

func (i RandomIntegerComponent) String() string {
	n := rand.Intn(i.Max-i.Min) + i.Min
	return strconv.Itoa(n)
}

const (
	LowercaseAlphabetChars = "abcdefghijklmnopqrstuvwxyz"
	UppercaseAlphabetChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	AlphabetChars          = LowercaseAlphabetChars + UppercaseAlphabetChars
	DigitChars             = "0123456789"
	PunctuationChars       = ".,-_+!()[]{}*"
)

type RandomStringComponent struct {
	Chars  []rune
	Format string
}

func makeRandomString(minLength, maxLength int, chars []rune) string {
	var targetLength int
	if maxLength == minLength {
		targetLength = maxLength
	} else {
		targetLength = rand.Intn(maxLength-minLength) + minLength
	}

	randomChars := make([]rune, targetLength)
	for i := 0; i < targetLength; i++ {
		randomChars[i] = chars[rand.Intn(len(chars))]
	}

	return string(randomChars)
}

func (r RandomStringComponent) String() string {
	// Set default format string if none was provided
	if r.Format == "" {
		r.Format = "%s"
	}

	// Regexp to identify the min,max values, if provided to the format spec
	formatSpecRegex := regexp.MustCompile(`%(?P<min>\d+),(?P<max>\d+)s`)

	// Make a slice containing all the format specs provided in the format string
	formatSpecs := formatSpecRegex.FindAllString(r.Format, -1)

	args := make([]interface{}, len(formatSpecs))
	for i, f := range formatSpecs {
		matches := formatSpecRegex.FindStringSubmatch(f)
		if matches != nil {
			min, _ := strconv.Atoi(matches[1])
			max, _ := strconv.Atoi(matches[2])
			args[i] = makeRandomString(min, max, r.Chars)
		}

	}

	// Replace custom format specs with %s to make it compatible
	// with golang format string
	return fmt.Sprintf(formatSpecRegex.ReplaceAllString(r.Format, "%s"), args...)

}
