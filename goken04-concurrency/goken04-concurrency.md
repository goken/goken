Go研 vol.4 まとめ
==================

## 参加者

- [Jxck](https://twitter.com/Jxck_)
- [tenntenn](https://twitter.com/tenntenn)
- [manji0112](https://twitter.com/manji0112)
- [hogedigo](https://twitter.com/hogedigo)
- [yssk22](https://twitter.com/yssk22)

## 無断欠席

- [aita](https://twitter.com/aita)

## 今回の概要

- 開催日: 2013/07/17(Wed)
- connpass: http://connpass.com/event/2855/
- 発表者: Jxck
- バージョン: go1.1
- テーマ: Advanced Go Concurrency Patterns 2

## 資料

Docker をテーマに予定していたけれど、発表担当者が来れなくなってしまったので、
急遽前回書いた Blog を元に、 Advanced Go Concurrency Patterns を復習。


- [blog](http://d.hatena.ne.jp/Jxck/20130623/1371999576)
- [gist](https://gist.github.com/Jxck/5831269)

Gist の Revision を追いながら、実装を追加していく様子を解説。


## 話題になったところ

### buffer のスライス

https://gist.github.com/Jxck/5831269#file-pooling-go-L80

buffer の中身が最後の一個の場合は大丈夫なのか？

```go
var buffer []string
buffer = append(buffer, "hoge")

fmt.Println(buffer[1:]) // []
```

これは別に問題ない。


### channel の buffer が 1

https://gist.github.com/Jxck/5831269#file-pooling-go-L64

channel の buffer を 1 にしているが、これは何故か。
おそらく buffer 0 だと書き込みがブロックしてしまい、
書き込む側の goroutine が終われなくなる可能性があるから。

https://gist.github.com/Jxck/5831269#file-pooling-go-L67

buffer を 1 にしておけば、とりあえず一つは goroutine が書き込めて、
goroutine が終われるようになる。
buffer が 0 でも動くけど、 goroutine が書き込み待ちで溜まったり、
読み出しがされなくなった goroutine がゾンビ化する可能性があるかもしれない。


### そもそも buffer は slice ?

この例では buffer として maxLength の slice を使っているけど、
いずれかの channel の buffer を maxLength にすればそれで済むのではないか？


一応やってみた。

3 つの時間を調整して、 buffer の中身や起動している Goroutine の数をモニタできるように出力してみた。

- Get() にかかる時間(getInterval)
- 配信の遅延(publishInterval)
- クライアント側の呼び出しの遅延(subscribeInterval)

配列の場合は、うまくやらないと buffer が溢れてたりしたけど、 channel の buffer はそれはない。
ブロックして goroutine が大量に起動し続けるようなバグもないように、 len(), cap() で buffer の空きで調整しているので goroutine の数も安定している。


```go
package main

import (
	"fmt"
	"log"
	"runtime"
	"time"
)

// server will panic after timeout second
func halt(timeout int64) {
	time.AfterFunc(time.Duration(timeout)*time.Second, func() {
		panic("fin")
	})
}

func Get() string {
	time.Sleep(getInterval)
	return "hello"
}

type Pooling struct {
	result chan string
	fin    chan bool
}

var (
	getInterval       time.Duration = 2 * time.Second
	subscribeInterval               = 1 * time.Second
	publishInterval                 = 200 * time.Millisecond
)

func (p *Pooling) Loop() {
	for {
		log.Println("before select", len(p.result), runtime.NumGoroutine())
		interval := time.After(publishInterval)

		select {
		case <-interval:
			if len(p.result) < cap(p.result) {
				go func() {
					p.result <- Get()
					log.Println("   after read", len(p.result), runtime.NumGoroutine())
				}()
			}
		case <-p.fin:
			log.Println("close")
			return
		}
	}
}

func (p *Pooling) Close() {
	close(p.fin)
}

func main() {

	//go halt(10)

	pooling := &Pooling{
		result: make(chan string, 10),
		fin:    make(chan bool, 1),
	}
	go pooling.Loop()

	var i = time.Tick(subscribeInterval)
	var fin = time.After(4 * time.Second)
	for {
		select {
		case <-i:
			log.Println("main loop")

			fmt.Println(<-pooling.result)
			fmt.Println("=====================================")
		case <-fin:
			log.Println("close loop")
			pooling.Close()
			return
		}
	}
}
```

