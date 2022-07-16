package useragent

import (
	"fmt"
	"math/rand"
)

// ref: https://github.com/gocolly/colly
func RandomUserAgent() string {
	return genChromeUA()
}

var versions = []string{
	"37.0.2062.124",
	"40.0.2214.93",
	"41.0.2228.0",
	"49.0.2623.112",
	"55.0.2883.87",
	"56.0.2924.87",
	"57.0.2987.133",
	"61.0.3163.100",
	"63.0.3239.132",
	"64.0.3282.0",
	"65.0.3325.146",
	"68.0.3440.106",
	"69.0.3497.100",
	"70.0.3538.102",
	"74.0.3729.169",
}

var oss = []string{
	"Macintosh; Intel Mac OS X 10_10",
	"Windows NT 10.0",
	"Windows NT 5.1",
	"Windows NT 6.1; WOW64",
	"Windows NT 6.1; Win64; x64",
	"X11; Linux x86_64",
}

func genChromeUA() string {
	version := versions[rand.Intn(len(versions))]
	os := oss[rand.Intn(len(oss))]
	return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36", os, version)
}
