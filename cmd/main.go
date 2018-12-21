package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/pbaettig/randurl"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	urlSpec := randurl.URLSpec{
		Scheme: "http",
		Host:   "httpbin.org",
		Components: []randurl.PathComponent{
			randurl.StaticComponent("api"),
			randurl.StaticComponent("user"),
			randurl.RandomComponent{
				Chars:     []rune(randurl.AlphabetChars),
				MinLength: 2,
				MaxLength: 12,
			},
			randurl.StaticComponent("setid"),
			randurl.RandomComponent{
				Chars:     []rune(randurl.DigitChars),
				MinLength: 12,
				MaxLength: 12,
			},
		},
	}
	for i := 0; i < 100; i++ {
		url := urlSpec.String()
		fmt.Println(url)
	}

	// resp, err := http.Get(url)
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	defer resp.Body.Close()
	// 	fmt.Printf("%s: %d\n", url, resp.StatusCode)
	// }

}
