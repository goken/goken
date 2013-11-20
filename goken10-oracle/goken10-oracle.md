# ソースコードの静的解析ツール Go Oracleを使ってみた

## Go Oracleとは？

Go言語のリポジトリを漁ってると、go toolsの中にoracleというディレクトリを発見しました。何だろうなぁと思ってると、ちょうど[golang nuts](https://groups.google.com/forum/#!topic/golang-nuts/CwdIJZs6Tfc)に話題がでたので、触ってみました。

名前がややこしくて、検索しづらいのですが、Go Oracleはソースコードの静的解析ツールのようです。

* https://code.google.com/p/go/source/browse?repo=tools#hg%2Foracle

使い方等は、以下のドキュメント（英語）を見れば良さそうです。
何はともあれ、とりあえず使ってみましょう。Emacsからでも使えるようですが、ここではコマンドから使ってみましょう。

* [ユーザマニュアル](http://golang.org/s/oracle-user-manual)
* [設計書?](https://docs.google.com/viewer?a=v&pid=forums&srcid=MDg3NjYzNDU1NTk0NjU2OTUyMDMBMDY5OTQ0ODMyMzY2OTU2MzIzNDcBU0x3cmtpZHcya1FKATUBAXYy)


## インストール

まず、インストールしてみます。ユーザマニュアルにある通り、いつもの方法でインストールします。

```
go get code.google.com/p/go.tools/cmd/oracle
```

ちなみに、Go言語は開発版を使用してます。

```
% go version
go version devel +06c20cdf7bc0 Mon Sep 23 18:11:25 2013 +1000 darwin/amd64
```

`$GOPATH/bin`以下にインストールされるため、`PATH`が通ってなかったら通します。

```
% export PATH=$GOPATH/bin/:$PATH
```

インストールがうまくいっていれば、`oracle -help`でずらずらっと何かでるはずです。ここでは長いので省略しています。

````
% oracle -help
Run 'oracle -help' for more information.
Go source code oracle.
Usage: oracle [<flag> ...] <args> ...

The -format flag controls the output format:
        plain   an editor-friendly format in which every line of output
                is of the form "pos: text", where pos is "-" if unknown.
        json    structured data in JSON syntax.
....
```

## コマンドについて

ユーザマニュアルにあるコマンンドの例を見てましょう。

```
% oracle -pos=src/pkg/net/http/triv.go:#1042,#1050 ­-format=json describe src/pkg/net/http/triv.go
```

つまり、こんな形式のようです。

```
% oracle -pos=<File>#<Start>,#<End> -format=<Format> <Mode> <Scope>
```

それぞれのオプションを説明します。

### `-pos`
検索対象のソースコードの位置です。
* `<File>` : ソースコードのファイルパス
* `<Start>` : 開始位置（先頭からのバイト数）
* `<End>` : 終了位置（先頭からのバイト数）

### `-format`
出力形式です。以下の形式が使用可能です。
* `json` : JSON形式。エディタなので解析するのに向いています
* `plain` : 人間が読みやすいテキスト形式
* `xml` : XML形式。

### `<Mode>`
クエリのモードで、以下のようなものがあります。
それぞれについては、後ほど説明します。

* `callees`
* `callers`
* `callgraph`
* `callstack`
* `describe`
* `freevars`
* `implements`
* `peers`
* `referrers`

### `<Scope>`
正直ここはあんまり分かっていません。
恐らくmain関数のあるファイルを書けばいいはずです。

##クエリの使い方
各モードのクエリの使い方と簡単な説明を行ないます。
`-pos`で指定しているバイト数ですが、私の場合はvimで`g CTL-g`でカーソル位置の先頭からのバイト数を取得してやっています。

### `callees`
呼び出している関数について調べるクエリです。
以下の例では、`fmt.Println`の`P`の位置を調べているので、`fmt.Println`の情報が表示されています。
#### 使用ファイル
```callee_sample.go
package main
 
import "fmt"

func main() {
    msg := "hello, oracle"
    fmt.Println(msg)
}
```

#### コマンドの例
```
% oracle -mode=callees -pos=hello.go:#73 -format=plain callee_sample.go
```

#### 結果
```
callee_sample.go:7:13: this static function call dispatches to:
/usr/local/go/src/pkg/fmt/print.go:295:6:       fmt.Println
```

### `callers`

関数を呼び出してる個所について調べるクエリです。
以下の例では、`hello`関数の宣言の`h`の位置で調べているため、`hello`を呼んでいる`main`関数の呼んでいる個所が出力されています。

#### 使用ファイル
```callers_sample.go
package main

import "fmt"

func hello() {
    msg := "hello, oracle"
    fmt.Println(msg)
}

func main() {
    hello()
}
```

#### コマンドの例
```
% oracle -mode=callers -pos=callers_sample.go:#37 -format=plain callers_sample.go
```

#### 結果
```
callers_sample.go:5:6: main.hello is called from these 1 sites:
callers_sample.go:11:7:         static function call from main.main
```

### `callgraph`

関数のコールグラフを表示するクエリです。
`main()`、`a()`、`b()`、`c()`の順で呼び出していることが分かります。

#### 使用ファイル
```callgraph_sample.go
package main

func a() {
    b()
}

func b() {
    c()
}

func c() {
}

func main() {
    a()
}
```

#### コマンドの例

```
% oracle -format=plain callgraph callgraph_sample.go 
```

#### 結果

```
-:
Below is a call graph of the entire program.
The numbered nodes form a spanning tree.
Non-numbered nodes indicate back- or cross-edges to the node whose
 number follows in parentheses.
Some nodes may appear multiple times due to context-sensitive
 treatment of some calls.

-: 0    <root>
-: 1        main.init
callgraph_sample.go:14:6: 2         main.main
callgraph_sample.go:3:6: 3              main.a
callgraph_sample.go:7:6: 4                  main.b
callgraph_sample.go:11:6: 5                     main.c
```

### `callstack`

指定した位置の関数のコールスタックを表示するクエリです。
以下の例では、`c()`の宣言の部分を対象としています。

#### 使用ファイル

```callstack_sample.go
package main

func a() {
    b()
}

func b() {
    c()
}

func c() {
}

func main() {
    a()
    c()
}
```

#### コマンドの例

```
% oracle -pos=callstack_sample.go:#58 -format=plain callstack callstack_sample.go
```

#### 結果

```
callstack_sample.go:11:7: Found a call path from root to main.c
callstack_sample.go:11:6: main.c
callstack_sample.go:8:3: static function call from main.b
callstack_sample.go:4:3: static function call from main.a
callstack_sample.go:15:3: static function call from main.main
```

### `describe`

対象の位置のサマリー情報を出します。
以下の例では、`do(foo Foo)`の`Foo`の位置を対象としています。

### 使用ファイル

```describe_sample.go
package main

type Foo interface {
    DoFoo()
}

type Bar struct {
}

func (bar *Bar) DoFoo() {
}

func do(foo Foo) {
    foo.DoFoo()
}

func main() {
    do(&Bar{})
}
```

#### コマンドの例

```
% oracle -pos=describe_sample.go:#110 -format=plain describe describe_sample.go 
```

#### 結果

```
describe_sample.go:13.13-13.15: reference to type main.Foo
describe_sample.go:3:6: defined as interface{DoFoo()}
describe_sample.go:13.13-13.15: Method set:
describe_sample.go:4:2:         method (main.Foo) DoFoo()
```

### `freevars`

自由変数を表示するクエリです。
以下の例では、`fmt.Println(i)`を対象としています。


#### 使用ファイル

```freevars_sample.go
package main

import (
    "fmt"
)

func main() {
    i := 0
    fmt.Print(i)
}
```

#### コマンドの例

```
% oracle -pos=freevars_sample.go:#57,#68 -format=plain freevars freevars_sample.go 
```

#### 結果

```
freevars_sample.go:9.3-9.13: Free identifiers:
freevars_sample.go:8:2: var i int
```

### `implements`

対象位置にあるインタフェースを実装している型の一覧を表示するクエリです。
以下の例では、`Foo`を対象にしています。

#### 使用ファイル

```implements_sample.go
package main

type Foo interface {
    Do()
}

type Bar struct {
}

func (bar *Bar) Do() {
}

type Hoge struct {
}

func (hoge *Hoge) Do() {
}

func main() {
}
```

#### コマンドの例

```
% oracle -pos=implements_sample.go:#15,#42 -format=plain implements implements_sample.go
```

#### 結果

```
implements_sample.go:3:6:       Interface main.Foo:
implements_sample.go:7:6:               *main.Bar
implements_sample.go:13:6:              *main.Hoge
```

### `peers`

チャネルの宣言や、使用位置を調べるクエリです。
以下の例では、`ch <- true`の`ch`を対象にしています。

#### 使用ファイル

```peers_sample.go
package main

func main() {
    ch := make(chan bool)
    go func() {
        ch <- true
    }()

    <-ch
}
```

#### コマンドの例

```
% oracle -pos=peers_sample.go:#67,#68 -format=plain peers peers_sample.go 
```

#### 結果

```
peers_sample.go:6:6: This channel of type chan bool may be:
peers_sample.go:4:12:   allocated here
peers_sample.go:6:6:    sent to, here
peers_sample.go:6:6:    sent to, here
peers_sample.go:9:2:    received from, here
```

### `referrers`

対象位置にある変数などの参照位置を調べるクエリです。
以下の例では、`f(n int)`の`n`を対象としています。

#### 使用ファイル
```referrers_sample.go
package main

import (
    "fmt"
)

func f(n int) {
    fmt.Println(n)
}

func main() {
    f(100)
}
```

#### コマンドの例

```
% oracle -pos=referrers_sample.go:#40 -format=plain referrers referrers_sample.go 
```

#### 結果

```
referrers_sample.go:7:8: defined here as var n int
referrers_sample.go:8:14: referenced here
```
