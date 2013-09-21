#Go研 Vol.8

* 発表者：[@tenntenn](https://twitter.com/tenntenn)

## reflectパッケージと良く使うパターン

Go言語でリフレクションを行なうには、`reflect`パッケージを使用する。
`reflect`パッケージは標準パッケージでも多く使われていて、`encoding`パッケージがよい例である。

* `reflect`パッケージ: http://golang.org/pkg/reflect
* `encoding`パッケージ: http://golang.org/pkg/encoding

JSONをエンコード／デコードする`encoding/json`パッケージでは、

````
type Hoge struct {
    Field1 string `json:"field1"`
    Field2 int    `json:"field2"`
    Field3 string `json:"-"`
}
````

のような構造体を作っておき、リフレクションすることでエンコードしてJSONにしたり、JSONをデコードして構造体に落とし込んでいる。また、デコード結果を変数に格納する時も、汎用的にするために`interface{}`型で返すのではなく、ポインタを`interface{}`型で引数として受け取って値を設定している。

`````
var hoge Hoge
// func Unmarshal([]byte, interface{}) error
err := json.Unmarshal(jsonStr, &hoge)
`````

リフレクションのパターンには以下のようなものがあり、以降でそれぞれ説明する。

* `struct`のリフレクションパターン
* `channel`のリフレクションパターン
* `func`のリフレクションパターン

まず`reflect`パッケージの基礎となる`Value`型と`Type`型について説明し、その後、各パターンについて説明する。

## Value型とType型

reflectパッケージを使いこなすには以下の2つの型を知っておく必要がある。

* Value型: http://golang.org/pkg/reflect/#Value
* Type型: http://golang.org/pkg/reflect/#Type

### Value型

### Type型
