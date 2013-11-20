# ソースコードの静的解析ツール Go Oracleを使ってみた

## Go Oracleとは？

Go言語のリポジトリを漁ってると、go toolsの中にoracleというディレクトリを発見しました。何だろうなぁと思ってると、ちょうど[golang nuts](https://groups.google.com/forum/#!topic/golang-nuts/CwdIJZs6Tfc)に話題がでたので、触ってみました。

名前がややこしくて、検索しづらいのですが、Go Oracleはソースコードの静的解析ツールのようです。

* https://code.google.com/p/go/source/browse?repo=tools#hg%2Foracle

使い方等は、以下のドキュメント（英語）を見れば良さそうです。
何はともあれ、とりあえず使ってみましょう。Emacsからでも使えるようですが、ここではコマンドから使ってみましょう。

* [ユーザマニュアル](https://docs.google.com/viewer?a=v&pid=forums&srcid=MDg3NjYzNDU1NTk0NjU2OTUyMDMBMDY5OTQ0ODMyMzY2OTU2MzIzNDcBU0x3cmtpZHcya1FKATQBAXYy)
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
% oracle ­-mode=describe -pos=src/pkg/net/http/triv.go:#1042,#1050 ­-format=json src/pkg/net/http/triv.go
```

つまり、こんな形式のようです。

```
% oracle -mode=<Mode> -pos=<File>#<Start>,#<End> -format=<Format> <Scope>
```

それぞれのオプションを説明します。

### `-mode`
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

### `-pos`
検索対象のソースコードの位置です。
* `<File>` : ソースコードのファイルパス
* `<Start>` : 開始位置（先頭からのバイト数）
* `<End>` : 終了位置（先頭からのバイト数）

### `-format`
出力形式です。以下の形式が使用可能です。
* `json` : JSON形式。エディタなので解析するのに向いています
* `plain` : 人間が読みやすいテキスト形式

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
### `callstack`
### `describe`
### `freevars`
### `implements`
### `peers`
### `referrers`
