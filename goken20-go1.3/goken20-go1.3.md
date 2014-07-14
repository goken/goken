# Go研 Vol.20
## 発表者
* [@tenntenn](https://twitter.com/tenntenn)

## Go1.3 の新機能研究

### リリースノート
* [本家](http://tip.golang.org/doc/go1.3)
* [日本語訳](https://docs.google.com/document/d/1drhNIsy44lpnmIHsSCWbP54ZzBHYm82QfaTLEA7pbVo/preview?hl=ja&forcehl=1)

### 気になる変更
分からないところは，分かる人に聞こう！
引用部分は，[@rui314](https://twitter.com/rui314)さんの日本語訳から引用した．

#### スタック
「分割された」モデルから連続したスタックになった？
>Go 1.3ではgoroutineのスタックの実装が変わり、以前の「分割された」モデルから連続したスタックになりました。Goroutineがより多くのスタック領域を必要とする場合、連続したより大きなメモリ領域にスタック全体がコピーされます。このコピーのオーバヘッドが全体に与える影響は少ない一方、以前の実装にあった「ホットスポット」問題（分割されたスタック領域のちょうど境界線上を何度もまたいでしまう場合の問題）が解決されています。実装の設計およびパフォーマンスについては[デザインドキュメント](https://docs.google.com/document/d/1wAaf1rYoM4S4gtnPh0zOlGzWtrZFQ5suE8qr2sD8uWQ/pub)を参照してください。

#### ガベージコレクタの変更
GCがスタック上の値でも正確になった？
> 以前より、ヒープ上の値についてはガベージコレクタの動作は正確でした。Go 1.3からはスタック上の値についてもそれと同じく正確になります。つまりGoが、整数のような非ポインタ値を間違ってポインタとして解釈して、差されているオブジェクトが回収されないということがなくなります。
>
>Go 1.3から、ランタイムはポインタ型の値はポインタであり、非ポインタ型の値はポインタではないものとして扱います。この想定はスタック拡張ルーチンやガベージコレクタにとって基本的なものです。unsafeパッケージを使ってポインタではない整数をポインタ型に代入しているプログラムは不正で、ランタイムがそれを検出するとクラッシュすることがあります。unsafeパッケージを使ってポインタ値を整数型の値に代入しているプログラムも不正ですが、こちらはより大きな問題です。ランタイムはポインタの値を知ることができないため、スタック拡張ルーチンまたはガベージコレクタがメモリを回収してしまい、不正な領域を指すポインタになってしまうことがあります。
>
>アップデート： unsafe.Pointerを使って整数型の値をポインタに変換しているコードは不正で、アップデートする必要があります。そのようなコードはgo vetで発見できます。

#### マップのイテレーション
元々仕様でもマップの順序は毎回異なる可能性があることは明記されていたが，Go1.1と1.2では順序に依存した処理が書けていた．
しかし，1.3からそのような実装はちゃんと動かなくなる．
[サンプル](http://play.golang.org/p/eROVtw7owL)

#### godocの変更
`-analysis`が追加された．静的解析が行うことができ，実装しているインタフェースなどが分かるようになった．
`go oracle`で使われていた技術が利用されているのかな？

>godocに-analysisフラグをつけて実行すると、洗練された静的解析を行うようになりました。解析の結果はソースビューとパッケージドキュメンテーションビューのどちらでも見ることができます。結果にはパッケージごとのコールグラフ、定義や参照の関係、型およびそのメソッド、インターフェイスとその実装、チャネルに対する送信・受信操作、関数とそれの呼び出し元、呼び出している場所、呼び出されている場所が含まれます。

#### パフォーマンス
>Goバイナリのパフォーマンスが全体的に向上しています。これはランタイム、ガベージコレクションおよびライブラリの改善によるものです。とくに顕著なものは次のとおりです。

>* ランタイムがdeferを効率的に扱うようになりました。これによりdeferを呼び出しているgoroutineのメモリ消費量が2キロバイト少なくなりました。
>* 並列スイープアルゴリズム、よりよい並列化、大きなページの採用により、ガベージコレクタが高速化しました。これにより停止時間が50〜70%減少しました。
>* 競合検出機能（race detector、ガイドを参照）が40%速くなりました。
>* 正規表現パッケージregexpが、特定の小さな正規表現について大幅に高速化しました。これは新たな1パス実行エンジンが実装されたためです。エンジンの選択は自動で行われるため、詳細はユーザには見えません。

>また、スタックダンプにgoroutineがどれくらいの時間ブロックしていたのかが表示されるようになりました。デッドロックやパフォーマンス問題があった場合にこの情報が役立ちます。

#### sync.Poolの追加
[Jxckさんに聞こう](http://jxck.hatenablog.com/entry/sync.Pool)．
syncパッケージについては，mattnさんも[書いている](http://mattn.kaoriya.net/software/lang/go/20140625223125.htm)．

>sync.Poolは特定の型に対するキャッシュの実装を提供しています。キャッシュの内容はシステムによって自動的にガベージコレクトされます。

#### testing.B.RunParallel
[B.RunParallel](http://golang.org/pkg/testing/#B.RunParallel)が追加された．

```
package main

import (
    "bytes"
    "testing"
    "text/template"
)

func main() {
    // Parallel benchmark for text/template.Template.Execute on a single object.
    testing.Benchmark(func(b *testing.B) {
        templ := template.Must(template.New("test").Parse("Hello, {{.}}!"))
        // RunParallel will create GOMAXPROCS goroutines
        // and distribute work among them.
        b.RunParallel(func(pb *testing.PB) {
            // Each goroutine has its own bytes.Buffer.
            var buf bytes.Buffer
            for pb.Next() {
                // The loop body is executed b.N times total across all goroutines.
                buf.Reset()
                templ.Execute(&buf, "World")
            }
        })
    })
}
```

>testingパッケージのベンチマークヘルパーBに、RunParallelメソッドが追加されました。これにより複数のCPU上で実行されるベンチマークを書くことが簡単になりました。
