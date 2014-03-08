package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

var (
	n   int
	c   int
	url string
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.Lshortfile)
	flag.IntVar(&n, "n", 1, "number of requests")
	flag.IntVar(&c, "c", 1, "number of clients")
	flag.Parse()
	url = os.Args[len(os.Args)-1]
}

func seq(max int) <-chan int {
	i := make(chan int)
	go func() {
		for {
			i <- max
			max = max - 1
			if max <= 0 {
				close(i)
				break
			}
		}
	}()
	return i
}

func main() {
	var wg sync.WaitGroup
	s := seq(n)

	start := time.Now()
	for i := 0; i < c; i++ {
		wg.Add(1)
		go func(j int) {
			for _ = range s {
				resp, err := http.Get(url)
				if err != nil {
					log.Println(resp, err)
				}
				resp.Body.Close()
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	total := time.Since(start)
	avg := total.Seconds() / float64(n) * 1000
	rps := (float64(n) / total.Seconds())

	format := `
total time: %.3f [s]
average time: %.3f [ms]
req per sec: %.3f [#/sec]
`

	fmt.Printf(format, total.Seconds(), avg, rps)
}
