# Go研 Vol.11

## testing

Go の標準パッケージでテストを実行するための仕組みを持つ。
正確には、以下を提供する。

- テストを落とす(Fail系, Error系)
- テストをスキップする(Skip 系)
- その他(Parallel etc)


逆によくあるテスティングフレームワークのような、以下の昨日は提供しない。

- Assert
- テストを構造化するなにか
- テストを非同期にするなにか
- その他


この辺の方針に付いては、 2013 Go Advent Calendar で書いた。

[Go の Test に対する考え方](http://qiita.com/Jxck_/items/8717a5982547cfa54ebc)

簡単なテストの実行方法と、命名規則周りも上記を参照とし省略。


また、実際には testing は Bench や Example など色々提供しているが、今回は純粋なテスト部分に注目する。


## go test コマンド

Go は標準の go コマンドで test の実行もサポートする。

```
$ go test fmt
```


ここからたどっていく。

[http://golang.org/src/cmd/go/test.go]


実行部は、テンプレートになっていてここでフラグなどを対応させたコードを生成している。
[http://golang.org/src/cmd/go/test.go#1132]


テスト自体は、 testing.Main()


```go
func main() {
{{if .CoverEnabled}}
  testing.RegisterCover(testing.Cover{
    Mode: {{printf "%q" .CoverMode}},
    Counters: coverCounters,
    Blocks: coverBlocks,
    CoveredPackages: {{printf "%q" .Covered}},
  })
{{end}}
  testing.Main(matchString, tests, benchmarks, examples)
}
```

tests は t.Tests に格納されたテストケースの配列。
これを読んでるのが以下。

[http://golang.org/src/cmd/go/test.go#1106]


```go
func (t *testFuncs) load(filename, pkg string, seen *bool) error {
	f, err := parser.ParseFile(testFileSet, filename, nil, parser.ParseComments)
	if err != nil {
		return expandScanner(err)
	}
	for _, d := range f.Decls {
		n, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if n.Recv != nil {
			continue
		}
		name := n.Name.String()
		switch {
		case isTest(name, "Test"):
			t.Tests = append(t.Tests, testFunc{pkg, name, ""})
			*seen = true
		case isTest(name, "Benchmark"):
			t.Benchmarks = append(t.Benchmarks, testFunc{pkg, name, ""})
			*seen = true
		}
	}
	ex := doc.Examples(f)
	sort.Sort(byOrder(ex))
	for _, e := range ex {
		if e.Output == "" && !e.EmptyOutput {
			// Don't run examples with no output.
			continue
		}
		t.Examples = append(t.Examples, testFunc{pkg, "Example" + e.Name, e.Output})
		*seen = true
	}
	return nil
}
```

go/parser でファイルをまるごとパースしている。

```
func ParseFile(fset *token.FileSet, filename string, src interface{}, mode Mode) (f *ast.File, err error)
```

ast.File.Decl はがトップレベルの宣言がとれ、これを ast.FuncDecl にキャストして関数宣言に。

isTest() は、 TestXxx という命名規則の確認。
Test() 自体でも良い模様。


```go
func isTest(name, prefix string) bool {
	if !strings.HasPrefix(name, prefix) {
		return false
	}
	if len(name) == len(prefix) { // "Test" is ok
		return true
		}
		rune, _ := utf8.DecodeRuneInString(name[len(prefix):])
		return !unicode.IsLower(rune)
	}
}
```


matchString() は -run regexp オプションで渡されるパターンを確認するための関数。


```go
var matchPat string
var matchRe *regexp.Regexp

func matchString(pat, str string) (result bool, err error) {
	if matchRe == nil || matchPat != pat {
		matchPat = pat
		matchRe, err = regexp.Compile(matchPat)
		if err != nil {
			return
		}
	}
	return matchRe.MatchString(str), nil
}
```


これらを渡す先が、 testing パッケージの Main()

[http://golang.org/src/pkg/testing/testing.go#L396]


```go
// An internal function but exported because it is cross-package; part of the implementation
// of the "go test" command.
func Main(matchString func(pat, str string) (bool, error),
          tests []InternalTest,
          benchmarks []InternalBenchmark,
          examples []InternalExample) {

	flag.Parse()
	parseCpuList()

	before()
	startAlarm()
	haveExamples = len(examples) > 0
	testOk := RunTests(matchString, tests)
	exampleOk := RunExamples(matchString, examples)
	stopAlarm()
	if !testOk || !exampleOk {
		fmt.Println("FAIL")
		os.Exit(1)
	}
	fmt.Println("PASS")
	RunBenchmarks(matchString, benchmarks)
	after()
}
```

色々前準備をしてから、 RunTests() を実行。


```go
func RunTests(matchString func(pat, str string) (bool, error), tests []InternalTest) (ok bool) {
	ok = true
	if len(tests) == 0 && !haveExamples {
		fmt.Fprintln(os.Stderr, "testing: warning: no tests to run")
		return
	}
	for _, procs := range cpuList { // -cpu 1,2,4 とか指定した値で MAXPROCS を変えて実行。
		runtime.GOMAXPROCS(procs)
		// We build a new channel tree for each run of the loop.
		// collector merges in one channel all the upstream signals from parallel tests.
		// If all tests pump to the same channel, a bug can occur where a test
		// kicks off a goroutine that Fails, yet the test still delivers a completion signal,
		// which skews the counting.
		var collector = make(chan interface{})

		numParallel := 0
		startParallel := make(chan bool)

		for i := 0; i < len(tests); i++ {
      // テストケースの名前が run で指定したものかをチェック
			matched, err := matchString(*match, tests[i].Name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "testing: invalid regexp for -test.run: %s\n", err)
				os.Exit(1)
			}
			if !matched {
				continue // マッチするものが一つも無くてもエラーとかにはならない。
			}
			testName := tests[i].Name
			if procs != 1 {
				// go test -cpu 2 とかすると -2 がつく
				testName = fmt.Sprintf("%s-%d", tests[i].Name, procs)
			}
			// testing.T の生成。これがテスト関数に渡される。
			t := &T{
				common: common{
					signal: make(chan interface{}),
				},
				name:          testName,
				startParallel: startParallel,
			}
			t.self = t
			if *chatty { // -v 詳細出力
				fmt.Printf("=== RUN %s\n", t.name)
			}
			go tRunner(t, &tests[i]) // ここは中の defer で t.signal <- t している
			out := (<-t.signal).(*T) // その読み出しでブロック、 out は実行したテスト自身
			if out == nil { // Parallel run.
				go func() {
					collector <- <-t.signal
				}()
				numParallel++
				continue
			}
			t.report()
			ok = ok && !out.Failed()
		}

		running := 0
		for numParallel+running > 0 {
			if running < *parallel && numParallel > 0 {
				startParallel <- true
				running++
				numParallel--
				continue
			}
			t := (<-collector).(*T)
			t.report()
			ok = ok && !t.Failed()
			running--
		}
	}
	return
}
```


testing.T は以下。
基本機能は common メソッドとして生えていて、それを mixin している。


```go
// common holds the elements common between T and B and
// captures common methods such as Errorf.
type common struct {
	mu      sync.RWMutex // guards output and failed
	output  []byte       // Output generated by test or benchmark.
	failed  bool         // Test or benchmark has failed.
	skipped bool         // Test of benchmark has been skipped.

	start    time.Time // Time test or benchmark started
	duration time.Duration
	self     interface{}      // To be sent on signal channel when done.
	signal   chan interface{} // Output for serial tests.
}

type T struct {
	common
	name          string    // Name of test.
	startParallel chan bool // Parallel tests will wait on this.
}
```


```go
func tRunner(t *T, test *InternalTest) {
	// When this goroutine is done, either because test.F(t)
	// returned normally or because a test failure triggered
	// a call to runtime.Goexit, record the duration and send
	// a signal saying that the test is done.
	defer func() {
		t.duration = time.Now().Sub(t.start)
		// If the test panicked, print any test output before dying.
		if err := recover(); err != nil {
			t.Fail()
			t.report()
			panic(err)
		}
		t.signal <- t
	}()

	t.start = time.Now()
	test.F(t)
}
```

テストの各関数に対して、 t を渡す。これにより、各テスト関数は t.*T を受け取るように実装されている必要がある。
テストが成功しても失敗しても、 defer で実行時間を算出しシグナルにテスト関数自身を送る。



tRunner の後を再掲

```go
	go tRunner(t, &tests[i]) // ここは中の defer で t.signal <- t している
	out := (<-t.signal).(*T) // その読み出しでブロック、 out は実行したテスト自身
	if out == nil { // Parallel run.
		go func() {
			collector <- <-t.signal
		}()
		numParallel++
		continue // メインループを継続させる
	}
	t.report()
	ok = ok && !out.Failed()
}
```

out が nil になるパターンは、 Parallel() が呼び出されているとき。
この場合は、t.signal への終了処理通知を collector にリダイレクトする goroutine を立てる。

```go
// Parallel signals that this test is to be run in parallel with (and only with)
// other parallel tests.
func (t *T) Parallel() {
	t.signal <- (*T)(nil) // Release main testing loop
	<-t.startParallel     // Wait for serial tests to finish
	// Assuming Parallel is the first thing a test does, which is reasonable,
	// reinitialize the test's start time because it's actually starting now.
	t.start = time.Now()
}
```

Parallel() は t.signal に偽の通知(nil) を送って、 t.startPrallel の通知を待つ。
numparallel が起動予定の parallel の数。




```go
running := 0
for numParallel+running > 0 {
	if running < *parallel && numParallel > 0 {
		startParallel <- true
		running++
		numParallel--
		continue
	}
	t := (<-collector).(*T)
	t.report()
	ok = ok && !t.Failed()
	running--
}
```


Parallel 予約された数が numParallel で、それを全て起動する。
同時に起動できるのは parallel の値までで、それはテストフラグで指定できる。
デフォルトは GOMAXPROCS 数。

collector から一個取り出して実行。




Parallel() を使った例

```go
func TestSum1(t *testing.T) {
	t.Parallel()
	a := 3
	b := Sum(1, 2)
	if a != b {
		t.Errorf("got %v\nwant %v", a, b)
	}
}

func TestSum2(t *testing.T) {
	a := 3
	b := Sum(1, 2)
	if a != b {
		t.Errorf("got %v\nwant %v", a, b)
	}
}
```

Parallel() を抜いた例と、入れた例の実行結果。
Run が先に来ていることがわかる。

```sh
$ go test -v # no Parallel
=== RUN TestSum1
--- PASS: TestSum1 (0.00 seconds)
=== RUN TestSum2
--- PASS: TestSum2 (0.00 seconds)
PASS
ok      _/tmp/sample/src        0.003s

$ go test -v # Parallel
=== RUN TestSum1
=== RUN TestSum2
--- PASS: TestSum2 (0.00 seconds)
--- PASS: TestSum1 (0.00 seconds)
PASS
ok      _/tmp/sample/src        0.002s
```




