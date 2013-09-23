#Go研 Vol.8

* 発表者：[@tenntenn](https://twitter.com/tenntenn)

## reflectパッケージと良く使うパターン

Go言語でリフレクションを行なうには、`reflect`パッケージを使用する。
`reflect`パッケージは標準パッケージでも多く使われていて、`encoding`パッケージがよい例である。

* `reflect`パッケージ: http://golang.org/pkg/reflect
* `encoding`パッケージ: http://golang.org/pkg/encoding

JSONをエンコード／デコードする`encoding/json`パッケージでは、

```
type Hoge struct {
    Field1 string `json:"field1"`
    Field2 int    `json:"field2"`
    Field3 string `json:"-"`
}
```

のような構造体を作っておき、リフレクションすることでエンコードしてJSONにしたり、JSONをデコードして構造体に落とし込んでいる。また、デコード結果を変数に格納する時も、汎用的にするために`interface{}`型で返すのではなく、ポインタを`interface{}`型で引数として受け取って値を設定している。

```
var hoge Hoge
// func Unmarshal([]byte, interface{}) error
err := json.Unmarshal(jsonStr, &hoge)
```

リフレクションのパターンには以下のようなものがあり、以降でそれぞれ説明する。

* `struct`のリフレクションパターン
* `channel`のリフレクションパターン
* `func`のリフレクションパターン

まず`reflect`パッケージの基礎となる`Value`型と`Type`型について説明し、その後、各パターンについて説明する。

## `Value`型と`Type`型

`reflect`パッケージを使いこなすには以下の2つの型を知っておく必要がある。

* `Value`型: http://golang.org/pkg/reflect/#Value
* `Type`型: http://golang.org/pkg/reflect/#Type

### `Value`型

`Value`型は以下のように定義(コメントは省略)されている。

https://code.google.com/p/go/source/browse/src/pkg/reflect/value.go#61

```src/pkg/reflect/value.go#61
type Value struct {
    typ *rtype
    val unsafe.Pointer
    flag
}
```

`rtype`型は後述する`Type`型の実体である。
つまり、`typ`は型情報である。
`val`はこの`Value`オブジェクトの表す値の実体へのポインタである。
`flag`は以下のようにメタ情報を保持している。

* 下位4ビット: フラグ用
    * flagRO: 読み込み専用か
    * flagIndir: ポインタかどうか
        * typ.size > ptrSizeならフラグが立つらしい
    * flagAddr: `v.CanAddr`が`true`かどうか
    * flagMethod: メソッドかどうか
* 次の5ビット: 値の種類
    * `Kind`メソッドで取得できる値
* 残りの23ビット: メソッドの為の領域?
    * この部分についてはよくわからない

なお、`flag`型は以下のように定義されている。
ビット操作の為のマスクやシフト幅などが`iota`を使ってうまく行なわれている。

https://code.google.com/p/go/source/browse/src/pkg/reflect/value.go#96

```src/pkg/reflect/value.go#96
type flag uintptr
const (
        flagRO flag = 1 << iota 
        flagIndir              
        flagAddr              
        flagMethod
        // 以下はフラグ操作の為に使用
        flagKindShift        = iota // 4 (4ビット分シフトが必要)
        flagKindWidth        = 5 // there are 27 kinds
        flagKindMask    flag = 1<<flagKindWidth - 1
        flagMethodShift      = flagKindShift + flagKindWidth
)
```

### `Type`型

