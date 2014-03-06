package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	n int64
	c int
	url string
)

func init() {
	log.SetFlags(log.Lshortfile)
	flag.Int64Var(&n, "n", 1, "number of requests")
	flag.IntVar(&c, "c", 1, "number of clients")
	flag.Parse()
	url = os.Args[len(os.Args)-1]
}

func main() {
	start := time.Now()
	var i int64
	for i = 0; i < n; i++ {
		resp, err := http.Get(url)
		if err != nil {
			log.Println(resp, err)
		}
	}
	total := time.Since(start)
	avg := time.Duration(total.Nanoseconds() / n)
	rps := int64(time.Second / avg)

	format := `
total time: %v [ms]
average time: %v [ms]
req per sec: %v [#/sec]
`

	fmt.Printf(format, total, avg, rps)
}
