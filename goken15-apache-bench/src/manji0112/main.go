// 多重度分のGoRoutineを起動し、request分の
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

var (
	concurrency int
	count       int
)

func main() {
	flag.IntVar(&concurrency, "c", 1, "Number of multiple requests")
	flag.IntVar(&count, "n", 1, "Number of requests")
	flag.Parse()
	url := flag.Arg(0)

	listOfResult := make([]int64, concurrency)
	resultChan := make(chan int64)
	listOfClient := make([]*Client, concurrency)

	for i := 0; i < concurrency; i++ {
		listOfClient[i] = NewClient(url, count, resultChan)
	}

	tic := time.Now()

	for _, c := range listOfClient {
		go c.Loop()
	}

	for i := 0; i < concurrency; i++ {
		listOfResult[i] = <-resultChan
	}

	toc := time.Now()

	total := int64(toc.Sub(tic) / time.Millisecond)
	average := sum(listOfResult) / int64(concurrency)
	reqPersec := 1000 / average

	fmt.Printf("total time: %v [ms]\n", total)
	fmt.Printf("average time: %v [ms]\n", average)
	fmt.Printf("req per sec: %v [#/sec]", reqPersec)
}

type Client struct {
	URL    string
	Count  int
	Result chan int64
}

func NewClient(url string, count int, result chan int64) *Client {
	return &Client{url, count, result}
}

func (c *Client) Loop() {
	listOfResult := make([]int64, c.Count)

	for i := 0; i < c.Count; i++ {
		tic := time.Now()

		res, err := http.Get(c.URL)
		if err != nil {
			log.Fatal(err)
		}
		res.Body.Close()

		toc := time.Now()

		listOfResult[i] = int64(toc.Sub(tic) / time.Millisecond)
	}
	c.Result <- sum(listOfResult) / int64(c.Count)
}

func sum(n []int64) int64 {
	var s int64 = 0

	for i := 0; i < len(n); i++ {
		s += n[i]
	}

	return s
}
