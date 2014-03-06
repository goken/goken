package main

/*
DNSキャッシュ
並列実行のスタート時点を可能な限り近づけた
*/

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	u "net/url"
	"time"
)

var requests int
var concurrency int

func init() {
	flag.IntVar(&requests, "n", 1, "Number of requests to perform")
	flag.IntVar(&concurrency, "c", 1, "Number of multiple requests to make")
}

type ResponseResult struct {
	status string
	start  time.Time
	end    time.Time
}

func get(address string, url *u.URL) (string, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	fmt.Fprintf(conn, "GET %v HTTP/1.1\r\nHost: %v\r\n\r\n", url.Path, url.Host)
	status, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", err
	}
	return status, nil
}

func run(c chan *ResponseResult, address string, url *u.URL, requests int) {
	for i := 0; i < requests; i++ {
		start := time.Now()
		status, err := get(address, url)
		if err != nil {
			log.Fatal(err)
		}
		end := time.Now()

		rr := &ResponseResult{status, start, end}
		c <- rr
	}
}

func msec(t time.Time) int64 {
	return t.Unix()*1000 + t.UnixNano()/1000000
}

func main() {
	flag.Parse()

	n := requests * concurrency
	url, err := u.Parse(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	host, port, err := net.SplitHostPort(url.Host)
	if err != nil {
		host = url.Host
		port = "80"
	}

	addrs, err := net.LookupHost(host)
	if err != nil {
		log.Fatal(err)
	}

	// prepare
	address := fmt.Sprintf("%v:%v", addrs[0], port)
	c := make(chan *ResponseResult, n)
	waits := make([]chan bool, 0, concurrency)
	for i := 0; i < concurrency; i++ {
		w := make(chan bool)
		go func(w <-chan bool) {
			<-w
			run(c, address, url, requests)
		}(w)
		waits = append(waits, w)
	}

	// start
	go func () {
		for _, w := range(waits) {
			close(w)
		}
	}()

	// total
	var d int64
	min_start := int64(math.MaxInt64)
	max_end := int64(0)
	sum := int64(0)
	for i := 0; i < n; i++ {
		rr := <-c

		start := msec(rr.start)
		end := msec(rr.end)

		if start < min_start {
			min_start = start
		}
		if end > max_end {
			max_end = end
		}

		d = end - start
		sum = sum + d
	}
	fmt.Printf("total time: %v [ms]\n", max_end-min_start)
	fmt.Printf("average time: %v [ms]\n", sum/int64(n))
	fmt.Printf("req per sec: %v [#/seq]\n", sum/(max_end-min_start))
}
