package main

/*
- 基本的なアイディア
多重化はgoroutine,複数のリクエストはループで処理する

-処理の流れ
flagのパース->アドレスのパース->多重化の数だけgoroutineでall_postを回す->全部終わったら集計


- すこし迷ったところ
結果の持ち方:全部を１つの配列に入れるか，goroutineごとに結果の配列をもつか

*/

import(
	"fmt"
	"flag"
	"time"
	"log"
	"net"
	"net/url"
	"bufio"
	"math"
)

var(
	num_request int
	concurrency int
)

// flag parse
func init(){
	flag.IntVar(&num_request,"n",1,"Number of requests")
	flag.IntVar(&concurrency,"c",1,"Number of concurency")
}

// Main Logic
type Result struct{
	status string
	start time.Time
	end time.Time
}

// one request
func post(adress string,url *url.URL) (string,error){
	conn,err := net.Dial("tcp",adress)
	if err != nil{
		log.Printf("connection error")
		return "",err
	}
	fmt.Fprintf(conn,"GET %v HTTP/1.1\r\nHost: %v\r\n\r\n", url.Path, url.Host)
	status, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil{
		log.Printf("fail to get Request")
		return "",err
	}
	return status,nil
}

// all request
func all_post(c chan * Result,adress string, url *url.URL,requests int,){
	for i := 0; i < requests; i++ {
		start_time := time.Now()
		status,err:= post(adress,url)
		if err != nil{
			log.Fatal(err)
		}
		end_time := time.Now()
		temp := &Result{status, start_time, end_time}
		c <- temp
	}
}

// Utility functions
func convert(t time.Time) int64 {
	return t.Unix()*1000 + t.UnixNano()/1000000
}


func main(){
	flag.Parse()
	url, err := url.Parse(flag.Arg(0))
	if err != nil{
		log.Fatal(err)
	}
	host,port,err := net.SplitHostPort(url.Host)
	if err != nil{
		host = url.Host
		port = "80"
	}	
	address := fmt.Sprintf("%v:%v",host,port)
	c := make(chan *Result, num_request * concurrency)

	min_start :=int64(math.MaxInt64)
	max_end := int64(0)
	sum := int64(0)

	for i := 0; i < concurrency; i++ {
		go func(){
			all_post(c,address,url,num_request)
		}()
	}

	// reduce
	for i := 0; i < num_request * concurrency;i++{
		temp := <-c

		start := convert(temp.start)
		end:= convert(temp.end)

		// Update start time
		if start < min_start{
			min_start = start
		}

		// Update end time
		if end > max_end{
			max_end = end
		}

		// sum 
		sum = sum + (end-start)
	}

	fmt.Printf("total time: %v [ms]\n",max_end - min_start)
	fmt.Printf("average time: %v [ms]\n",sum/int64(num_request * concurrency))
	fmt.Printf("req per sec: %v [#/seq]\n",sum/(max_end - min_start))
}
