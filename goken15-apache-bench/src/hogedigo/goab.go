package main

/*
	goken課題：apache benchmark
	concurrency数だけgoroutineつくった
	結果chan受診時に終了したgoroutineを数えて終了してるけどヘン?
	都度http.Get呼んでるのはやっぱり重いのかな・・・
 */
import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

type result struct {
	seq        int
	start      int64
	end        int64
	statusCode int
	err        error
}

type summary struct {
	start int64
	end   int64
	count      int
	successes  int
	failures   int
	errors     int
	sum      int64
	min        int64
	max        int64
}

func (s *summary) addResult(r *result) {

	s.count++

	if r.err != nil {
		s.errors++
	} else if r.statusCode >= 200 && r.statusCode < 300 {
		s.successes++
	} else {
		s.failures++
	}

	tat := r.end - r.start
	if s.min == 0 || s.min > tat {
		s.min = tat
	}
	if s.max == 0 || s.max < tat {
		s.max = tat
	}
	s.sum += r.end - r.start
}

func main() {
	var c, n int
	flag.IntVar(&c, "c", 3, "concurrency")
	flag.IntVar(&n, "n", 10, "number of requests")
	flag.Parse()

	url := flag.Arg(0)
	if url == "" {
		log.Fatalf("url must be specified.")
	}

	fmt.Printf("concurrency:%d\n", c)
	fmt.Printf("number of requests:%d\n", n)

	seqs := sequence(n)
	results := make(chan *result, c)

	var s summary

	s.start = time.Now().UnixNano()

	for i := 0; i < c; i++ {
		go test(url, results, seqs)
	}

	doneTests := 0
	for r := range results {
		if r == nil {
			doneTests++
			if doneTests >= c {
				break
			}
			continue
		}

		s.addResult(r)

		if s.count > 0 && s.count%100 == 0 {
			fmt.Printf("Completed %d requests\n", s.count)
		}
	}

	s.end = time.Now().UnixNano()

	printSummary(&s)
}

func test(url string, results chan *result, seqs chan int) {
	for {
		seq, ok := <-seqs
		if !ok {
			results <- nil
			return
		}
		start := time.Now().UnixNano()

		statusCode, err := doGet(url)

		end := time.Now().UnixNano()
		results <- &result{seq, start, end, statusCode, err}
	}
}

func doGet(url string) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func sequence(max int) chan int {
	ch := make(chan int)
	go func() {
		for i := 1; i <= max; i++ {
			ch <- i
		}
		close(ch)
	}()
	return ch
}

func printSummary(s *summary) {
	fmt.Printf("### done! ###\n")
	fmt.Printf("total time: %d[ms]\n", (s.end-s.start)/int64(time.Millisecond))
	fmt.Printf("max time: %.1f[ms]\n", float64(s.max)/float64(time.Millisecond))
	fmt.Printf("min time: %.1f[ms]\n", float64(s.min)/float64(time.Millisecond))
	fmt.Printf("average time: %.1f[ms]\n", float64(s.sum)/float64(s.count)/float64(time.Millisecond))
	fmt.Printf("req per sec: %d[#/seq]\n", int64(s.count)/((s.end-s.start)/int64(time.Second)))
	fmt.Printf("total requests: %d\n", s.count)
	fmt.Printf("success count: %d\n", s.successes)
	fmt.Printf("failure count: %d\n", s.failures)
	fmt.Printf("error count: %d\n", s.errors)
}
