Go研 vol.3 まとめ
==================

##参加者

- [Jxck](https://twitter.com/Jxck_)

## 今回の概要

- 開催日: 2013/06/12(Wed)
- connpass: http://connpass.com/event/2575/
- 発表者: Jxck
- バージョン: go1.1
- テーマ: Advanced Go Concurrency Patterns

## 資料

- [youtube] https://www.youtube.com/watch?v=QDDwwePbDtw
- [slide] http://talks.golang.org/2013/advconc.slide


## 内容

### p5

同じアドレススペースに goroutine を起動できる。

```go
go f()
go g(1, 2)
```

同期/情報共有のために channel が使える。型付きのメッセージングシステム。

```go
c := make(chan int)
go func() { c <- 3 }()
n := <-c
```

送信と受信の間ではブロックする点が重要。
このへんの話は去年 Rob Pike がしてる。

参照
- http://talks.golang.org/2012/concurrency.slide
- http://jxck.node-ninja.com/slides/gocon-2013spring.html


## #6

```go
type Ball struct{ hits int }

func main() {
    table := make(chan *Ball)
    go player("ping", table)
    go player("pong", table)

    table <- new(Ball) // game on; toss the ball
    time.Sleep(1 * time.Second)
    <-table // game over; grab the ball
}

func player(name string, table chan *Ball) {
    for {
        ball := <-table
        ball.hits++
        fmt.Println(name, ball.hits)
        time.Sleep(100 * time.Millisecond)
        table <- ball
    }
}
```

二つの Goroutine で同じ channel への参照を保持している。
main からの送信で始まり、各 goroutine がそれを受け取る。
1 秒後に main がそれを奪うことで終わる。
(main が終わると goroutine は全て終わる。)

## #7

concurrent の仕組みが Go 自体に組み込まれているメリットのひとつは、
デッドロックの検地が可能なこと。


```go
type Ball struct{ hits int }

func main() {
    table := make(chan *Ball)
    go player("ping", table)
    go player("pong", table)

    // table <- new(Ball) // game on; toss the ball
    time.Sleep(1 * time.Second)
    <-table // game over; grab the ball
}

func player(name string, table chan *Ball) {
    for {
        ball := <-table
        ball.hits++
        fmt.Println(name, ball.hits)
        time.Sleep(100 * time.Millisecond)
        table <- ball
    }
}
```

main からの送信を消す。
channel からの読み込みでデッドロックが発生するため、
1 秒後に stack trace が出力される。

stack trace の内容は、デッドロックの発生源を知るために非常に重要。

## #8

panic を用いると stack trace を取得することができる。
stack trace を取得する方法はいくつかある？


ping pong の最後に panic を起こして stack trace を見る。
main があり、 2 つの worker もまだシステム上では残っていることが確認できる。

無限ループになっているのでわかりやすいが、これは leak である。


## #9

本当なら不要になった Goroutine は死んでほしい。
それをどうやるか、それこそが今回の発表の重要なテーマ。

たとえば Google で作っているようなライフタイムの長いプログラムでは、
不要なリソースは適切に開放しないといけない。
Goroutine もその対象のひとつ。

ここで使えるのが Go の Select 文。
Concurrent をサポートしていて、 Goroutine, Channel に次ぐ三つ目の柱。
Switch に似ているが根本的に違う。


```go
select {
case xc <- x:
    // sent x on xc
case y := <-yc:
    // received y from yc
}
```

この例では、受信と送信をイベントとしてその後を実行できる。
実行中ほかはブロックされる。変数はブロック内で使用できる。



## #10-11

私の大好きな Feed Reader がもうすぐ消えるので、
これを応用して Feed Reader を作ってみよう。


でも、 Feed Reader には詳しくないので、 godoc.org で
RSS を検索してみる。
Bitbucket, Github, Google Code, LaunchPad から godoc を検索できる。


ひとつ選んでインタフェースを見てみる。

```go
// URI を引数にとり Item のスライス、次の時間、エラーを返す
func Fetch(uri string) (items []Item, next time.Time, err error)

// タイトル, Channel, GUID を持つ。
type Item struct{
    Title, Channel, GUID string // RSS 情報の一部
}
```

ほしいのは Item 型の読み込み専用 channel で、この上に UI を組みたい。

```go
<-chan Item
```

## #12

Subscription は複数のソースに対してしたい。
そこで、より詳しくやりたいことを確認する。


Fetch() を考えて、説明のためのフェイク実装がしやすいように Fetcher インタフェースを提供する。

domain を渡すと Fetcher を返す。

```go
type Fetcher interface {
    Fetch() (items []Item, next time.Time, err error)
}

 func Fetch(domain string) Fetcher {...} // fetches Items from domain
```

## #13

```go
// Item の Channel を取得するために、 Subscribe を定義する。
// いらない Stream (channel) は閉じられる(クリーンアップできる)ようにもする。
type Subscription interface {
    Updates() <-chan Item // stream of Items　
    Close() error         // shuts down the stream
}

// 定期的な購読に変換する
func Subscribe(fetcher Fetcher) Subscription {}

// 複数のドメインからの購読をまとめる
func Merge(subs ...Subscription) Subscription {}
```

## #14

すべてをまとめた Example
Fetcher は Fake を使っている。

```
func main() {
    // 各購読を channel で取り、それを合成した channel を作る
    merged := Merge(
        Subscribe(Fetch("blog.golang.org")),
        Subscribe(Fetch("googleblog.blogspot.com")),
        Subscribe(Fetch("googledevelopers.blogspot.com"))
    )

    // 時間がたったら閉じる
    time.AfterFunc(3*time.Second, func() {
        fmt.Println("closed:", merged.Close())
    })

    // Print the stream.
    // channel からの出力は for-range でまわせる
    for it := range merged.Updates() {
        fmt.Println(it.Channel, it.Title)
    }

    panic("show me the stacks")
}
```

channel の for-range は、 channel が close するまで継続して読み込むことを意味する。


最後のスタックトレースでは、 main と runtime が使っている goroutine だけが
表示されて、 subscribe に関する groutine はちゃんと閉じられてることがわかる。


全体: http://play.golang.org/p/jEm8Pt-nH9 [feedreader.go](https://github.com/goken/goken/blob/master/goken03/feedreader.go)

## #15

このトークでは Subscribe に注目する。
重要なのは、定期的な取得を channel を用いた stream に変換していること。

```go
func Subscribe(fetcher Fetcher) Subscription {
    // Subscription の実装
    s := &sub{
        fetcher: fetcher,
        updates: make(chan Item), // for Updates
    }
    go s.loop()
    return s
}

// Subscription を実装
type sub struct {
    fetcher Fetcher   // fetches items
    updates chan Item // delivers items to the user
}

// ここが定期購読を行う goroutine
// 実際は終了処理やデータの戻しも行う
// loop fetches items using s.fetcher and sends them
// on s.updates. loop exits when s.Close is called
func (s *sub) loop() {}
````

## #16

Subscription を実装。
インタフェースは Updates(), Close() なので、
その二つを実装すればよい。

```go
// 読み取り専用の channle を返すだけなので楽
func (s *sub) Updates() <-chan Item{
　    return s.updates
}

// 終了処理とエラー報告
func (s *sub) Close() error {
    // TODO: make loop exit
    // TODO: find out about any error
    return err
}
```

## #17

ループの goroutine は何をするか。

- Fetch を定期的に呼び出す
- 取得した Item を Updates channel に送る
- Close() で終了処理をし、エラーを報告する

## #18

単純な実装。

```go
for {
    if s.closed {
        close(s.updates) // channel を閉じる
        return // 終わる
    }
    items, next, err := s.fetcher.Fetch() // Fetch を呼ぶ
    if err { // エラーがあったら
        s.err = err
        time.Sleep(10 * time.Second)
        continue // 少し止まって繰り返す
    }
    for _, item := range items { // 成功したら
        s.updates <- item // channel に送る
    }
    if now := time.Now(); next.After(now) {
        time.Sleep(next.Sub(now)) // 次の購読まで停止
    }
}
```

```go
// boole にフラグを立てるだけ
func (s *naiveSub) Close() error {
    s.closed = true
    return s.err
}
```

でもこれじゃ buggy なので改善する。

## #19-20

一つ目は、 closed, err にレースコンディションがある。
(二つの goroutine が sync 無しにアクセスしている)

これは race detector でわかる。

## #21

二つ目は、 Sleep のせいで Close() がすぐに実行されない。
closed フラグがセットされた後、 next.Sub() が 1 日後になったら、
loop が止まって 1 日 cleanup されない。

## #22

三つ目は channel への書き込みはブロックするので、読み出されないと終わらない。
もしクライアントが Close() を呼んで、 updates の読み出しをやめたら、
s.updates への Item の書き込みは永遠に止まる。きちんと終了処理されない。

## #23 (15:50)

こうした問題は loop を select を使って書き換えれば直る。
Select で以下の三つのイベントを扱う。

- Close が呼び出された
- Fetch する時間になった
- s.updates に item を送る

## #24

for-select loop を使う。
最初に mutable state を定義、
for の頭では channel を定義。

これはひとつの goroutine で行われているので
race condition が発生しない。


```go
func (s *sub) loop() {
    // declare mutable state
    for {
        // set up channels for case
        select {
        case <-c1:
            // read/write state
        case c2 <- x:
            // read/write state
        case y := <-c3:
            // read/write state
        }
    }
}
```

## #25

Case 1: Close

```go
type sub struct {
    closing chan chan error
}
```

channel を送信できる channel


## #26

request/response パターンを使用

```go
func (s *sub) Close() error {
    errc := make(chan error)
    s.closing <- errc // loop に終了をリクエストする
    return <-errc // レスポンスを待つブロックをする
}

var err error // set when Fetch fails
for {
    select {
    case errc := <-s.closing: // リクエストの受け取り
        errc <- err // 受け取ったのは channel なのでこれにエラーを送る
        close(s.updates) // レシーバに終了を伝える
        return
    }
}
```

## #27

```go
    var pending []Item // appended by fetch; consumed by send
    var next time.Time // initially January 1, year 0
    var err error
    for {
        var fetchDelay time.Duration // initally 0 (no delay)
        if now := time.Now(); next.After(now) {
            fetchDelay = next.Sub(now)
        }
        startFetch := time.After(fetchDelay)

        select {
        case <-startFetch:
            var fetched []Item
            fetched, next, err = s.fetcher.Fetch()
            if err != nil {
                next = time.Now().Add(10 * time.Second)
                break
            }
            pending = append(pending, fetched...)
        }
    }
```

## #28

一個ずつ取り出して送る。

```go
var pending []Item // appended by fetch; consumed by send
for {
    select {
    case s.updates <- pending[0]:
        pending = pending[1:]
    }
}
```

でも、これは pending が空のときうまくいかない。(panic になる)


## #29

nil channel を使う
nil channel に send するとブロックする。 Panic はおこらない。

```go
func main() {
    a, b := make(chan string), make(chan string)
    go func() { a <- "a" }()
    go func() { b <- "b" }()
    if rand.Intn(2) == 0 {
        a = nil
        fmt.Println("nil a")
    } else {
        b = nil
        fmt.Println("nil b")
    }
    select {
    case s := <-a:
        fmt.Println("got", s)
    case s := <-b:
        fmt.Println("got", s)
    }
}
```

これをうまく使うと、 select で「実行したくない case」については
nil channel にしてしまうことで、実行しないようにできる。
これは select が複雑なイベントを処理し始めたとき有効。


## #30

nil channel で send を修正

```go
    var pending []Item // appended by fetch; consumed by send
    for {
        var first Item
        var updates chan Item // ここで nil なので
        if len(pending) > 0 {
            first = pending[0]
            updates = s.updates // ここを通らない限り
        }

        select {
        case updates <- first: // ここは実行されない
            pending = pending[1:]
        }
    }
```

## #31

ここまでをあわせると以下のように書ける。
straight-line で race condition もない。
コールバックも conditional var もない。

```go
   select {
        case errc := <-s.closing:
            errc <- err
            close(s.updates)
            return
        case <-startFetch:
            var fetched []Item
            fetched, next, err = s.fetcher.Fetch()
            if err != nil {
                next = time.Now().Add(10 * time.Second)
                break
            }
            pending = append(pending, fetched...)
        case updates <- first:
            pending = pending[1:]
        }
```

## #32(20:00)

3 つのバグを修正

- s.close , s.err の race
-- 変わりに communication を利用
- time.Sleep が長時間ブロック
-- case のひとつなのでブロックしない
- s.updates が永遠にブロック
-- case のひとつなのでブロックしない

```go
    select {
        case errc := <-s.closing:
            errc <- err
            close(s.updates)
            return
        case <-startFetch:
            var fetched []Item
            fetched, next, err = s.fetcher.Fetch()
            if err != nil {
                next = time.Now().Add(10 * time.Second)
                break
            }
            pending = append(pending, fetched...)
        case updates <- first:
            pending = pending[1:]
        }
```

## #33-35

もっと改善できる。
Fetch が重複を返すことがある。
それを防ぐために、GUID をキーにした map を使う。

```go
    var pending []Item
    var next time.Time
    var err error
    var seen = make(map[string]bool) // set of item.GUIDs

      case <-startFetch:
            var fetched []Item
            fetched, next, err = s.fetcher.Fetch()
            if err != nil {
                next = time.Now().Add(10 * time.Second)
                break
            }
            for _, item := range fetched {
                if !seen[item.GUID] {
                    pending = append(pending, item)
                    seen[item.GUID] = true
                }
            }
```

## #36

受信側が遅いと、 pendin が増えすぎてしまう。

```go
 case <-startFetch:
            var fetched []Item
            fetched, next, err = s.fetcher.Fetch()
            if err != nil {
                next = time.Now().Add(10 * time.Second)
                break
            }
            for _, item := range fetched {
                if !seen[item.GUID] {
                    pending = append(pending, item)
                    seen[item.GUID] = true
                }
            }
```

## #37

nil channel を使って解決できる。

```go
const maxPneding = 10
```

```go
   var fetchDelay time.Duration
        if now := time.Now(); next.After(now) {
            fetchDelay = next.Sub(now)
        }
        var startFetch <-chan time.Time
        if len(pending) < maxPending {
            startFetch = time.After(fetchDelay) // enable fetch case
        }
```

startFetch が fetch のトリガーになっているので、
maxPending を満たさない場合は、これを nil channel にしてしまうことで
シンプルに解決することができる。


## #38

fetcher.Fetch() はネットワーク I/O で外部サーバと話しているのでブロックする。
これを非同期にしたい。
goroutine に移して、終了を検知する必要がある。
select を追加すればいい。

```
        case <-startFetch:
            var fetched []Item
            fetched, next, err = s.fetcher.Fetch()
            if err != nil {
                next = time.Now().Add(10 * time.Second)
                break
            }
            for _, item := range fetched {
                if !seen[item.GUID] {
                    pending = append(pending, item)
                    seen[item.GUID] = true
                }
            }
```

## #39

fetch した結果を格納するための型を導入。

```go
type fetchResult struct{ fetched []Item; next time.Time; err error }
```

fetch が終わっていることを知るための channel を作る

```go
    var fetchDone chan fetchResult // if non-nil, Fetch is running
```

fetchDone が nil だったら fetch をスケジュールする。
これが初期値。

```go
        var startFetch <-chan time.Time
        if fetchDone == nil && len(pending) < maxPending {
            startFetch = time.After(fetchDelay) // enable fetch case
        }
```

case では fetchDone を設定して fetch が走る。
goroutine が走って fetchDone に結果を送る。
fetch が走ってる間も loop はブロックせず、
終わったら fetchDone の読み取り case で
結果を取得する。


```go
        select {
        case <-startFetch:
            fetchDone = make(chan fetchResult, 1)
            go func() {
                fetched, next, err := s.fetcher.Fetch()
                fetchDone <- fetchResult{fetched, next, err}
            }()
        case result := <-fetchDone:
            fetchDone = nil
            // Use result.fetched, result.next, result.err
```

## #40

Three techniques:
- for-select loop
- service channel, reply channels (chan chan error)
- nil channels in select cases

## #41

Go makes it easier

- channels convey data, timer events, cancellation signals
- goroutines serialize access to local mutable state
- stack traces & deadlock detector
- race detector


## Q & A

1. Goroutine の leak を発見するツールとか方法はあるか？

ここでは stack trace を出す基本的な方法をやった、実行中に leak を発見するのは難しい。
Go1.1 で Blocking profiles がある。何がブロックしてるのかを graph 取得できる?

2. 今回の 3 つのエラーなどに当たらないで書くあめの何かが Go の言語にはあるか？
静的解析ツールは？

race detector とかあるけど、あとは経験が必要な部分もある。
ツールはもう少し良くして行きたいけど、 race が一番でかいので、 race detector 大きいと思う。
3. appendgine go でもツール使えるの？ gae go はどうデバッグするの？
それは次のセッションで

4. あとで

