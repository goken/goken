package abc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"time"
)

// パラメータを表す構造体
type Params struct {
	Concurrency int
	Requests    int
	Url         string
}

func (p *Params) Print() {
	fmt.Printf("Concurrency: %d\n", p.Concurrency)
	fmt.Printf("Requests   : %d\n", p.Requests)
	fmt.Printf("URL        : %s\n", p.Url)
}

// ベンチマークを表す構造体
type Benchmark struct {
	params Params
}

func NewBenchmark(params Params) *Benchmark {
	return &Benchmark{params}
}

func (b *Benchmark) Params() Params {
	return b.params
}

// ベンチマークを実行
func (b *Benchmark) Run() {
	b.params.Print()
	fmt.Println()

	runtime.GOMAXPROCS(runtime.NumCPU())

	// 並行数分の容量を持つチャネルを作成
	resultChan := make(chan results, b.params.Concurrency)
	for i := 0; i < b.params.Concurrency; i++ {
		go func() {
			request := NewRequest(b.params.Url)
			resultChan <- request.Execute(b.params.Requests)
		}()
	}

	var rs results = make([]result, 0, b.params.Concurrency)
	for i := 0; i < b.params.Concurrency; i++ {
		rs = rs.add(<-resultChan)
	}

	fmt.Printf("total time: %v [sec]\n", rs.end().Sub(rs.start()).Seconds())
	fmt.Printf("average time: %v [ms]\n", rs.average())
	fmt.Printf("req per sec %v [req/sec]\n", rs.throughput())
}

type request struct {
	url *url.URL
}

type result struct {
	start time.Time
	end   time.Time
}

func (r result) interval() time.Duration {
	return r.end.Sub(r.start)
}

func NewRequest(s string) *request {
	u, _ := url.Parse(s)
	return &request{u}
}

// 集計用のメソッドを定義するため型定義
type results []result

// 実行時間の合計を計算
func (rs results) sum() time.Duration {
	sum := time.Duration(0)
	for _, r := range rs {
		sum += r.interval()
	}
	return sum
}

// 実行時間の平均を計算
func (rs results) average() float64 {
	return float64(rs.sum()/time.Millisecond) / float64(len(rs))
}

// 結果を連結
func (rs results) add(a results) results {
	rs = append(rs, a...)
	return rs
}

// 結果の中から最も開始時間が早いものを取得
func (rs results) start() time.Time {
	if len(rs) == 0 {
		return time.Unix(0, 0)
	}
	min := rs[0].start
	for _, r := range rs {
		if min.After(r.start) {
			min = r.start
		}
	}

	return min
}

// 単位時間あたりの実行回数を計算
func (rs results) throughput() int {
	return len(rs) / int(rs.end().Sub(rs.start())/time.Second)
}

// 結果の中から最も終了時間が遅いものを取得
func (rs results) end() time.Time {
	if len(rs) == 0 {
		return time.Unix(0, 0)
	}
	max := rs[0].end
	for _, r := range rs {
		if max.Before(r.end) {
			max = r.end
		}
	}

	return max
}

// リクエストを一回実行
func (r *request) ExecuteOnce() (result, error) {
	var result result
	result.start = time.Now()
	resp, err := http.Get(r.url.String())
	if err != nil {
		return result, err
	}
	ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	result.end = time.Now()

	return result, nil
}

// 指定回数だけリクエストを実行
func (r *request) Execute(requests int) results {
	rs := make([]result, 0, requests)
	for i := 0; i < requests; i++ {
		result, err := r.ExecuteOnce()
		if err != nil {
			continue
		}
		rs = append(rs, result)
	}

	return rs
}
